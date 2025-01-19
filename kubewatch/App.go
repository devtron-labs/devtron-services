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
	"context"
	"errors"
	"fmt"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/utils/k8s"
	api "github.com/devtron-labs/kubewatch/api/router"
	"github.com/devtron-labs/kubewatch/pkg/config"
	"github.com/devtron-labs/kubewatch/pkg/controller/inCluster"
	"github.com/devtron-labs/kubewatch/pkg/controller/multiCluster"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"
)

type App struct {
	muxRouter            *api.RouterImpl
	logger               *zap.SugaredLogger
	server               *http.Server
	db                   *pg.DB
	defaultTimeout       time.Duration
	appConfig            *config.AppConfig
	client               *pubsub.PubSubClientServiceImpl
	httpTransportConfig  *k8s.CustomK8sHttpTransportConfig
	multiClusterInformer multiCluster.Informer
	inClusterInformer    inCluster.Informer
}

func NewApp(muxRouter *api.RouterImpl,
	logger *zap.SugaredLogger,
	appConfig *config.AppConfig,
	db *pg.DB,
	multiClusterInformer multiCluster.Informer,
	informer inCluster.Informer) *App {
	return &App{
		muxRouter:            muxRouter,
		logger:               logger,
		appConfig:            appConfig,
		db:                   db,
		defaultTimeout:       time.Duration(appConfig.GetTimeout().SleepTimeout) * time.Second,
		multiClusterInformer: multiClusterInformer,
		inClusterInformer:    informer,
	}
}

func (app *App) Start() {
	port := 8080 //TODO: extract from environment variable
	app.logger.Infow("starting server on ", "port", port)
	app.muxRouter.Init()

	app.server = &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: app.muxRouter.Router}
	err := app.server.ListenAndServe()
	// checking for ErrServerClosed if graceful shutdown is triggered
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		app.logger.Errorw("error in startup", "err", err)
		os.Exit(2)
	}
}

func (app *App) Stop() {
	app.logger.Infow("kubewatch shutdown initiating")
	timeoutContext, cancel := context.WithTimeout(context.Background(), app.defaultTimeout)
	defer cancel()
	app.logger.Infow("closing router")
	err := app.server.Shutdown(timeoutContext)
	if err != nil {
		app.logger.Errorw("error in mux router shutdown", "err", err)
	}
	app.logger.Infow("router closed successfully")

	if app.appConfig.IsMultiClusterMode() {
		app.multiClusterInformer.Stop()
		app.logger.Infow("closing db connection")
		err = app.db.Close()
		if err != nil {
			app.logger.Errorw("error while closing DB", "error", err)
		}
		app.logger.Infow("db closed successfully")
	}
	app.logger.Infow("kubewatch closed successfully")
}
