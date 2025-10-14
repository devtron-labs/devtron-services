package veleroBSL

import (
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
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
		AddFunc: func(obj interface{}) {},
		UpdateFunc: func(oldObj, newObj interface{}) {
			impl.logger.Infow("backup storage location update detected")
			statusTime := time.Now()
			if oldBslObj, ok := oldObj.(*veleroBslBean.BackupStorageLocation); ok {
				if newBslObj, ok := newObj.(*veleroBslBean.BackupStorageLocation); ok {
					if oldBslObj.Status.Phase != newBslObj.Status.Phase {
						impl.sendBslUpdate(clusterLabels.ClusterId, newBslObj, statusTime)
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
	return nil, nil
}

func (impl *InformerImpl) sendBslUpdate(clusterId int, bsl *veleroBslBean.BackupStorageLocation, statusTime time.Time) {
	// implement message publishing logic after creating stream and consumer for natsjstream
}
