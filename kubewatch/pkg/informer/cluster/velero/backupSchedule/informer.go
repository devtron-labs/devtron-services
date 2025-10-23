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

package veleroBackupScheduleInformer

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
	logger                              *zap.SugaredLogger
	k8sUtil                             utils.K8sUtil
	appConfig                           *config.AppConfig
	informerClient                      resource.InformerClient
	asyncRunnable                       *async.Runnable
	veleroBackupScheduleInformerStopper map[int]*informerBean.SharedStopper
}

func NewInformerImpl(logger *zap.SugaredLogger,
	k8sUtil utils.K8sUtil,
	appConfig *config.AppConfig,
	informerClient resource.InformerClient,
	asyncRunnable *async.Runnable) *InformerImpl {
	return &InformerImpl{
		logger:                              logger,
		k8sUtil:                             k8sUtil,
		appConfig:                           appConfig,
		informerClient:                      informerClient,
		asyncRunnable:                       asyncRunnable,
		veleroBackupScheduleInformerStopper: make(map[int]*informerBean.SharedStopper),
	}
}

func (impl *InformerImpl) StartInformerForCluster(clusterInfo *repository.Cluster) error {
	if impl.appConfig.GetExternalConfig().External {
		impl.logger.Warnw("external mode is enabled, skipping the informer start for velero backup schedule", "cluster", clusterInfo)
		return nil
	}
	startTime := time.Now()
	defer func() {
		impl.logger.Debugw("time taken to start informer for velero backup schedule", "time", time.Since(startTime))
	}()
	clusterLabels := informerBean.NewClusterLabels(clusterInfo.ClusterName, clusterInfo.Id)
	stopChannel, err := impl.checkAndGetStopChannel(clusterLabels)
	if err != nil {
		impl.logger.Errorw("error in getting stop channel, velero informer already exists ", "clusterId", clusterInfo.Id, "err", err)
		return err
	}
	impl.logger.Infow("starting velero informer for cluster", "clusterId", clusterInfo.Id)
	restConfig := impl.k8sUtil.GetK8sConfigForCluster(clusterInfo)
	backupScheduleInformerClient := impl.informerClient.GetSharedInformerClient(resourceBean.VeleroBackupScheduleResourceType)
	backupScheduleInformer, err := backupScheduleInformerClient.GetSharedInformer(clusterLabels, impl.appConfig.GetVeleroNamespace(), restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting velero backup schedule informer", "clusterId", clusterInfo.Id, "err", err)
		return err
	}
	runnable := func() {
		backupScheduleInformer.Run(stopChannel)
		impl.logger.Infow("informer started for velero backup schedule", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
	}
	impl.asyncRunnable.Execute(runnable)
	return nil
}

func (impl *InformerImpl) StopInformerForCluster(clusterId int) error {
	stopper, found := impl.getVeleroBackupScheduleStopper(clusterId)
	if found {
		stopper.Stop()
		delete(impl.veleroBackupScheduleInformerStopper, clusterId)
		impl.logger.Infow("velero bsl informer stopped for cluster", "clusterId", clusterId)
	}
	return nil
}

func (impl *InformerImpl) StopAll() {
	for _, stopper := range impl.veleroBackupScheduleInformerStopper {
		stopper.Stop()
	}
}
