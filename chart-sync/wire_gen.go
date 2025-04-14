// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/devtron-labs/chart-sync/internals"
	"github.com/devtron-labs/chart-sync/internals/logger"
	"github.com/devtron-labs/chart-sync/internals/sql"
	"github.com/devtron-labs/chart-sync/pkg"
	"github.com/devtron-labs/common-lib/helmLib/registry"
)

// Injectors from wire.go:

func InitializeApp() (*App, error) {
	sugaredLogger := logger.NewSugardLogger()
	config, err := sql.GetConfig()
	if err != nil {
		return nil, err
	}
	db, err := sql.NewDbConnection(config, sugaredLogger)
	if err != nil {
		return nil, err
	}
	chartRepoRepositoryImpl := sql.NewChartRepoRepositoryImpl(db)
	helmRepoManagerImpl := pkg.NewHelmRepoManagerImpl(sugaredLogger)
	dockerArtifactStoreRepositoryImpl := sql.NewDockerArtifactStoreRepositoryImpl(db)
	ociRegistryConfigRepositoryImpl := sql.NewOCIRegistryConfigRepositoryImpl(db)
	appStoreRepositoryImpl := sql.NewAppStoreRepositoryImpl(sugaredLogger, db)
	appStoreApplicationVersionRepositoryImpl := sql.NewAppStoreApplicationVersionRepositoryImpl(sugaredLogger, db)
	configuration, err := internals.ParseConfiguration()
	if err != nil {
		return nil, err
	}
	defaultSettingsGetterImpl := registry.NewDefaultSettingsGetter(sugaredLogger)
	settingsFactoryImpl := registry.NewSettingsFactoryImpl(defaultSettingsGetterImpl)
	syncServiceImpl := pkg.NewSyncServiceImpl(chartRepoRepositoryImpl, sugaredLogger, helmRepoManagerImpl, dockerArtifactStoreRepositoryImpl, ociRegistryConfigRepositoryImpl, appStoreRepositoryImpl, appStoreApplicationVersionRepositoryImpl, configuration, settingsFactoryImpl)
	app := NewApp(sugaredLogger, db, syncServiceImpl, configuration)
	return app, nil
}
