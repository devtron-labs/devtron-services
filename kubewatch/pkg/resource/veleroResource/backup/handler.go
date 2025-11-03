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
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/utils/storage"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	veleroBackupBean "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned"
	veleroBackupInformer "github.com/vmware-tanzu/velero/pkg/generated/informers/externalversions/velero/v1"
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
		impl.logger.Debugw("registered velero backup informer", "namespace", namespace, "time", time.Since(startTime))
	}()
	clientset := versioned.NewForConfigOrDie(k8sConfig)
	backupInformer := veleroBackupInformer.NewBackupInformer(clientset, namespace, 0, cache.Indexers{})
	_, err := backupInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			impl.logger.Debugw("velero backup add event received")
			if backupObj, ok := obj.(*veleroBackupBean.Backup); ok {
				impl.logger.Debugw("velero backup add event received", "backupObj", backupObj)
				backupChangeObj := &storage.VeleroResourceEvent{
					EventType:    storage.EventTypeAdded,
					ResourceKind: storage.ResourceBackup,
					ClusterId:    clusterLabels.ClusterId,
					ResourceName: backupObj.Name,
					Data: storage.BackupStatus{
						CompletionTimestamp: backupObj.Status.CompletionTimestamp,
						Expiration:          backupObj.Status.Expiration,
						FormatVersion:       backupObj.Status.FormatVersion,
						StartTimestamp:      backupObj.Status.StartTimestamp,
						Version:             backupObj.Status.FormatVersion,
						Phase:               backupObj.Status.Phase,
						ValidationErrors:    backupObj.Status.ValidationErrors,
					},
				}
				err := impl.sendBackupUpdate(backupChangeObj)
				if err != nil {
					impl.logger.Errorw("error in sending velero backup add event", "err", err)
				}
			} else {
				impl.logger.Errorw("velero backup object add detected, but could not cast to velero backup object", "obj", obj)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			impl.logger.Debugw("velero backup update event received")
			if oldBackupObj, ok := oldObj.(*veleroBackupBean.Backup); ok {
				if newBackupObj, ok := newObj.(*veleroBackupBean.Backup); ok {
					backupChangeObj := &storage.VeleroResourceEvent{
						EventType:    storage.EventTypeUpdated,
						ResourceKind: storage.ResourceBackup,
						ClusterId:    clusterLabels.ClusterId,
						ResourceName: newBackupObj.Name,
					}
					if isChangeInBackupObject(oldBackupObj, newBackupObj, backupChangeObj) {
						err := impl.sendBackupUpdate(backupChangeObj)
						if err != nil {
							impl.logger.Errorw("error in sending velero backup update event", "err", err)
						}
					} else {
						impl.logger.Debugw("no change in velero backup object", "oldObj", oldBackupObj, "newObj", newBackupObj)
					}
				} else {
					impl.logger.Errorw("velero backup object update detected, but could not cast to velero backup object", "newObj", newObj)
				}
			} else {
				impl.logger.Errorw("velero backup object update detected, but could not cast to velero backup object", "oldObj", oldObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			impl.logger.Debugw("velero backup delete event received")
			if backupObj, ok := obj.(*veleroBackupBean.Backup); ok {
				backupChangeObj := &storage.VeleroResourceEvent{
					EventType:    storage.EventTypeDeleted,
					ResourceKind: storage.ResourceBackup,
					ClusterId:    clusterLabels.ClusterId,
					ResourceName: backupObj.Name,
				}
				err := impl.sendBackupUpdate(backupChangeObj)
				if err != nil {
					impl.logger.Errorw("error in sending velero backup delete event", "err", err)
				}
			} else {
				impl.logger.Errorw("velero backup object delete detected, but could not cast to velero backup object", "obj", obj)
			}
		},
	})
	if err != nil {
		impl.logger.Errorw("error in adding velero backup event handler", "err", err)
		return nil, err
	}
	return backupInformer, nil
}
