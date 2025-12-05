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

package veleroBackup

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/utils/storage"
	"github.com/pkg/errors"
	veleroBackupBean "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

func (impl *InformerImpl) sendBackupUpdate(backupChangeObj *storage.VeleroResourceEvent) error {
	if impl.client == nil {
		impl.logger.Errorw("pubsub client is nil - STORAGE_MODULE_TOPIC, skipping the publish")
		return errors.New("pubsub client is nil - STORAGE_MODULE_TOPIC, skipping the publish")
	}
	backupChangeObjByte, err := json.Marshal(backupChangeObj)
	if err != nil {
		impl.logger.Errorw("error in marshalling velero status update", "err", err)
		return err
	}
	err = impl.client.Publish(pubsub.STORAGE_MODULE_TOPIC, string(backupChangeObjByte))
	if err != nil {
		impl.logger.Errorw("error in publishing velero status update", "err", err)
		return err
	} else {
		impl.logger.Info("velero status update sent", "veleroStatusUpdate:", string(backupChangeObjByte))
		return nil
	}
}

func isChangeInBackupObject(oldObj, newObj *veleroBackupBean.Backup) bool {
	return oldObj.Status.Version != newObj.Status.Version ||
		oldObj.Status.FormatVersion != newObj.Status.FormatVersion ||
		!oldObj.Status.Expiration.Equal(newObj.Status.Expiration) ||
		oldObj.Status.Phase != newObj.Status.Phase ||
		len(oldObj.Status.ValidationErrors) != len(newObj.Status.ValidationErrors) ||
		!oldObj.Status.StartTimestamp.Equal(newObj.Status.StartTimestamp) ||
		!oldObj.Status.CompletionTimestamp.Equal(newObj.Status.CompletionTimestamp) ||
		oldObj.Status.VolumeSnapshotsAttempted != newObj.Status.VolumeSnapshotsAttempted ||
		oldObj.Status.VolumeSnapshotsCompleted != newObj.Status.VolumeSnapshotsCompleted ||
		oldObj.Status.FailureReason != newObj.Status.FailureReason ||
		oldObj.Status.Warnings != newObj.Status.Warnings ||
		oldObj.Status.Errors != newObj.Status.Errors ||
		(oldObj.Status.Progress == nil && newObj.Status.Progress != nil) ||
		(oldObj.Status.Progress != nil && newObj.Status.Progress != nil &&
			(oldObj.Status.Progress.ItemsBackedUp != newObj.Status.Progress.ItemsBackedUp ||
				oldObj.Status.Progress.TotalItems != newObj.Status.Progress.TotalItems)) ||
		oldObj.Status.CSIVolumeSnapshotsAttempted != newObj.Status.CSIVolumeSnapshotsAttempted ||
		oldObj.Status.CSIVolumeSnapshotsCompleted != newObj.Status.CSIVolumeSnapshotsCompleted ||
		oldObj.Status.BackupItemOperationsAttempted != newObj.Status.BackupItemOperationsAttempted ||
		oldObj.Status.BackupItemOperationsCompleted != newObj.Status.BackupItemOperationsCompleted ||
		oldObj.Status.BackupItemOperationsFailed != newObj.Status.BackupItemOperationsFailed ||
		(oldObj.Status.HookStatus == nil && newObj.Status.HookStatus != nil) ||
		(oldObj.Status.HookStatus != nil && newObj.Status.HookStatus != nil &&
			oldObj.Status.HookStatus.HooksAttempted != newObj.Status.HookStatus.HooksAttempted ||
			oldObj.Status.HookStatus.HooksFailed != newObj.Status.HookStatus.HooksFailed)
}
