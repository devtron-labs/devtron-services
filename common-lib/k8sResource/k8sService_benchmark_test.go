/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package k8sResource

import (
	"context"
	"flag"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/kubelink/config"
	"github.com/devtron-labs/kubelink/internals/logger"
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"io"
	k8sApiV1 "k8s.io/api/core/v1"
	k8sApiMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	k8sClinetV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

var sugaredLogger *zap.SugaredLogger

const multiplyFactor = 10

func init() {
	sugaredLogger = logger.NewSugaredLogger()
}

func dependencyInit() (k8sUtil *k8s.K8sServiceImpl, restConfig *rest.Config, client *k8sClinetV1.CoreV1Client) {
	// Mock K8sService implementation
	runtimeCfg, err := k8s.GetRuntimeConfig()
	if err != nil {
		log.Errorf("Failed to get runtime config: %v", err)
		return
	}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	flag.CommandLine.Usage = flag.Usage
	runtimeCfg.LocalDevMode = true
	k8sUtil, err = k8s.NewK8sUtil(sugaredLogger, runtimeCfg)
	if err != nil {
		log.Errorf("Failed to create K8sUtil: %v", err)
		return
	}
	restConfig, err = k8sUtil.GetK8sInClusterRestConfig()
	if err != nil {
		log.Errorf("Failed to get rest config: %v", err)
		return
	}
	client, err = k8sUtil.GetCoreV1ClientByRestConfig(restConfig)
	if err != nil {
		log.Errorf("Failed to get core v1 client: %v", err)
		return
	}
	return
}

func BenchmarkGetChildObjectsV2_ResourceByLimit_1_1(b *testing.B) {
	var pageLimit int64 = 100 * multiplyFactor
	resourceCount := 100 * multiplyFactor
	benchmarkGetChildObjectsV2(b, resourceCount, pageLimit)
}

func BenchmarkGetChildObjectsV2_ResourceByLimit_3_4(b *testing.B) {
	var pageLimit int64 = 75 * multiplyFactor
	resourceCount := 100 * multiplyFactor
	benchmarkGetChildObjectsV2(b, resourceCount, pageLimit)
}

func BenchmarkGetChildObjectsV2_ResourceByLimit_1_2(b *testing.B) {
	var pageLimit int64 = 50 * multiplyFactor
	resourceCount := 100 * multiplyFactor
	benchmarkGetChildObjectsV2(b, resourceCount, pageLimit)
}

func BenchmarkGetChildObjectsV2_ResourceByLimit_1_4(b *testing.B) {
	var pageLimit int64 = 25 * multiplyFactor
	resourceCount := 100 * multiplyFactor
	benchmarkGetChildObjectsV2(b, resourceCount, pageLimit)
}

func BenchmarkGetChildObjectsV2_ResourceByLimit_1_5(b *testing.B) {
	var pageLimit int64 = 20 * multiplyFactor
	resourceCount := 100 * multiplyFactor
	benchmarkGetChildObjectsV2(b, resourceCount, pageLimit)
}

func benchmarkGetChildObjectsV2(b *testing.B, resourceCount int, pageLimit int64) {
	b.Helper()
	b.SetParallelism(1)
	testName := fmt.Sprintf("test-listing-page-limit-%d", pageLimit)
	k8sUtil, restConfig, client := dependencyInit()
	helmReleaseConfig, err := config.GetHelmReleaseConfig()
	if err != nil {
		b.Fatalf("Failed to get helm release config: %v", err)
		return
	}
	helmReleaseConfig.ParentChildGvkMapping = getParentChildGvkMapForBenchmarkTest()
	helmReleaseConfig.ChildObjectListingPageSize = pageLimit
	k8sService, err := NewK8sServiceImpl(sugaredLogger, helmReleaseConfig)
	if err != nil {
		b.Fatalf("Failed to create K8sService: %v", err)
		return
	}
	highlightedLog("\n\n\n==========================================================================")
	highlightedLog(fmt.Sprintf("=== Running benchmark for %s with resource count: %d\n", testName, resourceCount))
	ctx, cancel := context.WithCancel(context.Background())
	// setup logic
	releaseName := fmt.Sprintf("test-deployment-limit-%d", pageLimit)
	err = createDeploymentWithCMs(ctx, k8sUtil, client, releaseName, resourceCount)
	if err != nil {
		b.Fatalf("Failed to setup test case: %v", err)
		return
	}
	b.Cleanup(func() {
		// cleanup logic
		cancel()
		_ = tearDownDeploymentWithCMs(releaseName)
		highlightedLog("==========================================================================\n\n\n")
	})
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		startTime := time.Now()
		_, err = k8sService.GetChildObjectsV2(restConfig, "ent-8-env-1", schema.GroupVersionKind{
			Kind:    "Deployment",
			Group:   "apps",
			Version: "v1",
		}, fmt.Sprintf("%s-memcached", releaseName))
		highlightedLog("==========================================================================")
		highlightedLog(fmt.Sprintf("=== Time taken to get child objects: %v", time.Since(startTime).Seconds()))
		highlightedLog("==========================================================================")
		if err != nil {
			b.Fatalf("Failed to get child objects: %v", err)
			return
		}
	}
}

func createDeploymentWithCMs(ctx context.Context, k8sUtil k8s.K8sService,
	client *k8sClinetV1.CoreV1Client, releaseName string, cmCount int) error {
	highlightedLog("=== Creating helm release...")
	// Create helm release
	err := executeCommand(ctx, "helm", "install", releaseName, "oci://registry-1.docker.io/bitnamicharts/memcached", "--namespace", "ent-8-env-1", "--create-namespace")
	if err != nil {
		return err
	}
	// Get the UID of the deployment
	ucidBytes, err := executeCommandWithOutput(ctx, "kubectl", "get", "deployment", fmt.Sprintf("%s-memcached", releaseName), "--namespace", "ent-8-env-1", "-o", "jsonpath={.metadata.uid}")
	if err != nil {
		return err
	}
	deploymentUid := types.UID(ucidBytes)
	// validate the UID
	if len(deploymentUid) == 0 || strings.ContainsAny(string(deploymentUid), " ") {
		return fmt.Errorf("no deployment uid found")
	}
	// Create configmap with owner reference
	for i := 0; i < cmCount; i++ {
		newConfigMap := &k8sApiV1.ConfigMap{ObjectMeta: k8sApiMetaV1.ObjectMeta{Name: fmt.Sprintf("benchmark-test-%d-configmap-%s", i, rand.SafeEncodeString(rand.String(10)))}}
		newConfigMap.Data = map[string]string{
			"createdBy":   "integration-testing",
			"createdAt":   time.Now().Format(time.RFC3339),
			"fileName":    "test-file",
			"filePath":    "/tmp/test-file",
			"fileContent": rand.SafeEncodeString(rand.String(90)),
		}
		newConfigMap.OwnerReferences = []k8sApiMetaV1.OwnerReference{
			{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       fmt.Sprintf("%s-memcached", releaseName),
				UID:        deploymentUid,
				Controller: ptr.To(true),
			},
		}
		_, err := k8sUtil.CreateConfigMap("ent-8-env-1", newConfigMap, client)
		if err != nil {
			return err
		}
	}
	highlightedLog("=== Created configmaps with owner reference ✔")
	return nil
}

func tearDownDeploymentWithCMs(releaseName string) error {
	highlightedLog("=== Deleting helm release...")
	// Delete helm release
	_ = executeCommand(context.Background(), "helm", "uninstall", releaseName, "--namespace", "ent-8-env-1")
	return nil
}

// executeCommandWithOutput executes a command with the given arguments and returns the output as a byte slice.
func executeCommandWithOutput(ctx context.Context, name string, arg ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, arg...)
	return cmd.Output()
}

// getParentChildGvkMapForBenchmarkTest
// DAG Graph Representation of Parent-Child Objects:
//
//	ReplicaSet (v1, apps)
//	└─── Pods (v1, namespace)
//
//	Deployment (v1, apps)
//	├─── ReplicaSets (v1, apps, namespace)
//	│    └─── Pods (v1, namespace)
//	└─── ConfigMaps (v1, namespace)
func getParentChildGvkMapForBenchmarkTest() string {
	return `[{"childObjects":[{"Group":"","Resource":"pods","Scope":"namespace","Version":"v1"}],"group":"apps","kind":"ReplicaSet","version":"v1"},{"childObjects":[{"Group":"apps","Resource":"replicasets","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"configmaps","Scope":"namespace","Version":"v1"}],"group":"apps","kind":"Deployment","version":"v1"}]`
}
