// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/image-scanner/api"
	"github.com/devtron-labs/image-scanner/internal/logger"
	"github.com/devtron-labs/image-scanner/internal/sql"
	"github.com/devtron-labs/image-scanner/internal/sql/repository"
	"github.com/devtron-labs/image-scanner/pkg/clairService"
	"github.com/devtron-labs/image-scanner/pkg/grafeasService"
	"github.com/devtron-labs/image-scanner/pkg/klarService"
	"github.com/devtron-labs/image-scanner/pkg/security"
	"github.com/devtron-labs/image-scanner/pkg/user"
	"github.com/devtron-labs/image-scanner/pprof"
	"github.com/devtron-labs/image-scanner/pubsub"
)

// Injectors from Wire.go:

func InitializeApp() (*App, error) {
	sugaredLogger := logger.NewSugardLogger()
	pubSubClientServiceImpl := pubsub_lib.NewPubSubClientServiceImpl(sugaredLogger)
	testPublishImpl := pubsub.NewTestPublishImpl(pubSubClientServiceImpl, sugaredLogger)
	apiClient := grafeasService.GetGrafeasClient()
	client := logger.NewHttpClient()
	grafeasServiceImpl := grafeasService.NewKlarServiceImpl(sugaredLogger, apiClient, client)
	config, err := sql.GetConfig()
	if err != nil {
		return nil, err
	}
	db, err := sql.NewDbConnection(config, sugaredLogger)
	if err != nil {
		return nil, err
	}
	userRepositoryImpl := repository.NewUserRepositoryImpl(db)
	userServiceImpl := user.NewUserServiceImpl(sugaredLogger, userRepositoryImpl)
	imageScanHistoryRepositoryImpl := repository.NewImageScanHistoryRepositoryImpl(db, sugaredLogger)
	imageScanResultRepositoryImpl := repository.NewImageScanResultRepositoryImpl(db, sugaredLogger)
	imageScanObjectMetaRepositoryImpl := repository.NewImageScanObjectMetaRepositoryImpl(db, sugaredLogger)
	cveStoreRepositoryImpl := repository.NewCveStoreRepositoryImpl(db, sugaredLogger)
	imageScanDeployInfoRepositoryImpl := repository.NewImageScanDeployInfoRepositoryImpl(db, sugaredLogger)
	ciArtifactRepositoryImpl := repository.NewCiArtifactRepositoryImpl(db, sugaredLogger)
	scanToolExecutionHistoryMappingRepositoryImpl := repository.NewScanToolExecutionHistoryMappingRepositoryImpl(db, sugaredLogger)
	scanToolMetadataRepositoryImpl := repository.NewScanToolMetadataRepositoryImpl(db, sugaredLogger)
	scanStepConditionRepositoryImpl := repository.NewScanStepConditionRepositoryImpl(db, sugaredLogger)
	scanToolStepRepositoryImpl := repository.NewScanToolStepRepositoryImpl(db, sugaredLogger)
	scanStepConditionMappingRepositoryImpl := repository.NewScanStepConditionMappingRepositoryImpl(db, sugaredLogger)
	imageScanConfig, err := security.GetImageScannerConfig()
	if err != nil {
		return nil, err
	}
	dockerArtifactStoreRepositoryImpl := repository.NewDockerArtifactStoreRepositoryImpl(db, sugaredLogger)
	registryIndexMappingRepositoryImpl := repository.NewRegistryIndexMappingRepositoryImpl(db, sugaredLogger)
	imageScanServiceImpl := security.NewImageScanServiceImpl(sugaredLogger, imageScanHistoryRepositoryImpl, imageScanResultRepositoryImpl, imageScanObjectMetaRepositoryImpl, cveStoreRepositoryImpl, imageScanDeployInfoRepositoryImpl, ciArtifactRepositoryImpl, scanToolExecutionHistoryMappingRepositoryImpl, scanToolMetadataRepositoryImpl, scanStepConditionRepositoryImpl, scanToolStepRepositoryImpl, scanStepConditionMappingRepositoryImpl, imageScanConfig, dockerArtifactStoreRepositoryImpl, registryIndexMappingRepositoryImpl)
	klarConfig, err := klarService.GetKlarConfig()
	if err != nil {
		return nil, err
	}
	klarServiceImpl := klarService.NewKlarServiceImpl(sugaredLogger, klarConfig, grafeasServiceImpl, userRepositoryImpl, imageScanServiceImpl, dockerArtifactStoreRepositoryImpl, scanToolMetadataRepositoryImpl)
	clairConfig, err := clairService.GetClairConfig()
	if err != nil {
		return nil, err
	}
	clairServiceImpl := clairService.NewClairServiceImpl(sugaredLogger, clairConfig, client, imageScanServiceImpl, dockerArtifactStoreRepositoryImpl, scanToolMetadataRepositoryImpl)
	restHandlerImpl := api.NewRestHandlerImpl(sugaredLogger, testPublishImpl, grafeasServiceImpl, userServiceImpl, imageScanServiceImpl, klarServiceImpl, clairServiceImpl, imageScanConfig)
	pProfRestHandlerImpl := pprof.NewPProfRestHandler(sugaredLogger)
	pProfRouterImpl := pprof.NewPProfRouter(sugaredLogger, pProfRestHandlerImpl)
	router := api.NewRouter(sugaredLogger, restHandlerImpl, pProfRouterImpl)
	natSubscriptionImpl, err := pubsub.NewNatSubscription(pubSubClientServiceImpl, sugaredLogger, clairServiceImpl)
	if err != nil {
		return nil, err
	}
	app := NewApp(router, sugaredLogger, db, natSubscriptionImpl, pubSubClientServiceImpl)
	return app, nil
}
