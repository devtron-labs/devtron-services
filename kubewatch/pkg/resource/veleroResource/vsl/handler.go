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
				vslChangeObj := storage.NewVeleroResourceEvent().
					SetEventType(storage.EventTypeAdded).
					SetResourceKind(storage.ResourceVolumeSnapshotLocation).
					SetClusterId(clusterLabels.ClusterId).
					SetResourceName(vslObj.Name)
				err := impl.sendVslUpdate(vslChangeObj)
				if err != nil {
					impl.logger.Errorw("error in sending velero vsl add event", "err", err)
				}
			} else {
				impl.logger.Errorw("velero vsl object add detected, but could not cast to velero vsl object", "obj", obj)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {},
		DeleteFunc: func(obj interface{}) {
			impl.logger.Debugw("velero vsl delete event received")
			if vslObj, ok := obj.(*veleroVslBean.VolumeSnapshotLocation); ok {
				vslChangeObj := storage.NewVeleroResourceEvent().
					SetEventType(storage.EventTypeDeleted).
					SetResourceKind(storage.ResourceVolumeSnapshotLocation).
					SetClusterId(clusterLabels.ClusterId).
					SetResourceName(vslObj.Name)
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
