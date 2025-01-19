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

package inCluster

import (
	"github.com/devtron-labs/kubewatch/pkg/config"
	"github.com/devtron-labs/kubewatch/pkg/middleware"
	"github.com/devtron-labs/kubewatch/pkg/resource"
	"github.com/devtron-labs/kubewatch/pkg/resource/bean"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

type Informer interface {
	Start(stopChan <-chan int)
}
type InformerImpl struct {
	logger           *zap.SugaredLogger
	appConfig        *config.AppConfig
	informerFactory  resource.InformerClient
	defaultK8sConfig *rest.Config
}

func NewStartController(logger *zap.SugaredLogger,
	appConfig *config.AppConfig,
	informerFactory resource.InformerClient,
	defaultK8sConfig *rest.Config) *InformerImpl {
	return &InformerImpl{
		logger:           logger,
		appConfig:        appConfig,
		informerFactory:  informerFactory,
		defaultK8sConfig: defaultK8sConfig,
	}
}

func (impl *InformerImpl) Start(stopChan <-chan int) {
	var namespace string
	if impl.appConfig.GetCiConfig().CiInformer {
		if impl.appConfig.GetExternalConfig().External {
			namespace = impl.appConfig.GetExternalConfig().Namespace
		} else {
			namespace = impl.appConfig.GetCiConfig().DefaultNamespace
		}
		ciWfInformer := impl.informerFactory.GetSharedInformerClient(bean.CiWorkflowResourceType)
		workflowInformer, err := ciWfInformer.GetSharedInformer(namespace, impl.defaultK8sConfig)
		if err != nil {
			impl.logger.Errorw("error in starting workflow informer", "err", err)
			middleware.IncUnregisteredInformers("Default cluster", "1", "CIStageArgoWorkflow")
		}
		stopCh := make(chan struct{})
		defer close(stopCh)
		go workflowInformer.Run(stopCh)
	}

	///-------------------

	if impl.appConfig.GetCdConfig().CdInformer {
		if impl.appConfig.GetExternalConfig().External {
			namespace = impl.appConfig.GetExternalConfig().Namespace
		} else {
			namespace = impl.appConfig.GetCdConfig().DefaultNamespace
		}
		cdWfInformer := impl.informerFactory.GetSharedInformerClient(bean.CdWorkflowResourceType)
		workflowInformer, err := cdWfInformer.GetSharedInformer(namespace, impl.defaultK8sConfig)
		if err != nil {
			impl.logger.Errorw("error in starting workflow informer", "err", err)
			middleware.IncUnregisteredInformers("Default cluster", "1", "CDStageArgoWorkflow")
		}
		stopCh := make(chan struct{})
		defer close(stopCh)
		go workflowInformer.Run(stopCh)
	}

	if impl.appConfig.GetAcdConfig().ACDInformer && !impl.appConfig.GetExternalConfig().External && !impl.appConfig.IsMultiClusterArgoCD() {
		impl.logger.Infow("starting acd informer", "namespace", impl.appConfig.GetAcdConfig().ACDNamespace)
		applicationInformer := impl.informerFactory.GetSharedInformerClient(bean.ApplicationResourceType)
		acdInformer, err := applicationInformer.GetSharedInformer(impl.appConfig.GetAcdConfig().ACDNamespace, impl.defaultK8sConfig)
		if err != nil {
			impl.logger.Errorw("error in registering acd informer", "err", err)
			middleware.IncUnregisteredInformers("Default cluster", "1", "ArgoCD")
		}
		appStopCh := make(chan struct{})
		defer close(appStopCh)
		go acdInformer.Run(appStopCh)
	}
	<-stopChan
}
