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

func (impl *InformerImpl) sendBackupUpdate(backupChangeObj *storage.VeleroStorageEvent[storage.BackupStatus]) error {
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

// TODO: Currently we are only intercepting changes of the status section, but do we nee to also to intercept the specs section
func isChangeInBackupObject(oldObj, newObj *veleroBackupBean.Backup, backupChangeObj *storage.VeleroStorageEvent[storage.BackupStatus]) bool {
	if oldObj.Status.Progress == newObj.Status.Progress && oldObj.Status.Phase == newObj.Status.Phase &&
		oldObj.Status.CompletionTimestamp.Equal(newObj.Status.CompletionTimestamp) && oldObj.Status.Expiration.Equal(newObj.Status.Expiration) &&
		oldObj.Status.FormatVersion == newObj.Status.FormatVersion && oldObj.Status.StartTimestamp.Equal(newObj.Status.StartTimestamp) &&
		oldObj.Status.FormatVersion == newObj.Status.FormatVersion {
		return false
	} else {
		backupChangeObj.Data = storage.BackupStatus{
			Phase:               newObj.Status.Phase,
			CompletionTimestamp: newObj.Status.CompletionTimestamp,
			Expiration:          newObj.Status.Expiration,
			FormatVersion:       newObj.Status.FormatVersion,
			Progress:            *newObj.Status.Progress,
			StartTimestamp:      newObj.Status.StartTimestamp,
			Version:             newObj.Status.FormatVersion,
		}
	}
	return true
}
