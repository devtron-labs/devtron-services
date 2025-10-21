package veleroVslInformer

import (
	"fmt"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	informerErr "github.com/devtron-labs/kubewatch/pkg/informer/errors"
	"golang.org/x/exp/maps"
)

func (impl *InformerImpl) getVeleroVslStopper(clusterId int) (*informerBean.SharedStopper, bool) {
	stopper, ok := impl.veleroVslInformerStopper[clusterId]
	if ok {
		return stopper, stopper.HasInformer()
	}
	return stopper, false
}
func (impl *InformerImpl) checkAndGetStopChannel(clusterLabels *informerBean.ClusterLabels) (chan struct{}, error) {
	stopChannel := make(chan struct{})
	stopper, ok := impl.getVeleroVslStopper(clusterLabels.ClusterId)
	if ok && stopper.HasInformer() {
		impl.logger.Debug(fmt.Sprintf("velero vsl informer for %s already exist", clusterLabels.ClusterName))
		// TODO: should we return the stop channel here?
		return nil, informerErr.AlreadyExists
	}
	stopper = stopper.GetStopper(stopChannel)
	impl.veleroVslInformerStopper[clusterLabels.ClusterId] = stopper
	return stopChannel, nil
}
func (impl *InformerImpl) getStoppableClusterIds() []int {
	return maps.Keys(impl.veleroVslInformerStopper)
}
