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

package helper

import (
	"context"
	"fmt"
	cicxt "github.com/devtron-labs/ci-runner/executor/context"
	"os/exec"
	"testing"
)

func getDockerHelperImpl() *DockerHelperImpl {
	commandExecutorImpl := NewCommandExecutorImpl()
	dockerHelperImpl := NewDockerHelperImpl(commandExecutorImpl)
	return dockerHelperImpl
}

func TestCreateBuildXK8sDriver(t *testing.T) {
	buildxOpts := make([]map[string]string, 0)
	buildxOpts = append(buildxOpts, map[string]string{"node": "builder-amd64", "driverOptions": "namespace=devtron-ci,nodeselector=kubernetes.io/arch:amd64"})
	buildxOpts = append(buildxOpts, map[string]string{"node": "builder-amd64-test", "driverOptions": "namespace=devtron-ci,nodeselector=kubernetes.io/arch:amd64"})
	dockerBuildConfig := &DockerBuildConfig{
		BuildxK8sDriverOptions: buildxOpts,
		TargetPlatform:         "linux/amd64",
	}
	eligibleK8sNodes := dockerBuildConfig.GetEligibleK8sDriverNodes()
	impl := getDockerHelperImpl()
	ciContext := cicxt.BuildCiContext(context.Background(), true)
	_, err := impl.createBuildxBuilderWithK8sDriver(ciContext, false, "", "", eligibleK8sNodes, 1, 1)
	t.Cleanup(func() {
		buildxDelete := fmt.Sprintf("docker buildx rm %s", BUILDX_K8S_DRIVER_NAME)
		builderRemoveCmd := exec.Command("/bin/sh", "-c", buildxDelete)
		builderRemoveCmd.Run()
	})
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
}

func TestCleanBuildxK8sDriver(t *testing.T) {
	buildxOpts := make([]map[string]string, 0)
	buildxOpts = append(buildxOpts, map[string]string{"node": "", "driverOptions": "namespace=devtron-ci,nodeselector=kubernetes.io/arch:amd64"})
	buildxOpts = append(buildxOpts, map[string]string{"node": "builder-amd64-test", "driverOptions": "namespace=devtron-ci,nodeselector=kubernetes.io/arch:amd64"})
	dockerBuildConfig := &DockerBuildConfig{
		BuildxK8sDriverOptions: buildxOpts,
		TargetPlatform:         "linux/amd64",
	}
	eligibleK8sNodes := dockerBuildConfig.GetEligibleK8sDriverNodes()
	impl := getDockerHelperImpl()
	ciContext := cicxt.BuildCiContext(context.Background(), true)
	_, err := impl.createBuildxBuilderWithK8sDriver(ciContext, false, "", "", eligibleK8sNodes, 1, 1)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}

	err = impl.CleanBuildxK8sDriver(ciContext, buildxOpts)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}

}
