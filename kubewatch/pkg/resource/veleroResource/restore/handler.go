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

package veleroRestore

import (
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/utils/storage"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	veleroRestoreBean "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned"
	veleroRestoreInformer "github.com/vmware-tanzu/velero/pkg/generated/informers/externalversions/velero/v1"
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
		impl.logger.Debugw("registered velero restore informer", "namespace", namespace, "time", time.Since(startTime))
	}()

	clientset := versioned.NewForConfigOrDie(k8sConfig)
	restoreInformer := veleroRestoreInformer.NewRestoreInformer(clientset, namespace, 0, cache.Indexers{})
	_, err := restoreInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			impl.logger.Debugw("velero restore add event received")
			if restoreObj, ok := obj.(*veleroRestoreBean.Restore); ok {
				impl.logger.Debugw("velero restore add event received", "restoreObj", restoreObj)
				restoreChangeObj := &storage.VeleroStorageEvent[storage.RestoreStatus]{
					EventType:    storage.EventTypeAdded,
					ResourceKind: storage.ResourceRestore,
					ClusterId:    clusterLabels.ClusterId,
					ResourceName: restoreObj.Name,
					Data: storage.RestoreStatus{
						BackupName:     restoreObj.Spec.BackupName,
						ScheduleName:   restoreObj.Spec.ScheduleName,
						StartTimestamp: restoreObj.Status.StartTimestamp,
						Phase:          restoreObj.Status.Phase,
						Progress:       *restoreObj.Status.Progress,
					},
				}
				err := impl.sendRestoreUpdate(restoreChangeObj)
				if err != nil {
					impl.logger.Errorw("error in sending velero restore add event", "err", err)
				}
			} else {
				impl.logger.Errorw("velero restore object add detected, but could not cast to velero restore object", "obj", obj)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			impl.logger.Debugw("velero restore update event received")
			if oldRestoreObj, ok := oldObj.(*veleroRestoreBean.Restore); ok {
				if newRestoreObj, ok := newObj.(*veleroRestoreBean.Restore); ok {
					restoreChangeObj := &storage.VeleroStorageEvent[storage.RestoreStatus]{
						EventType:    storage.EventTypeUpdated,
						ResourceKind: storage.ResourceRestore,
						ClusterId:    clusterLabels.ClusterId,
						ResourceName: newRestoreObj.Name,
					}
					if isChangeInRestoreObject(oldRestoreObj, newRestoreObj, restoreChangeObj) {
						err := impl.sendRestoreUpdate(restoreChangeObj)
						if err != nil {
							impl.logger.Errorw("error in sending velero restore update event", "err", err)
						}
					} else {
						impl.logger.Debugw("no change in velero restore object", "oldObj", oldRestoreObj, "newObj", newRestoreObj)
					}
				} else {
					impl.logger.Errorw("velero restore object update detected, but could not cast to velero restore object", "newObj", newObj)
				}
			} else {
				impl.logger.Errorw("velero restore object update detected, but could not cast to velero restore object", "oldObj", oldObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			impl.logger.Debugw("velero restore delete event received")
			if restoreObj, ok := obj.(*veleroRestoreBean.Restore); ok {
				restoreChangeObj := &storage.VeleroStorageEvent[storage.RestoreStatus]{
					EventType:    storage.EventTypeDeleted,
					ResourceKind: storage.ResourceRestore,
					ClusterId:    clusterLabels.ClusterId,
					ResourceName: restoreObj.Name,
				}
				err := impl.sendRestoreUpdate(restoreChangeObj)
				if err != nil {
					impl.logger.Errorw("error in sending velero restore delete event", "err", err)
				}
			} else {
				impl.logger.Errorw("velero restore object delete detected, but could not cast to velero restore object", "obj", obj)
			}
		},
	})
	if err != nil {
		impl.logger.Errorw("error in adding velero restore event handler", "err", err)
		return nil, err
	}
	return restoreInformer, nil
}
