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
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/utils/storage"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	veleroBackupScheduleBean "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned"
	veleroVslInformer "github.com/vmware-tanzu/velero/pkg/generated/informers/externalversions/velero/v1"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"time"
)

type InformerImpl struct {
	logger *zap.SugaredLogger
	client *pubsub.PubSubClientServiceImpl
}

func NewInformerImpl(logger *zap.SugaredLogger,
	client *pubsub.PubSubClientServiceImpl) *InformerImpl {
	return &InformerImpl{
		logger: logger,
		client: client,
	}
}

func (impl *InformerImpl) GetSharedInformer(clusterLabels *informerBean.ClusterLabels, namespace string, k8sConfig *rest.Config) (cache.SharedIndexInformer, error) {
	startTime := time.Now()
	defer func() {
		impl.logger.Debugw("registered velero backup schedule informer", "namespace", namespace, "time", time.Since(startTime))
	}()
	clientSet := versioned.NewForConfigOrDie(k8sConfig)
	backupScheduleInformer := veleroVslInformer.NewScheduleInformer(clientSet, namespace, 0, cache.Indexers{})
	_, err := backupScheduleInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			impl.logger.Debugw("velero backup schedule added", "obj", obj)
			if backupSchedule, ok := obj.(*veleroBackupScheduleBean.Schedule); ok {
				impl.logger.Debugw("velero backup schedule added", "backupSchedule", backupSchedule)
				backupScheduleChangeObj := &storage.VeleroStorageEvent[storage.BackupScheduleStatus]{
					EventType:    storage.EventTypeAdded,
					ResourceKind: storage.ResourceBackupSchedule,
					ClusterId:    clusterLabels.ClusterId,
					ResourceName: backupSchedule.Name,
					Data: storage.BackupScheduleStatus{
						Status:               backupSchedule.Spec.Paused,
						StorageLocation:      backupSchedule.Spec.Template.StorageLocation,
						Cron:                 backupSchedule.Spec.Schedule,
						LastBackupTimestamp:  backupSchedule.Status.LastBackup,
						LastSkippedTimestamp: backupSchedule.Status.LastSkipped,
					},
				}
				err := impl.sendBackupScheduleUpdate(backupScheduleChangeObj)
				if err != nil {
					impl.logger.Errorw("error in sending velero backup schedule update", "err", err)
				}
			} else {
				impl.logger.Errorw("error in casting velero backup schedule", "obj", obj)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			impl.logger.Debugw("velero backup schedule updated", "oldObj", oldObj, "newObj", newObj)
			if oldBackupSchedule, ok := oldObj.(*veleroBackupScheduleBean.Schedule); ok {
				if newBackupSchedule, ok := newObj.(*veleroBackupScheduleBean.Schedule); ok {
					backupScheduleChangeObj := &storage.VeleroStorageEvent[storage.BackupScheduleStatus]{
						EventType:    storage.EventTypeUpdated,
						ResourceKind: storage.ResourceBackupSchedule,
						ClusterId:    clusterLabels.ClusterId,
						ResourceName: newBackupSchedule.Name,
					}
					if isChangeInBackupScheduleObject(oldBackupSchedule, newBackupSchedule, backupScheduleChangeObj) {
						err := impl.sendBackupScheduleUpdate(backupScheduleChangeObj)
						if err != nil {
							impl.logger.Errorw("error in sending velero backup schedule update", "err", err)
						}
					} else {
						impl.logger.Debugw("no change in velero backup schedule, skipping the publish", "oldObj", oldObj, "newObj", newObj)
					}
				} else {
					impl.logger.Errorw("error in casting velero backup schedule", "newObj", newObj)
				}
			} else {
				impl.logger.Errorw("error in casting velero backup schedule", "oldObj", oldObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			impl.logger.Debugw("velero backup schedule deleted", "obj", obj)
			if backupSchedule, ok := obj.(*veleroBackupScheduleBean.Schedule); ok {
				backupScheduleChangeObj := &storage.VeleroStorageEvent[storage.BackupScheduleStatus]{
					EventType:    storage.EventTypeDeleted,
					ResourceKind: storage.ResourceBackupSchedule,
					ClusterId:    clusterLabels.ClusterId,
					ResourceName: backupSchedule.Name,
				}
				err := impl.sendBackupScheduleUpdate(backupScheduleChangeObj)
				if err != nil {
					impl.logger.Errorw("error in sending velero backup schedule update", "err", err)
				}
			} else {
				impl.logger.Errorw("error in casting velero backup schedule", "obj", obj)
			}
		},
	})
	if err != nil {
		impl.logger.Errorw("error in adding event handler for velero backup schedule", "err", err)
		return nil, err
	}
	return backupScheduleInformer, nil
}
