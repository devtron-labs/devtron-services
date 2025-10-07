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
	"github.com/caarlos0/env"
	"github.com/devtron-labs/common-lib/securestore"
	"github.com/devtron-labs/image-scanner/pkg/middleware"
	"net/http"
	"os"
	"time"

	"github.com/devtron-labs/common-lib/middlewares"
	client "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/image-scanner/api"
	"github.com/devtron-labs/image-scanner/pubsub"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type App struct {
	Router            *api.Router
	Logger            *zap.SugaredLogger
	server            *http.Server
	db                *pg.DB
	natsSubscription  *pubsub.NatSubscriptionImpl
	pubSubClient      *client.PubSubClientServiceImpl
	serverConfig      *ServerConfig
	encryptionService securestore.EncryptionKeyService
}

func NewApp(Router *api.Router, Logger *zap.SugaredLogger,
	db *pg.DB, natsSubscription *pubsub.NatSubscriptionImpl,
	pubSubClient *client.PubSubClientServiceImpl,
	encryptionService securestore.EncryptionKeyService) *App {
	serverConfig, err := GetServerConfig()
	if err != nil {
		Logger.Errorw("error in getting server config", "err", err)
	}
	return &App{
		Router:            Router,
		Logger:            Logger,
		db:                db,
		natsSubscription:  natsSubscription,
		pubSubClient:      pubSubClient,
		serverConfig:      serverConfig,
		encryptionService: encryptionService,
	}
}

type ServerConfig struct {
	SERVER_HTTP_PORT      int           `env:"SERVER_HTTP_PORT" envDefault:"8080"`
	ServerShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" envDefault:"5m" envDescription:"server shutdown timeout , default is 5 minutes"`
}

func GetServerConfig() (*ServerConfig, error) {
	serverConfig := ServerConfig{}
	err := env.Parse(&serverConfig)
	if err != nil {
		return &serverConfig, err
	}
	return &serverConfig, nil
}

func (app *App) Start() {

	httpPort := app.serverConfig.SERVER_HTTP_PORT
	app.Logger.Infow("starting server on ", "httpPort", httpPort)
	app.Router.Init()
	server := &http.Server{Addr: fmt.Sprintf(":%d", httpPort), Handler: app.Router.Router}
	app.Router.Router.Use(middleware.PrometheusMiddleware)
	app.Router.Router.Use(middlewares.Recovery)
	app.server = server
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		app.Logger.Errorw("error in startup", "err", err)
		os.Exit(2)
	}
}

func (app *App) Stop() {
	app.Logger.Infow("image scanner shutdown initiating")
	app.Logger.Infow("server shutdown timeout", "timeout", app.serverConfig.ServerShutdownTimeout)
	timeoutContext, cancel := context.WithTimeout(context.Background(), app.serverConfig.ServerShutdownTimeout)
	defer cancel()
	app.Logger.Infow("closing router")
	err := app.server.Shutdown(timeoutContext)
	if err != nil {
		app.Logger.Errorw("error in mux router shutdown", "err", err)
	}
	app.Logger.Infow("closing db connection")
	err = app.db.Close()
	if err != nil {
		app.Logger.Errorw("Error while closing DB", "error", err)
	}

	app.Logger.Infow("housekeeping done. exiting now")
}
