package veleroBSL

import (
	"encoding/json"
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
				veleroStatusUpdate := &storage.VeleoroBslStatusUpdate{
					ClusterId: clusterLabels.ClusterId,
					BslName:   bslObj.Name,
					Status:    string(bslObj.Status.Phase),
				}
				impl.sendBslUpdate(veleroStatusUpdate)
			} else {
				impl.logger.Errorw("backup storage location object add detected, but could not cast to backup storage location object", "obj", obj)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			impl.logger.Infow("backup storage location update detected")
			//statusTime := time.Now()
			if oldBslObj, ok := oldObj.(*veleroBslBean.BackupStorageLocation); ok {
				if newBslObj, ok := newObj.(*veleroBslBean.BackupStorageLocation); ok {
					if oldBslObj.Status.Phase != newBslObj.Status.Phase {
						veleroStatusUpdate := &storage.VeleoroBslStatusUpdate{
							ClusterId: clusterLabels.ClusterId,
							BslName:   newBslObj.Name,
							Status:    string(newBslObj.Status.Phase),
						}
						impl.sendBslUpdate(veleroStatusUpdate)
					}
					impl.logger.Debugw("backup storage location object update detected", "oldObj", oldBslObj, "newObj", newBslObj)

				} else {
					impl.logger.Errorw("backup storage location object update detected, but could not cast to backup storage location object", "newObj", newObj)
				}
			} else {
				impl.logger.Errorw("backup storage location object update detected, but could not cast to backup storage location object", "oldObj", oldObj)
			}
		},
		DeleteFunc: func(obj interface{}) {},
	})
	if err != nil {
		impl.logger.Errorw("error in creating clientset", "err", err)
		return nil, err
	}
	return bslInformer, nil
}

func (impl *InformerImpl) sendBslUpdate(veleroStatusUpdate *storage.VeleoroBslStatusUpdate) {
	if impl.client == nil {
		impl.logger.Errorw("pubsub client is nil, skipping the publish")
		return
	}
	veleroStatusUpdateJson, err := json.Marshal(veleroStatusUpdate)
	if err != nil {
		impl.logger.Errorw("error in marshalling velero status update", "err", err)
		return
	}
	err = impl.client.Publish(pubsub.STORAGE_MODULE_TOPIC, string(veleroStatusUpdateJson))
	if err != nil {
		impl.logger.Errorw("error in publishing velero status update", "err", err)
		return
	} else {
		impl.logger.Info("velero status update sent", "veleroStatusUpdate:", string(veleroStatusUpdateJson))
		return
	}
}
