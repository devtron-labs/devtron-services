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

package cluster

import (
	"errors"
	"fmt"
	informerBean "github.com/devtron-labs/common-lib/informer"
	repository "github.com/devtron-labs/kubewatch/pkg/cluster"
	"github.com/devtron-labs/kubewatch/pkg/informer/bean"
	informerErr "github.com/devtron-labs/kubewatch/pkg/informer/errors"
	"github.com/devtron-labs/kubewatch/pkg/middleware"
	"github.com/go-pg/pg"
	coreV1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"strconv"
	"time"
)

func (impl *InformerImpl) getStopChannel(informerFactory kubeinformers.SharedInformerFactory, clusterLabels *bean.ClusterLabels) (chan struct{}, error) {
	stopChannel := make(chan struct{})
	stopper, found := impl.getClusterInformerStopper()
	if found {
		impl.logger.Debug(fmt.Sprintf("cluster informer for %s already exist", clusterLabels.ClusterName))
		return stopChannel, informerErr.AlreadyExists
	}
	stopper = stopper.GetStopper(informerFactory, stopChannel)
	impl.setClusterInformerStopper(stopper)
	return stopChannel, nil
}

func (impl *InformerImpl) getClusterInformerStopper() (*bean.FactoryStopper, bool) {
	return impl.clusterInformerStopper, impl.clusterInformerStopper.HasInformer()
}

func (impl *InformerImpl) setClusterInformerStopper(stopper *bean.FactoryStopper) {
	impl.clusterInformerStopper = stopper
}

func (impl *InformerImpl) stopDevtronClusterWatcher() error {
	stopper, found := impl.getClusterInformerStopper()
	if found {
		stopper.Stop()
		impl.logger.Info("cluster informer stopped for default cluster")
	}
	return nil
}

func (impl *InformerImpl) startClientInformers(clusterInfo *repository.Cluster) error {
	for supportedClient := range bean.SupportedClientMap {
		clientAdvisor, err := impl.GetClient(supportedClient, clusterInfo)
		if err != nil {
			impl.logger.Errorw("error in getting client advisor", "supportedClient", supportedClient, "err", err)
			return err
		}
		err = clientAdvisor.StartInformerForCluster(clusterInfo)
		if err != nil && !errors.Is(err, informerErr.AlreadyExists) {
			impl.logger.Errorw("error in starting informer for cluster", "supportedClient", supportedClient, "clusterId", clusterInfo.Id, "err", err)
			return err
		} else if errors.Is(err, informerErr.AlreadyExists) {
			impl.logger.Warnw("informer already exist for cluster", "supportedClient", supportedClient, "clusterId", clusterInfo.Id)
		}
	}
	return nil
}

func (impl *InformerImpl) stopInformersForCluster(clusterId int) error {
	for supportedClient := range bean.SupportedClientMap {
		clientAdvisor, err := impl.GetClientStopper(supportedClient)
		if err != nil {
			impl.logger.Errorw("error in getting client advisor", "supportedClient", supportedClient, "err", err)
			return err
		}
		err = clientAdvisor.StopInformerForCluster(clusterId)
		if err != nil {
			impl.logger.Errorw("error in stopping informer for cluster", "supportedClient", supportedClient, "clusterId", clusterId, "err", err)
			// ignore error and continue with other clients
		}
	}
	return nil
}

func (impl *InformerImpl) startInformerForCluster(clusterInfo *repository.Cluster) error {
	startTime := time.Now()
	defer func() {
		impl.logger.Debugw("time taken to start informers for cluster", "clusterId", clusterInfo.Id, "time", time.Since(startTime))
	}()
	if len(clusterInfo.ErrorInConnecting) > 0 {
		impl.logger.Debugw("cluster is not reachable", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
		middleware.IncUnreachableCluster(bean.NewClusterLabels(clusterInfo.ClusterName, clusterInfo.Id))
	}
	// FIXME: If cluster is not reachable, we won't be able to start the informer for it.
	// But once orchestrator is able to connect to the cluster, we should start the informer using it's event.
	// Currently, we are not handling this case.
	err := impl.startClientInformers(clusterInfo)
	if err != nil {
		return err
	}
	return nil
}

func (impl *InformerImpl) handleClusterChangeEvent(cmObject *coreV1.ConfigMap) error {
	if labelValue, exists := cmObject.Labels[informerBean.ClusterModifyEventSecretTypeKey]; !exists || labelValue != informerBean.ClusterModifyEventCmLabelValue {
		impl.logger.Infow("label value not found in cm, hence ignoring cluster change event", "labelKey", informerBean.ClusterModifyEventSecretTypeKey, "labelValue", labelValue)
		return nil
	}
	data := cmObject.Data
	action := data[informerBean.CmFieldAction]
	id := string(data[informerBean.CmFieldClusterId])
	clusterId, convErr := strconv.Atoi(id)
	if convErr != nil {
		impl.logger.Errorw("error in converting cluster id to int", "clusterId", id, "err", convErr)
		return convErr
	}
	if string(action) == informerBean.ClusterActionAdd {
		if err := impl.reloadInformerForCluster(clusterId); err != nil {
			impl.logger.Errorw("error in starting informer for cluster", "clusterId", clusterId, "err", err)
			return err
		}
	} else if string(action) == informerBean.ClusterActionUpdate {
		if err := impl.syncMultiClusterInformer(clusterId); err != nil {
			impl.logger.Errorw("error in updating informer for cluster", "id", clusterId, "err", err)
			return err
		}
	}
	return nil
}

func (impl *InformerImpl) handleClusterDeleteEvent(cmObject *coreV1.ConfigMap) error {
	if labelValue, exists := cmObject.Labels[informerBean.ClusterModifyEventSecretTypeKey]; !exists || labelValue != informerBean.ClusterModifyEventCmLabelValue {
		impl.logger.Infow("label value not found in cm, hence ignoring cluster delete event", "labelKey", informerBean.ClusterModifyEventSecretTypeKey, "labelValue", labelValue)
		return nil
	}
	data := cmObject.Data
	action := data[informerBean.CmFieldAction]
	id := string(data[informerBean.CmFieldClusterId])
	clusterId, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	if string(action) == informerBean.ClusterActionDelete {
		if err = impl.handleClusterDelete(clusterId); err != nil {
			impl.logger.Errorw("error in handling cluster delete event", "clusterId", clusterId, "err", err)
			return err
		}
	}
	return nil
}

func (impl *InformerImpl) handleClusterDelete(clusterId int) error {
	deleteClusterInfo, err := impl.clusterRepository.FindByIdWithActiveFalse(clusterId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching cluster by id", "cluster-id ", clusterId, "err", err)
		return err
	} else if errors.Is(err, pg.ErrNoRows) {
		impl.logger.Warnw("cluster not found", "clusterId", clusterId)
		return nil
	}
	if stopErr := impl.stopInformersForCluster(deleteClusterInfo.Id); stopErr != nil {
		impl.logger.Errorw("error in stopping informer for cluster", "clusterId", deleteClusterInfo.Id, "err", stopErr)
		return stopErr
	}
	return nil
}

func (impl *InformerImpl) syncMultiClusterInformer(clusterId int) error {
	clusterInfo, err := impl.clusterRepository.FindById(clusterId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Error("error in fetching cluster info by id", "err", err)
		return err
	} else if errors.Is(err, pg.ErrNoRows) {
		impl.logger.Warnw("cluster not found", "clusterId", clusterId)
		return nil
	}
	// before creating a new informer for cluster, close the existing one
	impl.logger.Debugw("stopping informer for cluster", "cluster-name", clusterInfo.ClusterName, "cluster-id", clusterInfo.Id)
	if stopErr := impl.stopInformersForCluster(clusterInfo.Id); stopErr != nil {
		impl.logger.Errorw("error in stopping informer for cluster", "clusterId", clusterInfo.Id, "err", stopErr)
		return stopErr
	}
	impl.logger.Debugw("informer stopped", "cluster-name", clusterInfo.ClusterName, "cluster-id", clusterInfo.Id)
	// create new informer for cluster with new config
	err = impl.reloadInformerForCluster(clusterInfo.Id)
	if err != nil {
		impl.logger.Errorw("error in starting informer for cluster", "clusterId", clusterInfo.Id, "err", err)
		return err
	}
	return nil
}

func (impl *InformerImpl) reloadInformerForCluster(clusterId int) error {
	startTime := time.Now()
	defer func() {
		impl.logger.Debugw("time taken to reload informer for cluster", "clusterId", clusterId, "time", time.Since(startTime))
	}()

	clusterInfo, err := impl.clusterRepository.FindById(clusterId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching cluster", "clusterId", clusterId, "err", err)
		return err
	} else if errors.Is(err, pg.ErrNoRows) {
		impl.logger.Warnw("cluster not found", "clusterId", clusterId)
		return nil
	}
	return impl.startInformerForCluster(clusterInfo)
}
