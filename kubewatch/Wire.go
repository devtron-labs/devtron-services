//go:build wireinject
// +build wireinject

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

package main

import (
	"github.com/devtron-labs/common-lib/monitoring"
	k8s "github.com/devtron-labs/common-lib/utils/k8s"
	api "github.com/devtron-labs/kubewatch/api/router"
	repository "github.com/devtron-labs/kubewatch/pkg/cluster"
	"github.com/devtron-labs/kubewatch/pkg/config"
	"github.com/devtron-labs/kubewatch/pkg/controller/inCluster"
	"github.com/devtron-labs/kubewatch/pkg/controller/multiCluster"
	"github.com/devtron-labs/kubewatch/pkg/logger"
	"github.com/devtron-labs/kubewatch/pkg/pubsub"
	"github.com/devtron-labs/kubewatch/pkg/resource"
	"github.com/devtron-labs/kubewatch/pkg/sql"
	"github.com/devtron-labs/kubewatch/pkg/utils"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		logger.NewSugaredLogger,
		config.GetAppConfig,
		sql.GetConfig,
		utils.GetDefaultK8sConfig,

		sql.NewDbConnection,

		repository.WireSet,

		k8s.NewCustomK8sHttpTransportConfig,
		utils.NewK8sUtilImpl,
		wire.Bind(new(utils.K8sUtil), new(*utils.K8sUtilImpl)),

		pubsub.NewPubSubClientServiceImpl,

		resource.NewInformerClientImpl,
		wire.Bind(new(resource.InformerClient), new(*resource.InformerClientImpl)),

		multiCluster.NewMultiClusterInformerImpl,
		wire.Bind(new(multiCluster.Informer), new(*multiCluster.InformerImpl)),

		inCluster.NewStartController,
		wire.Bind(new(inCluster.Informer), new(*inCluster.InformerImpl)),

		NewApp,
		api.NewRouter,
		monitoring.NewMonitoringRouter,
	)
	return &App{}, nil
}
