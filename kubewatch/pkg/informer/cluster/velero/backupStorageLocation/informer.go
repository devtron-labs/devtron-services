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

package veleroBslInformer

import (
	"github.com/devtron-labs/common-lib/async"
	repository "github.com/devtron-labs/kubewatch/pkg/cluster"
	"github.com/devtron-labs/kubewatch/pkg/config"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	"github.com/devtron-labs/kubewatch/pkg/resource"
	resourceBean "github.com/devtron-labs/kubewatch/pkg/resource/bean"
	"github.com/devtron-labs/kubewatch/pkg/utils"
	"go.uber.org/zap"
	"time"
)

type InformerImpl struct {
	logger                   *zap.SugaredLogger
	k8sUtil                  utils.K8sUtil
	appConfig                *config.AppConfig
	informerClient           resource.InformerClient
	asyncRunnable            *async.Runnable
	veleroBslInformerStopper map[int]*informerBean.SharedStopper
}

func NewInformerImpl(logger *zap.SugaredLogger,
	appConfig *config.AppConfig,
	k8sUtil utils.K8sUtil,
	informerClient resource.InformerClient,
	asyncRunnable *async.Runnable) *InformerImpl {
	return &InformerImpl{
		logger:                   logger,
		appConfig:                appConfig,
		k8sUtil:                  k8sUtil,
		informerClient:           informerClient,
		asyncRunnable:            asyncRunnable,
		veleroBslInformerStopper: make(map[int]*informerBean.SharedStopper),
	}
}

func (impl *InformerImpl) StartInformerForCluster(clusterInfo *repository.Cluster) error {
	if impl.appConfig.GetExternalConfig().External {
		impl.logger.Warnw("velero informer is not enabled for external mode, skipping...", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName, "appConfig", impl.appConfig)
		return nil
	}
	startTime := time.Now()
	defer func() {
		impl.logger.Infow("time taken to start velero informer", "clusterId", clusterInfo.Id, "time", time.Since(startTime))
	}()
	clusterLabels := informerBean.NewClusterLabels(clusterInfo.ClusterName, clusterInfo.Id)
	stopChannel, err := impl.checkAndGetStopChannel(clusterLabels)
	if err != nil {
		impl.logger.Errorw("error in getting stop channel, velero informer already exists ", "clusterId", clusterInfo.Id, "err", err)
		return err
	}
	impl.logger.Infow("starting velero informer for cluster", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
	restConfig := impl.k8sUtil.GetK8sConfigForCluster(clusterInfo)
	bslInformerClient := impl.informerClient.GetSharedInformerClient(resourceBean.VeleroBslResourceType)
	bslInformer, err := bslInformerClient.GetSharedInformer(clusterLabels, impl.appConfig.GetVeleroNamespace(), restConfig)
	if err != nil {
		impl.logger.Errorw("error in registering velero bsl informer", "err", err, "clusterId", clusterInfo.Id)
		return err
	}
	runnable := func() {
		bslInformer.Run(stopChannel)
		impl.logger.Infow("informer started for velero bsl", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
	}
	impl.asyncRunnable.Execute(runnable)
	return nil
}

func (impl *InformerImpl) StopInformerForCluster(clusterId int) error {
	stopper, found := impl.getVeleroBslStopper(clusterId)
	if found {
		stopper.Stop()
		delete(impl.veleroBslInformerStopper, clusterId)
		impl.logger.Infow("velero bsl informer stopped for cluster", "clusterId", clusterId)
	}
	return nil
}

func (impl *InformerImpl) StopAll() {
	for _, stopper := range impl.veleroBslInformerStopper {
		stopper.Stop()
	}
}
