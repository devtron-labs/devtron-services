// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/devtron-labs/common-lib/helmLib/registry"
	"github.com/devtron-labs/common-lib/k8sResource"
	"github.com/devtron-labs/common-lib/monitoring"
	"github.com/devtron-labs/common-lib/utils/grpc"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/kubelink/api/router"
	"github.com/devtron-labs/kubelink/config"
	"github.com/devtron-labs/kubelink/converter"
	"github.com/devtron-labs/kubelink/internals/lock"
	"github.com/devtron-labs/kubelink/internals/logger"
	"github.com/devtron-labs/kubelink/pkg/asyncProvider"
	"github.com/devtron-labs/kubelink/pkg/cluster"
	"github.com/devtron-labs/kubelink/pkg/k8sInformer"
	"github.com/devtron-labs/kubelink/pkg/service"
	"github.com/devtron-labs/kubelink/pkg/service/commonHelmService"
	"github.com/devtron-labs/kubelink/pkg/service/fluxService"
	"github.com/devtron-labs/kubelink/pkg/service/helmApplicationService"
	"github.com/devtron-labs/kubelink/pkg/sql"
)

// Injectors from Wire.go:

func InitializeApp() (*App, error) {
	sugaredLogger := logger.NewSugaredLogger()
	chartRepositoryLocker := lock.NewChartRepositoryLocker(sugaredLogger)
	serviceConfig, err := k8sResource.GetK8sResourceConfig()
	if err != nil {
		return nil, err
	}
	v := commonBean.GetGvkVsChildGvrAndScope()
	k8sServiceImpl, err := k8sResource.NewK8sServiceImpl(sugaredLogger, serviceConfig, v)
	if err != nil {
		return nil, err
	}
	sqlConfig, err := sql.GetConfig()
	if err != nil {
		return nil, err
	}
	db, err := sql.NewDbConnection(sqlConfig, sugaredLogger)
	if err != nil {
		return nil, err
	}
	clusterRepositoryImpl := repository.NewClusterRepositoryImpl(db, sugaredLogger)
	helmReleaseConfig, err := config.GetHelmReleaseConfig()
	if err != nil {
		return nil, err
	}
	runtimeConfig, err := k8s.GetRuntimeConfig()
	if err != nil {
		return nil, err
	}
	k8sK8sServiceImpl, err := k8s.NewK8sUtil(sugaredLogger, runtimeConfig)
	if err != nil {
		return nil, err
	}
	clusterBeanConverterImpl := converter.NewConverterImpl()
	runnable := asyncProvider.NewAsyncRunnable(sugaredLogger)
	k8sInformerImpl, err := k8sInformer.Newk8sInformerImpl(sugaredLogger, clusterRepositoryImpl, helmReleaseConfig, k8sK8sServiceImpl, clusterBeanConverterImpl, runnable)
	if err != nil {
		return nil, err
	}
	resourceTreeServiceImpl := commonHelmService.NewResourceTreeServiceImpl(k8sServiceImpl, sugaredLogger, helmReleaseConfig)
	commonHelmServiceImpl := commonHelmService.NewCommonHelmServiceImpl(sugaredLogger, k8sK8sServiceImpl, clusterBeanConverterImpl, k8sServiceImpl, helmReleaseConfig, resourceTreeServiceImpl)
	defaultSettingsGetterImpl := registry.NewDefaultSettingsGetter(sugaredLogger)
	settingsFactoryImpl := registry.NewSettingsFactoryImpl(defaultSettingsGetterImpl)
	helmAppServiceImpl, err := helmApplicationService.NewHelmAppServiceImpl(sugaredLogger, k8sServiceImpl, k8sInformerImpl, helmReleaseConfig, k8sK8sServiceImpl, clusterBeanConverterImpl, clusterRepositoryImpl, commonHelmServiceImpl, settingsFactoryImpl, resourceTreeServiceImpl)
	if err != nil {
		return nil, err
	}
	fluxApplicationServiceImpl := fluxService.NewFluxApplicationServiceImpl(sugaredLogger, clusterRepositoryImpl, k8sK8sServiceImpl, clusterBeanConverterImpl, commonHelmServiceImpl)
	applicationServiceServerImpl := service.NewApplicationServiceServerImpl(sugaredLogger, chartRepositoryLocker, helmAppServiceImpl, fluxApplicationServiceImpl)
	monitoringRouter := monitoring.NewMonitoringRouter(sugaredLogger)
	routerImpl := router.NewRouter(sugaredLogger, monitoringRouter)
	configuration, err := grpc.GetConfiguration()
	if err != nil {
		return nil, err
	}
	app := NewApp(sugaredLogger, applicationServiceServerImpl, routerImpl, k8sInformerImpl, db, configuration)
	return app, nil
}
