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

package veleroBSL

import (
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/utils/storage"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	veleroBslBean "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned"
	veleroBslInformer "github.com/vmware-tanzu/velero/pkg/generated/informers/externalversions/velero/v1"

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
		impl.logger.Debugw("registered velero bsl informer", "namespace", namespace, "time", time.Since(startTime))
	}()
	clientSet := versioned.NewForConfigOrDie(k8sConfig)
	bslInformer := veleroBslInformer.NewBackupStorageLocationInformer(clientSet, namespace, 0, cache.Indexers{})
	_, err := bslInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			impl.logger.Infow("backup storage location add detected")
			if bslObj, ok := obj.(*veleroBslBean.BackupStorageLocation); ok {
				bslChangeObj := &storage.VeleroStorageEvent[storage.LocationsStatus]{
					EventType:    storage.EventTypeAdded,
					ResourceKind: storage.ResourceBackupStorageLocation,
					ClusterId:    clusterLabels.ClusterId,
					ResourceName: bslObj.Name,
					Data: storage.LocationsStatus{
						Status: bslObj.Status,
					},
				}
				err := impl.sendBslUpdate(bslChangeObj)
				if err != nil {
					impl.logger.Errorw("error in sending backup storage location add event", "err", err)
					return
				}
			} else {
				impl.logger.Errorw("backup storage location object add detected, but could not cast to backup storage location object", "obj", obj)
				return
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			impl.logger.Infow("backup storage location update detected")
			if oldBslObj, ok := oldObj.(*veleroBslBean.BackupStorageLocation); ok {
				if newBslObj, ok := newObj.(*veleroBslBean.BackupStorageLocation); ok {
					bslChangeObj := &storage.VeleroStorageEvent[storage.LocationsStatus]{
						EventType:    storage.EventTypeUpdated,
						ResourceKind: storage.ResourceBackupStorageLocation,
						ClusterId:    clusterLabels.ClusterId,
						ResourceName: newBslObj.Name,
					}
					if isChangeInBslStatusObject(&oldBslObj.Status, &newBslObj.Status) {
						bslChangeObj.Data = storage.LocationsStatus{
							Status: newBslObj.Status,
						}
						err := impl.sendBslUpdate(bslChangeObj)
						if err != nil {
							impl.logger.Errorw("error in sending backup storage location update event", "err", err)
						}
					}
				} else {
					impl.logger.Errorw("backup storage location object update detected, but could not cast to backup storage location object", "newObj", newObj)
				}
			} else {
				impl.logger.Errorw("backup storage location object update detected, but could not cast to backup storage location object", "oldObj", oldObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			impl.logger.Infow("backup storage location delete detected")
			if bslObj, ok := obj.(*veleroBslBean.BackupStorageLocation); ok {
				bslChangeObj := &storage.VeleroStorageEvent[storage.LocationsStatus]{
					EventType:    storage.EventTypeDeleted,
					ResourceKind: storage.ResourceBackupStorageLocation,
					ClusterId:    clusterLabels.ClusterId,
					ResourceName: bslObj.Name,
				}
				err := impl.sendBslUpdate(bslChangeObj)
				if err != nil {
					impl.logger.Errorw("error in sending backup storage location delete event", "err", err)
				}
			} else {
				impl.logger.Errorw("backup storage location object delete detected, but could not cast to backup storage location object", "obj", obj)
			}
		},
	})
	if err != nil {
		impl.logger.Errorw("error in adding velero bsl event handler", "err", err)
		return nil, err
	}
	return bslInformer, nil
}
