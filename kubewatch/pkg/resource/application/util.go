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
	synccommon "github.com/argoproj/gitops-engine/pkg/sync/common"
	"go.uber.org/zap"
)

// isNewDeploymentHistoryFound checks
// if a new deployment is found by
// comparing the history of old and new application objects.
func isNewDeploymentHistoryFound(oldAppObj, newAppObj *v1alpha1.Application) bool {
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

func getApplicationOperationRevision(appObj *v1alpha1.Application) string {
	if appObj.Status.OperationState == nil || appObj.Status.OperationState.Operation.Sync == nil {
		return ""
	}
	return appObj.Status.OperationState.Operation.Sync.Revision
}

func getApplicationOperationPhase(appObj *v1alpha1.Application) synccommon.OperationPhase {
	if appObj.Status.OperationState == nil {
		return ""
	}
	return appObj.Status.OperationState.Phase
}

func IsApplicationObjectUpdated(logger *zap.SugaredLogger, oldAppObj, newAppObj *v1alpha1.Application) bool {
	// Check if the application sync revision has changed
	oldRevision := oldAppObj.Status.Sync.Revision
	newRevision := newAppObj.Status.Sync.Revision
	// Check if the operation revision has changed
	oldOperationRevision := getApplicationOperationRevision(oldAppObj)
	newOperationRevision := getApplicationOperationRevision(newAppObj)
	// Check if the health status has changed
	oldStatus := string(oldAppObj.Status.Health.Status)
	newStatus := string(newAppObj.Status.Health.Status)
	// Check if the operation phase has changed
	oldOperationPhase := getApplicationOperationPhase(oldAppObj)
	newOperationPhase := getApplicationOperationPhase(newAppObj)
	// Check if the last synced resources count has changed
	oldAppLastSyncedResourcesCount := getApplicationLastSyncedResourcesCount(oldAppObj)
	newAppLastSyncedResourcesCount := getApplicationLastSyncedResourcesCount(newAppObj)

	logger.Debugw("oldRevision", oldRevision, "newRevision", newRevision,
		"oldOperationRevision", oldOperationRevision, "newOperationRevision", newOperationRevision,
		"oldStatus", oldStatus, "newStatus", newStatus,
		"oldOperationPhase", oldOperationPhase, "newOperationPhase", newOperationPhase,
		"oldAppLastSyncedResourcesCount", oldAppLastSyncedResourcesCount, "newAppLastSyncedResourcesCount", newAppLastSyncedResourcesCount)

	return (oldRevision != newRevision) || (newOperationRevision != oldOperationRevision) ||
		(oldStatus != newStatus) || (oldOperationPhase != newOperationPhase) ||
		(oldAppLastSyncedResourcesCount != newAppLastSyncedResourcesCount)
}
