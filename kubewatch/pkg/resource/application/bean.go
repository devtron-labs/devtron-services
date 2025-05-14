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

package application

import (
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"time"
)

type applicationDetail struct {
	Application *v1alpha1.Application `json:"application"`
	StatusTime  time.Time             `json:"statusTime"`
	ClusterId   int                   `json:"clusterId"`
}

// isNewDeploymentFound checks if a new deployment is found by comparing the history of old and new application objects.
func isNewDeploymentFound(oldAppObj, newAppObj *v1alpha1.Application) bool {
	if len(oldAppObj.Status.History) < len(newAppObj.Status.History) {
		return true
	} else if len(oldAppObj.Status.History) != 0 && len(oldAppObj.Status.History) == len(newAppObj.Status.History) {
		return oldAppObj.Status.History.LastRevisionHistory().ID < newAppObj.Status.History.LastRevisionHistory().ID
	}
	return false
}

// getApplicationLastSyncedResourcesCount returns the count of resources that were last synced in the application.
func getApplicationLastSyncedResourcesCount(appObj *v1alpha1.Application) int {
	if appObj.Status.OperationState == nil || appObj.Status.OperationState.SyncResult == nil {
		return 0
	}
	return len(appObj.Status.OperationState.SyncResult.Resources)
}
