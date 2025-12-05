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

package veleroBackupSchedule

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/utils/storage"
	"github.com/pkg/errors"
	veleroBackupScheduleBean "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

func (impl *InformerImpl) sendBackupScheduleUpdate(backupScheduleChangeObj *storage.VeleroResourceEvent) error {
	if impl.client == nil {
		impl.logger.Errorw("pubsub client is nil - STORAGE_MODULE_TOPIC, skipping the publish")
		return errors.New("pubsub client is nil - STORAGE_MODULE_TOPIC, skipping the publish")
	}
	backupScheduleChangeObjByte, err := json.Marshal(backupScheduleChangeObj)
	if err != nil {
		impl.logger.Errorw("error in marshalling velero backup schedule status update", "err", err)
		return err
	}
	err = impl.client.Publish(pubsub.STORAGE_MODULE_TOPIC, string(backupScheduleChangeObjByte))
	if err != nil {
		impl.logger.Errorw("error in publishing velero status update", "err", err)
		return err
	} else {
		impl.logger.Info("velero status update sent", "veleroStatusUpdate:", string(backupScheduleChangeObjByte))
		return nil
	}
}

func isChangeInBackupScheduleObject(oldObj, newObj *veleroBackupScheduleBean.Schedule) bool {
	return oldObj.Status.Phase != newObj.Status.Phase ||
		!oldObj.Status.LastBackup.Equal(newObj.Status.LastBackup) ||
		oldObj.Status.LastSkipped.Equal(newObj.Status.LastSkipped) ||
		len(oldObj.Status.ValidationErrors) != len(newObj.Status.ValidationErrors)
}
