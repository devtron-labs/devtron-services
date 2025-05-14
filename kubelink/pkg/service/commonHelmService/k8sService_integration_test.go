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

package commonHelmService

import (
	"context"
	"fmt"
	"github.com/devtron-labs/kubelink/config"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func getIntegrationTestCases() []integrationTestCase {
	return []integrationTestCase{
		getFluxCDDeploymentTestCase(),
		fluentBitDaemonsetTestCase(),
		getRolloutDeploymentTestCase(),
		getDeploymentTestCase(),
		getCronJobTestCase(),
		getMongodbCommunityTestCase(),
		getKubePrometheusDaemonsetTestcase(),
		getKubePrometheusDeploymentTestcase(),
		getKubePrometheusAlertManagerTestcase(),
		getKubePrometheusTestcase(),
	}
}

func TestGetChildObjectsIntegration(t *testing.T) {
	// Mock K8sService implementation
	helmReleaseConfig, err := config.GetHelmReleaseConfig()
	if err != nil {
		t.Fatalf("Failed to get helm release config: %v", err)
		return
	}
	helmReleaseConfig.ParentChildGvkMapping = getParentChildGvkMapForIntegrationTest()
	k8sService, err := NewK8sServiceImpl(sugaredLogger, helmReleaseConfig)
	if err != nil {
		t.Fatalf("Failed to create K8sService: %v", err)
		return
	}
	_, restConfig, _ := dependencyInit()
	for _, tt := range getIntegrationTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			// Create context
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(func() {
				// cleanup logic
				cancel()
				_ = tt.cleanup()
			})
			// setup logic
			if err = tt.setup(ctx); err != nil {
				t.Fatalf("Failed to setup test: %v", err)
				return
			}
			testGetChildObjects(t, tt, k8sService, restConfig)
			return
		})
	}
	return
}

func testGetChildObjects(t *testing.T, tt integrationTestCase, k8sService K8sService, restConfig *rest.Config) {
	t.Run(fmt.Sprintf("GVK-%q", tt.parentGvk.String()), func(t *testing.T) {
		parentApiVersion, _ := tt.parentGvk.ToAPIVersionAndKind()
		resultV1, errV1 := k8sService.GetChildObjectsV1(restConfig, tt.namespace, tt.parentGvk, tt.parentName, parentApiVersion)
		resultV2, errV2 := k8sService.GetChildObjectsV2(restConfig, tt.namespace, tt.parentGvk, tt.parentName)
		assert.Truef(t, compareValidResourceTree(resultV1, resultV2), "Resource trees do not match for %s", tt.name)
		if errV1 != nil {
			assert.EqualError(t, errV1, errV2.Error(), "Errors do not match for %s", tt.name)
		} else {
			assert.NoError(t, errV2, "Errors do not match for %s", tt.name)
		}
		fmt.Println("=== Child Object Count:", len(resultV1))
		for _, manifest := range resultV1 {
			gvk := manifest.GroupVersionKind()
			if k8sService.CanHaveChild(gvk) {
				tt.parentGvk = gvk
				tt.parentName = manifest.GetName()
				tt.namespace = manifest.GetNamespace()
				// Recursively call the test function for child objects
				testGetChildObjects(t, tt, k8sService, restConfig)
			}
		}
	})
}

// compareValidResourceTree compares two slices of unstructured.Unstructured objects
func compareValidResourceTree(v1Result []*unstructured.Unstructured, v2Result []*unstructured.Unstructured) bool {
	if len(v1Result) != len(v2Result) {
		return false
	}
	// group by gvk and compare for each gvk
	v1Map := make(map[string][]*unstructured.Unstructured)
	v2Map := make(map[string][]*unstructured.Unstructured)
	for _, v1Obj := range v1Result {
		gvk := v1Obj.GroupVersionKind()
		key := gvk.String()
		v1Map[key] = append(v1Map[key], v1Obj)
	}
	// group by gvk and compare for each gvk
	for _, v2Obj := range v2Result {
		gvk := v2Obj.GroupVersionKind()
		key := gvk.String()
		v2Map[key] = append(v2Map[key], v2Obj)
	}
	// compare length of each gvk
	if len(v1Map) != len(v2Map) {
		return false
	}
	// compare each gvk
	for key, v1Objs := range v1Map {
		v2Objs, ok := v2Map[key]
		if !ok {
			return false
		}
		if len(v1Objs) != len(v2Objs) {
			return false
		}
		for i, v1Obj := range v1Objs {
			v2Obj := v2Objs[i]
			if v1Obj.GetKind() != v2Obj.GetKind() {
				return false
			}
			if v1Obj.GetAPIVersion() != v2Obj.GetAPIVersion() {
				return false
			}
			if v1Obj.GetNamespace() != v2Obj.GetNamespace() {
				return false
			}
			if v1Obj.GetName() != v2Obj.GetName() {
				return false
			}
			if v1Obj.GetResourceVersion() != v2Obj.GetResourceVersion() {
				return false
			}
			if v1Obj.GetCreationTimestamp() != v2Obj.GetCreationTimestamp() {
				return false
			}
		}
	}
	return true
}

const (
	Reset  = "\033[0m"
	Yellow = "\033[33m"
)

// highlightedLog logs the setup and cleanup messages with yellow color
func highlightedLog(msg string) {
	fmt.Println(Yellow + msg + Reset)
}

// executeCommand executes a command and returns the error if any
func executeCommand(ctx context.Context, name string, arg ...string) error {
	cmd := exec.CommandContext(ctx, name, arg...)
	// write stdout and stderr to os.Stdout and os.Stderr with color code
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type integrationTestCase struct {
	name       string
	namespace  string
	parentGvk  schema.GroupVersionKind
	parentName string
	setup      func(context.Context) error
	cleanup    func() error
}

func getCronJobTestCase() integrationTestCase {
	return integrationTestCase{
		name:      "helm-release-cron-job",
		namespace: "ent-8-env-1",
		parentGvk: schema.GroupVersionKind{
			Kind:    "Job",
			Group:   "batch",
			Version: "v1",
		},
		parentName: "test-cron-job-v1-1",
		setup: func(ctx context.Context) error {
			highlightedLog("=== Test helm release daemonset resource tree!")
			highlightedLog("✔ Pre-requisite to run helm-release-daemonset: helm should be installed locally.")
			highlightedLog("► helm chart should be installed in the cluster.")
			highlightedLog("Creating helm release...")
			// Create helm release
			err := executeCommand(ctx, "helm", "install", "test-cron-job-v1", "oci://registry-1.docker.io/ashexp/cron-job", "--version", "1.0.173-DEPLOY-test-app-v1.0.0", "--namespace", "ent-8-env-1", "--create-namespace")
			if err != nil {
				return err
			}
			return nil
		},
		cleanup: func() error {
			highlightedLog("Deleting helm release...")
			// Delete helm release
			_ = executeCommand(context.Background(), "helm", "uninstall", "test-cron-job-v1", "--namespace", "ent-8-env-1")
			return nil
		},
	}
}

func getDeploymentTestCase() integrationTestCase {
	return integrationTestCase{
		name:      "helm-release-deployment",
		namespace: "ent-8-env-1",
		parentGvk: schema.GroupVersionKind{
			Kind:    "Deployment",
			Group:   "apps",
			Version: "v1",
		},
		parentName: "test-deployment-v1",
		setup: func(ctx context.Context) error {
			highlightedLog("=== Test helm release daemonset resource tree!")
			highlightedLog("✔ Pre-requisite to run helm-release-daemonset: helm should be installed locally.")
			highlightedLog("► helm chart should be installed in the cluster.")
			highlightedLog("Creating helm release...")
			// Create helm release
			err := executeCommand(ctx, "helm", "install", "test-deployment-v1", "oci://registry-1.docker.io/ashexp/deployment", "--version", "1.0.174-DEPLOY-test-app-v1.0.0", "--namespace", "ent-8-env-1", "--create-namespace")
			if err != nil {
				return err
			}
			return nil
		},
		cleanup: func() error {
			highlightedLog("Deleting helm release...")
			// Delete helm release
			_ = executeCommand(context.Background(), "helm", "uninstall", "test-deployment-v1", "--namespace", "ent-8-env-1")
			return nil
		},
	}
}

func fluentBitDaemonsetTestCase() integrationTestCase {
	return integrationTestCase{
		name:      "helm-release-daemonset",
		namespace: "ent-8-env-1",
		parentGvk: schema.GroupVersionKind{
			Kind:    "DaemonSet",
			Group:   "apps",
			Version: "v1",
		},
		parentName: "test-daemonset-helm-chart-fluent-bit",
		setup: func(ctx context.Context) error {
			highlightedLog("=== Test helm release daemonset resource tree!")
			highlightedLog("✔ Pre-requisite to run helm-release-daemonset: helm should be installed locally.")
			highlightedLog("► helm chart should be installed in the cluster.")
			highlightedLog("Creating helm release...")
			// Add helm repo
			err := executeCommand(ctx, "helm", "repo", "add", "fluent", "https://fluent.github.io/helm-charts")
			if err != nil {
				return err
			}
			_ = executeCommand(ctx, "helm", "repo", "update")
			// Create helm release
			err = executeCommand(ctx, "helm", "install", "test-daemonset-helm-chart", "fluent/fluent-bit", "--namespace", "ent-8-env-1", "--create-namespace")
			if err != nil {
				return err
			}
			return nil
		},
		cleanup: func() error {
			highlightedLog("Deleting helm release...")
			// Delete helm release
			_ = executeCommand(context.Background(), "helm", "uninstall", "test-daemonset-helm-chart", "--namespace", "ent-8-env-1")
			// Delete helm repo
			_ = executeCommand(context.Background(), "helm", "repo", "remove", "fluent")
			return nil
		},
	}
}

func getFluxCDDeploymentTestCase() integrationTestCase {
	return integrationTestCase{
		name:      "flux-cd-application",
		namespace: "ent-8-env-1",
		parentGvk: schema.GroupVersionKind{
			Kind:    "Deployment",
			Group:   "apps",
			Version: "v1",
		},
		parentName: "react-application",
		setup: func(ctx context.Context) error {
			highlightedLog("✔ Pre-requisite to run flux-cd-application: flux cli should be installed locally.")
			highlightedLog("► flux cd should be installed in the cluster.")
			highlightedLog("◎ installation command: brew install fluxcd/tap/flux")
			highlightedLog("Creating flux cd kustomization...")
			// Create flux source
			err := executeCommand(ctx, "flux", "create", "source", "git", "react", "--url=https://github.com/AnaisUrlichs/react-article-display", "--branch=main")
			if err != nil {
				return err
			}
			// Create flux helm release
			err = executeCommand(ctx, "flux", "create", "kustomization", "react-app", "--target-namespace=ent-8-env-1", "--source=react", "--path=./deploy/manifests", "--prune=true", "--interval=5m")
			if err != nil {
				return err
			}
			return nil
		},
		cleanup: func() error {
			highlightedLog("Deleting flux cd kustomization...")
			// Delete flux source
			_ = executeCommand(context.Background(), "flux", "delete", "source", "git", "react", "-s")
			// Delete flux helm release
			_ = executeCommand(context.Background(), "flux", "delete", "kustomization", "react-app", "-s")
			return nil
		},
	}
}

func getKubePrometheusDaemonsetTestcase() integrationTestCase {
	return integrationTestCase{
		name:      "helm-release-kube-prometheus/daemonset",
		namespace: "utils",
		parentGvk: schema.GroupVersionKind{
			Kind:    "DaemonSet",
			Group:   "apps",
			Version: "v1",
		},
		parentName: "shared-monitoring-stack-prometheus-node-exporter",
		setup: func(ctx context.Context) error {
			highlightedLog("=== Test helm release kube-prometheus resource tree!")
			highlightedLog("✔ Pre-requisite to run helm-release-kube-prometheus: kube-prometheus should be installed in the cluster.")
			highlightedLog("► kube-prometheus should have a daemonset named 'shared-monitoring-stack-prometheus-node-exporter'.")
			return nil
		},
		cleanup: func() error {
			return nil
		},
	}
}

func getKubePrometheusDeploymentTestcase() integrationTestCase {
	return integrationTestCase{
		name:      "helm-release-kube-prometheus/deployment",
		namespace: "utils",
		parentGvk: schema.GroupVersionKind{
			Kind:    "Deployment",
			Group:   "apps",
			Version: "v1",
		},
		parentName: "shared-monitoring-stack-grafana",
		setup: func(ctx context.Context) error {
			highlightedLog("=== Test helm release kube-prometheus resource tree!")
			highlightedLog("✔ Pre-requisite to run helm-release-kube-prometheus: kube-prometheus should be installed in the cluster.")
			highlightedLog("► kube-prometheus should have a deployment named 'shared-monitoring-stack-grafana'.")
			return nil
		},
		cleanup: func() error {
			return nil
		},
	}
}

func getKubePrometheusAlertManagerTestcase() integrationTestCase {
	return integrationTestCase{
		name:      "helm-release-kube-prometheus/alertmanager",
		namespace: "utils",
		parentGvk: schema.GroupVersionKind{
			Kind:    "Alertmanager",
			Group:   "monitoring.coreos.com",
			Version: "v1",
		},
		parentName: "shared-monitoring-stack-ku-alertmanager",
		setup: func(ctx context.Context) error {
			highlightedLog("=== Test helm release kube-prometheus resource tree!")
			highlightedLog("✔ Pre-requisite to run helm-release-kube-prometheus: kube-prometheus should be installed in the cluster.")
			highlightedLog("► kube-prometheus should have an alertmanager named 'shared-monitoring-stack-ku-alertmanager'.")
			return nil
		},
		cleanup: func() error {
			return nil
		},
	}
}

func getKubePrometheusTestcase() integrationTestCase {
	return integrationTestCase{
		name:      "helm-release-kube-prometheus/prometheus",
		namespace: "utils",
		parentGvk: schema.GroupVersionKind{
			Kind:    "Prometheus",
			Group:   "monitoring.coreos.com",
			Version: "v1",
		},
		parentName: "shared-monitoring-stack-ku-prometheus",
		setup: func(ctx context.Context) error {
			highlightedLog("=== Test helm release kube-prometheus resource tree!")
			highlightedLog("✔ Pre-requisite to run helm-release-kube-prometheus: kube-prometheus should be installed in the cluster.")
			highlightedLog("► kube-prometheus should have a prometheus named 'shared-monitoring-stack-ku-prometheus'.")
			return nil
		},
		cleanup: func() error {
			return nil
		},
	}
}

func getMongodbCommunityTestCase() integrationTestCase {
	return integrationTestCase{
		name:      "helm-release-mongodbcommunity",
		namespace: "mongodb",
		parentGvk: schema.GroupVersionKind{
			Kind:    "MongoDBCommunity",
			Group:   "mongodbcommunity.mongodb.com",
			Version: "v1",
		},
		parentName: "mongodb",
		setup: func(ctx context.Context) error {
			highlightedLog("=== Test helm release mongodbcommunity resource tree!")
			highlightedLog("✔ Pre-requisite to run helm-release-mongodbcommunityt: helm should be installed locally.")
			highlightedLog("► helm chart should be installed in the cluster.")
			highlightedLog("Creating helm release...")
			// Add helm repo
			err := executeCommand(ctx, "helm", "repo", "add", "mongodb", "https://mongodb.github.io/helm-charts")
			if err != nil {
				return err
			}
			_ = executeCommand(ctx, "helm", "repo", "update")
			// Install mongodb community operator
			err = executeCommand(ctx, "helm", "install", "community-operator", "mongodb/community-operator", "--set", "community-operator-crds.enabled=false", "--namespace", "mongodb", "--create-namespace")
			if err != nil {
				return err
			}
			// Install mongodb database
			err = executeCommand(ctx, "helm", "install", "mongodb-database", "oci://registry-1.docker.io/ashexp/mongodb-database", "--version", "v1.0.0", "--namespace", "mongodb", "--create-namespace")
			if err != nil {
				return err
			}
			return nil
		},
		cleanup: func() error {
			highlightedLog("Deleting helm release...")
			// Delete mongodb secret
			_ = executeCommand(context.Background(), "kubectl", "delete", "secret", "admin-user", "--namespace", "mongodb")
			// Delete helm release
			_ = executeCommand(context.Background(), "helm", "uninstall", "mongodb-database", "--namespace", "mongodb")
			_ = executeCommand(context.Background(), "helm", "uninstall", "community-operator", "--namespace", "mongodb")
			// Delete helm repo
			_ = executeCommand(context.Background(), "helm", "repo", "remove", "mongodb")
			return nil
		},
	}
}

func getRolloutDeploymentTestCase() integrationTestCase {
	return integrationTestCase{
		name:      "helm-release-rollout-deployment",
		namespace: "ent-8-env-1",
		parentGvk: schema.GroupVersionKind{
			Kind:    "Rollout",
			Group:   "argoproj.io",
			Version: "v1alpha1",
		},
		parentName: "test-rollout-deployment-v1",
		setup: func(ctx context.Context) error {
			highlightedLog("=== Test helm release daemonset resource tree!")
			highlightedLog("✔ Pre-requisite to run helm-release-daemonset: helm should be installed locally.")
			highlightedLog("► helm chart should be installed in the cluster.")
			highlightedLog("Creating helm release...")
			// Create helm release
			err := executeCommand(ctx, "helm", "install", "test-rollout-deployment-v1", "oci://registry-1.docker.io/ashexp/rollout-deployment", "--version", "1.0.175-DEPLOY-test-app-v1.0.0", "--namespace", "ent-8-env-1", "--create-namespace")
			if err != nil {
				return err
			}
			return nil
		},
		cleanup: func() error {
			highlightedLog("Deleting helm release...")
			// Delete helm release
			_ = executeCommand(context.Background(), "helm", "uninstall", "test-rollout-deployment-v1", "--namespace", "ent-8-env-1")
			return nil
		},
	}
}

// getParentChildGvkMapForIntegrationTest
// DAG Graph Representation of Parent-Child Objects:
//
//	FlinkDeployment (v1beta1, flink.apache.org)
//	├── ConfigMaps (v1, namespace)
//	├── PersistentVolumeClaims (v1, namespace)
//	├── Pods (v1, namespace)
//	├── Services (v1, namespace)
//	├── Deployments (v1, apps, namespace)
//	│   └── ReplicaSets (v1, apps, namespace)
//	│       └── Pods (v1, namespace)
//	└── StatefulSets (v1, apps, namespace)
//	    ├── PersistentVolumeClaims (v1, namespace)
//	    └── Pods (v1, namespace)
//
//	VMCluster (v1beta1, operator.victoriametrics.com)
//	├── ConfigMaps (v1, namespace)
//	├── Secrets (v1, namespace)
//	├── ServiceAccounts (v1, namespace)
//	├── Services (v1, namespace)
//	├── Deployments (v1, apps, namespace)
//	│   └── ReplicaSets (v1, apps, namespace)
//	│       └── Pods (v1, namespace)
//	└── StatefulSets (v1, apps, namespace)
//	    ├── PersistentVolumeClaims (v1, namespace)
//	    └── Pods (v1, namespace)
//
//	VMAgent (v1beta1, operator.victoriametrics.com)
//	├── ConfigMaps (v1, namespace)
//	├── Secrets (v1, namespace)
//	├── ServiceAccounts (v1, namespace)
//	├── Services (v1, namespace)
//	└── Deployments (v1, apps, namespace)
//	    └── ReplicaSets (v1, apps, namespace)
//	        └── Pods (v1, namespace)
//
//	VMAlert (v1beta1, operator.victoriametrics.com)
//	├── ConfigMaps (v1, namespace)
//	├── Secrets (v1, namespace)
//	├── ServiceAccounts (v1, namespace)
//	├── Services (v1, namespace)
//	└── Deployments (v1, apps, namespace)
//	    └── ReplicaSets (v1, apps, namespace)
//	        └── Pods (v1, namespace)
//
//	VMAlertmanager (v1beta1, operator.victoriametrics.com)
//	├── ConfigMaps (v1, namespace)
//	├── Secrets (v1, namespace)
//	├── ServiceAccounts (v1, namespace)
//	├── Services (v1, namespace)
//	└── StatefulSets (v1, apps, namespace)
//	    ├── PersistentVolumeClaims (v1, namespace)
//	    └── Pods (v1, namespace)
//
//	Alertmanager (v1, monitoring.coreos.com)
//	├── ConfigMaps (v1, namespace)
//	└── StatefulSets (v1, apps, namespace)
//	    ├── PersistentVolumeClaims (v1, namespace)
//	    └── Pods (v1, namespace)
//
//	Prometheus (v1, monitoring.coreos.com)
//	├── ConfigMaps (v1, namespace)
//	└── StatefulSets (v1, apps, namespace)
//	    ├── PersistentVolumeClaims (v1, namespace)
//	    └── Pods (v1, namespace)
//
//	Service (v1)
//	├── Endpoints (v1, namespace)
//	├── EndpointSlices (v1, discovery.k8s.io, namespace)
//	├── EndpointSlices (v1beta1, discovery.k8s.io, namespace)
//	└── AdmissionReports (v2, kyverno.io, namespace)
//
//	StatefulSet (v1, apps)
//	├── PersistentVolumeClaims (v1, namespace)
//	└── Pods (v1, namespace)
//
//	DaemonSet (v1, apps)
//	├── Pods (v1, namespace)
//	└── ControllerRevisions (v1, apps, namespace)
//
//	ReplicaSet (v1, apps)
//	└── Pods (v1, namespace)
//
//	HorizontalPodAutoscaler (v2, autoscaling)
//	└── Pods (v1, namespace)
//
//	Job (v1, batch)
//	└── Pods (v1, namespace)
//
//	Canary (v1beta1, flagger.app)
//	├── Services (v1, namespace)
//	├── DaemonSets (v1, apps, namespace)
//	└── Deployments (v1, apps, namespace)
//	    └── ReplicaSets (v1, apps, namespace)
//	        └── Pods (v1, namespace)
//
//	Deployment (v1, apps)
//	└── ReplicaSets (v1, apps, namespace)
//	    └── Pods (v1, namespace)
//
//	Rollout (v1alpha1, argoproj.io)
//	└── ReplicaSets (v1, apps, namespace)
//	    └── Pods (v1, namespace)
//
//	MongoDBCommunity (v1, mongodbcommunity.mongodb.com)
//	└── StatefulSets (v1, apps, namespace)
//	    ├── PersistentVolumeClaims (v1, namespace)
//	    └── Pods (v1, namespace)
//
//	ScaledObject (v1alpha1, keda.sh)
//	└── HorizontalPodAutoscalers (v2, autoscaling, namespace)
//	    └── Pods (v1, namespace)
//
//	CronJob (v1, batch)
//	└── Jobs (v1, batch, namespace)
//	    └── Pods (v1, namespace)
func getParentChildGvkMapForIntegrationTest() string {
	return `[{"childObjects":[{"Group":"","Resource":"configmaps","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"persistentvolumeclaims","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"pods","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"services","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"deployments","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"statefulsets","Scope":"namespace","Version":"v1"}],"group":"flink.apache.org","kind":"FlinkDeployment","version":"v1beta1"},{"childObjects":[{"Group":"","Resource":"configmaps","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"secrets","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"serviceAccounts","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"services","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"deployments","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"replicaset","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"statefulsets","Scope":"namespace","Version":"v1"}],"group":"operator.victoriametrics.com","kind":"VMCluster","version":"v1beta1"},{"childObjects":[{"Group":"","Resource":"configmaps","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"secrets","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"serviceAccounts","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"services","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"deployments","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"replicaset","Scope":"namespace","Version":"v1"}],"group":"operator.victoriametrics.com","kind":"VMAgent","version":"v1beta1"},{"childObjects":[{"Group":"","Resource":"configmaps","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"secrets","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"serviceAccounts","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"services","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"deployments","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"replicaset","Scope":"namespace","Version":"v1"}],"group":"operator.victoriametrics.com","kind":"VMAlert","version":"v1beta1"},{"childObjects":[{"Group":"","Resource":"configmaps","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"secrets","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"serviceAccounts","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"services","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"statefulsets","Scope":"namespace","Version":"v1"}],"group":"operator.victoriametrics.com","kind":"VMAlertmanager","version":"v1beta1"},{"childObjects":[{"Group":"","Resource":"configmaps","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"statefulsets","Scope":"namespace","Version":"v1"}],"group":"monitoring.coreos.com","kind":"Alertmanager","version":"v1"},{"childObjects":[{"Group":"","Resource":"configmaps","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"statefulsets","Scope":"namespace","Version":"v1"}],"group":"monitoring.coreos.com","kind":"Prometheus","version":"v1"},{"childObjects":[{"Group":"","Resource":"endpoints","Scope":"namespace","Version":"v1"},{"Group":"discovery.k8s.io","Resource":"endpointslices","Scope":"namespace","Version":"v1"},{"Group":"discovery.k8s.io","Resource":"endpointslices","Scope":"namespace","Version":"v1beta1"},{"Group":"kyverno.io","Resource":"admissionreports","Scope":"namespace","Version":"v2"}],"group":"","kind":"Service","version":"v1"},{"childObjects":[{"Group":"","Resource":"persistentvolumeclaims","Scope":"namespace","Version":"v1"},{"Group":"","Resource":"pods","Scope":"namespace","Version":"v1"}],"group":"apps","kind":"StatefulSet","version":"v1"},{"childObjects":[{"Group":"","Resource":"pods","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"controllerrevisions","Scope":"namespace","Version":"v1"}],"group":"apps","kind":"DaemonSet","version":"v1"},{"childObjects":[{"Group":"","Resource":"pods","Scope":"namespace","Version":"v1"}],"group":"apps","kind":"ReplicaSet","version":"v1"},{"childObjects":[{"Group":"","Resource":"pods","Scope":"namespace","Version":"v1"}],"group":"autoscaling","kind":"HorizontalPodAutoscaler","version":"v2"},{"childObjects":[{"Group":"","Resource":"pods","Scope":"namespace","Version":"v1"}],"group":"batch","kind":"Job","version":"v1"},{"childObjects":[{"Group":"","Resource":"services","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"daemonsets","Scope":"namespace","Version":"v1"},{"Group":"apps","Resource":"deployments","Scope":"namespace","Version":"v1"}],"group":"flagger.app","kind":"Canary","version":"v1beta1"},{"childObjects":[{"Group":"apps","Resource":"replicasets","Scope":"namespace","Version":"v1"}],"group":"apps","kind":"Deployment","version":"v1"},{"childObjects":[{"Group":"apps","Resource":"replicasets","Scope":"namespace","Version":"v1"}],"group":"argoproj.io","kind":"Rollout","version":"v1alpha1"},{"childObjects":[{"Group":"apps","Resource":"statefulsets","Scope":"namespace","Version":"v1"}],"group":"mongodbcommunity.mongodb.com","kind":"MongoDBCommunity","version":"v1"},{"childObjects":[{"Group":"autoscaling","Resource":"horizontalpodautoscalers","Scope":"namespace","Version":"v2"}],"group":"keda.sh","kind":"ScaledObject","version":"v1alpha1"},{"childObjects":[{"Group":"batch","Resource":"jobs","Scope":"namespace","Version":"v1"}],"group":"batch","kind":"CronJob","version":"v1"}]`
}
