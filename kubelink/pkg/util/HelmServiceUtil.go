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

package util

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/common-lib/utils/k8s/health"
	"github.com/devtron-labs/kubelink/pkg/service/helmApplicationService/release"
	helmRelease "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// GetAppId returns AppID by logic
//   - format: cluster_id|namespace|release_name
func GetAppId(clusterId int32, release *release.Release) string {
	return fmt.Sprintf("%d|%s|%s", clusterId, release.Namespace, release.Name)
}

func GetMessageFromReleaseStatus(releaseStatus helmRelease.Status) string {
	switch releaseStatus {
	case helmRelease.StatusUnknown:
		return "The helmRelease is in an uncertain state"
	case helmRelease.StatusDeployed:
		return "The helmRelease has been pushed to Kubernetes"
	case helmRelease.StatusUninstalled:
		return "The helmRelease has been uninstalled from Kubernetes"
	case helmRelease.StatusSuperseded:
		return "The helmRelease object is outdated and a newer one exists"
	case helmRelease.StatusFailed:
		return "The helmRelease was not successfully deployed"
	case helmRelease.StatusUninstalling:
		return "The helmRelease uninstall operation is underway"
	case helmRelease.StatusPendingInstall:
		return "The helmRelease install operation is underway"
	case helmRelease.StatusPendingUpgrade:
		return "The helmRelease upgrade operation is underway"
	case helmRelease.StatusPendingRollback:
		return "The helmRelease rollback operation is underway"
	default:
		fmt.Println("un handled helmRelease status", releaseStatus)
	}

	return ""
}

func GetAppStatusOnBasisOfHealthyNonHealthy(healthStatusArray []*commonBean.HealthStatus) *commonBean.HealthStatusCode {
	appHealthStatus := commonBean.HealthStatusHealthy
	for _, node := range healthStatusArray {
		nodeHealth := node
		if nodeHealth == nil {
			continue
		}
		//if any node's health is worse than healthy then we break the loop and return
		if health.IsWorseStatus(health.HealthStatusCode(appHealthStatus), health.HealthStatusCode(nodeHealth.Status)) {
			appHealthStatus = nodeHealth.Status
			break
		}
	}
	return &appHealthStatus
}

func IsReleaseNotFoundError(err error) bool {
	return errors.Is(err, driver.ErrReleaseNotFound) || errors.Is(err, driver.ErrNoDeployedReleases)
}
