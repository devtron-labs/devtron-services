package veleroVSL

import (
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/utils/storage"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	veleroVslBean "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
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
		impl.logger.Debugw("registered velero vsl informer", "namespace", namespace, "time", time.Since(startTime))
	}()
	clientSet := versioned.NewForConfigOrDie(k8sConfig)
	vslInformer := veleroVslInformer.NewVolumeSnapshotLocationInformer(clientSet, namespace, 0, cache.Indexers{})
	_, err := vslInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			impl.logger.Debugw("velero vsl add event received")
			if vslObj, ok := obj.(*veleroVslBean.VolumeSnapshotLocation); ok {
				impl.logger.Infow("velero vsl add event received", "vslObj", vslObj)
				vslChangeObj := &storage.VeleroStorageEvent[storage.LocationsStatus]{
					EventType:    storage.EventTypeAdded,
					ResourceKind: storage.ResourceVolumeSnapshotLocation,
					ClusterId:    clusterLabels.ClusterId,
					ResourceName: vslObj.Name,
					Data: storage.LocationsStatus{
						Provider: vslObj.Spec.Provider,
					},
				}
				err := impl.sendVslUpdate(vslChangeObj)
				if err != nil {
					impl.logger.Errorw("error in sending velero vsl add event", "err", err)
				}
			} else {
				impl.logger.Errorw("velero vsl object add detected, but could not cast to velero vsl object", "obj", obj)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			impl.logger.Debugw("velero vsl update event received")
			if oldVslObj, ok := oldObj.(*veleroVslBean.VolumeSnapshotLocation); ok {
				if newVslObj, ok := newObj.(*veleroVslBean.VolumeSnapshotLocation); ok {
					vslChangeObj := &storage.VeleroStorageEvent[storage.LocationsStatus]{
						EventType:    storage.EventTypeUpdated,
						ResourceKind: storage.ResourceVolumeSnapshotLocation,
						ClusterId:    clusterLabels.ClusterId,
						ResourceName: newVslObj.Name,
					}
					if isChangeInVslObject(oldVslObj, newVslObj, vslChangeObj) {
						err := impl.sendVslUpdate(vslChangeObj)
						if err != nil {
							impl.logger.Errorw("error in sending velero vsl update event", "err", err)
						}
					} else {
						impl.logger.Debugw("no change in velero vsl object", "oldObj", oldVslObj, "newObj", newVslObj)
					}
				} else {
					impl.logger.Errorw("velero vsl object update detected, but could not cast to velero vsl object", "newObj", newObj)
				}
			} else {
				impl.logger.Errorw("velero vsl object update detected, but could not cast to velero vsl object", "oldObj", oldObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			impl.logger.Debugw("velero vsl delete event received")
			if vslObj, ok := obj.(*veleroVslBean.VolumeSnapshotLocation); ok {
				vslChangeObj := &storage.VeleroStorageEvent[storage.LocationsStatus]{
					EventType:    storage.EventTypeDeleted,
					ResourceKind: storage.ResourceVolumeSnapshotLocation,
					ClusterId:    clusterLabels.ClusterId,
					ResourceName: vslObj.Name,
				}
				err := impl.sendVslUpdate(vslChangeObj)
				if err != nil {
					impl.logger.Errorw("error in sending velero vsl delete event", "err", err)
				}
			} else {
				impl.logger.Errorw("velero vsl object delete detected, but could not cast to velero vsl object", "obj", obj)
			}
		},
	})
	if err != nil {
		impl.logger.Errorw("error in adding velero vsl event handler", "err", err)
		return nil, err
	}
	return vslInformer, nil
}
