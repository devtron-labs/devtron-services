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

package veleroRestoreInformer

import (
	"github.com/devtron-labs/common-lib/async"
	"github.com/devtron-labs/kubewatch/pkg/config"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	"github.com/devtron-labs/kubewatch/pkg/resource"
	"github.com/devtron-labs/kubewatch/pkg/utils"
	"go.uber.org/zap"
)

type InformerImpl struct {
	logger                       *zap.SugaredLogger
	k8sUtil                      utils.K8sUtil
	appConfig                    *config.AppConfig
	informerClient               resource.InformerClient
	asyncRunnable                *async.Runnable
	veleroRestoreInformerStopper map[int]*informerBean.SharedStopper
}

func NewInformerImpl(logger *zap.SugaredLogger,
	k8sUtil utils.K8sUtil,
	appConfig *config.AppConfig,
	informerClient resource.InformerClient,
	asyncRunnable *async.Runnable) *InformerImpl {
	return &InformerImpl{
		logger:                       logger,
		k8sUtil:                      k8sUtil,
		appConfig:                    appConfig,
		informerClient:               informerClient,
		asyncRunnable:                asyncRunnable,
		veleroRestoreInformerStopper: make(map[int]*informerBean.SharedStopper),
	}
}
