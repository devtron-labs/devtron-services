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

package argoWf

import (
	"github.com/devtron-labs/common-lib/async"
	repository "github.com/devtron-labs/kubewatch/pkg/cluster"
	"github.com/devtron-labs/kubewatch/pkg/config"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	"github.com/devtron-labs/kubewatch/pkg/resource"
	"github.com/devtron-labs/kubewatch/pkg/utils"
	"go.uber.org/zap"
)

type InformerImpl struct {
	logger                  *zap.SugaredLogger
	appConfig               *config.AppConfig
	k8sUtil                 utils.K8sUtil
	informerClient          resource.InformerClient
	asyncRunnable           *async.Runnable
	argoWfCiInformerStopper map[int]*informerBean.SharedStopper
	argoWfCdInformerStopper map[int]*informerBean.SharedStopper
}

func NewInformerImpl(logger *zap.SugaredLogger,
	appConfig *config.AppConfig,
	k8sUtil utils.K8sUtil,
	informerClient resource.InformerClient,
	asyncRunnable *async.Runnable) *InformerImpl {
	return &InformerImpl{
		logger:                  logger,
		appConfig:               appConfig,
		k8sUtil:                 k8sUtil,
		informerClient:          informerClient,
		asyncRunnable:           asyncRunnable,
		argoWfCiInformerStopper: make(map[int]*informerBean.SharedStopper),
		argoWfCdInformerStopper: make(map[int]*informerBean.SharedStopper),
	}
}

func (impl *InformerImpl) StartInformerForCluster(clusterInfo *repository.Cluster) error {
	if err := impl.startCiArgoWfInformer(clusterInfo); err != nil {
		return err
	}
	if err := impl.startCdArgoWfInformer(clusterInfo); err != nil {
		return err
	}
	return nil
}

func (impl *InformerImpl) StopInformerForCluster(clusterId int) error {
	stopper, found := impl.getCiArgoWfStopper(clusterId)
	if found {
		stopper.Stop()
		delete(impl.argoWfCiInformerStopper, clusterId)
		impl.logger.Infow("argo workflow ci informer stopped for cluster", "clusterId", clusterId)
	}
	stopper, found = impl.getCdArgoWfStopper(clusterId)
	if found {
		stopper.Stop()
		delete(impl.argoWfCdInformerStopper, clusterId)
		impl.logger.Infow("argo workflow cd informer stopped for cluster", "clusterId", clusterId)
	}
	return nil
}

func (impl *InformerImpl) StopAll() {
	for _, clusterId := range impl.getStoppableClusterIds() {
		if err := impl.StopInformerForCluster(clusterId); err != nil {
			impl.logger.Errorw("error in stopping argo workflow ci informer", "clusterId", clusterId, "err", err)
		}
	}
}
