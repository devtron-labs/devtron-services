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

package resource

import (
	bean2 "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	"github.com/devtron-labs/kubewatch/pkg/resource/bean"
	"github.com/devtron-labs/kubewatch/pkg/resource/pod"
	"github.com/devtron-labs/kubewatch/pkg/resource/secret"
	coreV1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
)

type InformerFactory[T any] interface {
	GetSharedInformerFactory(config *rest.Config, clusterLabels *bean2.ClusterLabels,
		eventHandlers *bean.EventHandlers[T], options ...kubeinformers.SharedInformerOption) (kubeinformers.SharedInformerFactory, error)
}

func (impl *InformerClientImpl) GetConfigMapInformerFactory() InformerFactory[coreV1.ConfigMap] {
	return secret.NewInformerFactoryImpl(impl.logger, impl.k8sUtil)
}

func (impl *InformerClientImpl) GetPodInformerFactory() InformerFactory[coreV1.Pod] {
	return pod.NewInformerFactoryImpl(impl.logger, impl.k8sUtil)
}
