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

package config

import (
	"fmt"
	"github.com/caarlos0/env"
)

type AppConfig struct {
	externalConfig *ExternalConfig
	clusterCfg     *ClusterConfig
	ciConfig       *CiConfig
	cdConfig       *CdConfig
	acdConfig      *AcdConfig
	timeout        *Timeout
}

func (app *AppConfig) GetClusterConfig() *ClusterConfig {
	return app.clusterCfg
}

func (app *AppConfig) GetExternalConfig() *ExternalConfig {
	return app.externalConfig
}

func (app *AppConfig) GetCiConfig() *CiConfig {
	return app.ciConfig
}

func (app *AppConfig) GetCdConfig() *CdConfig {
	return app.cdConfig
}

func (app *AppConfig) GetAcdConfig() *AcdConfig {
	return app.acdConfig
}

func (app *AppConfig) GetTimeout() *Timeout {
	return app.timeout
}

func (app *AppConfig) IsMultiClusterMode() bool {
	return app.IsMultiClusterSystemExec() || app.IsMultiClusterArgoCD()
}

func (app *AppConfig) IsDBConnectionRequired() bool {
	return app.IsMultiClusterMode()
}

func (app *AppConfig) IsMultiClusterArgoCD() bool {
	return app.GetClusterConfig().ClusterArgoCDType == ClusterTypeAll && !app.GetExternalConfig().External
}

func (app *AppConfig) IsMultiClusterSystemExec() bool {
	return app.GetClusterConfig().SystemExecClusterType == ClusterTypeAll && !app.GetExternalConfig().External
}

func GetAppConfig() (*AppConfig, error) {
	clusterConfig, err := getClusterConfig()
	if err != nil {
		return nil, err
	}
	externalConfig, err := getExternalConfig()
	if err != nil {
		return nil, err
	}
	ciConfig, err := getCiConfig()
	if err != nil {
		return nil, err
	}
	cdConfig, err := getCdConfig()
	if err != nil {
		return nil, err
	}
	acdConfig, err := getAcdConfig()
	if err != nil {
		return nil, err
	}
	timeout, err := getTimeout()
	if err != nil {
		return nil, err
	}
	return &AppConfig{
		clusterCfg:     clusterConfig,
		externalConfig: externalConfig,
		ciConfig:       ciConfig,
		cdConfig:       cdConfig,
		acdConfig:      acdConfig,
		timeout:        timeout,
	}, nil
}

// ExternalConfig is being used by CI as well as CD
type ExternalConfig struct {
	External    bool   `env:"CD_EXTERNAL_REST_LISTENER" envDefault:"false"`
	Token       string `env:"CD_EXTERNAL_ORCHESTRATOR_TOKEN" envDefault:""`
	ListenerUrl string `env:"CD_EXTERNAL_LISTENER_URL" envDefault:"http://devtroncd-orchestrator-service-prod.devtroncd:80"`
	Namespace   string `env:"CD_EXTERNAL_NAMESPACE" envDefault:""`
}

func getExternalConfig() (*ExternalConfig, error) {
	externalConfig := &ExternalConfig{}
	err := env.Parse(externalConfig)
	if err != nil {
		return nil, err
	}
	return externalConfig, err
}

type ClusterConfig struct {
	SystemExecClusterType string `env:"CLUSTER_TYPE" envDefault:"IN_CLUSTER"`
	ClusterArgoCDType     string `env:"CLUSTER_ARGO_CD_TYPE" envDefault:"IN_CLUSTER"`
}

func getClusterConfig() (*ClusterConfig, error) {
	clusterCfg := &ClusterConfig{}
	err := env.Parse(clusterCfg)
	if err != nil {
		return nil, err
	}
	return clusterCfg, err
}

type CiConfig struct {
	DefaultNamespace string `env:"DEFAULT_NAMESPACE" envDefault:"devtron-ci"`
	CiInformer       bool   `env:"CI_INFORMER" envDefault:"true"`
}

func getCiConfig() (*CiConfig, error) {
	ciConfig := &CiConfig{}
	err := env.Parse(ciConfig)
	return ciConfig, err
}

type CdConfig struct {
	DefaultNamespace string `env:"CD_DEFAULT_NAMESPACE" envDefault:"devtron-cd"`
	CdInformer       bool   `env:"CD_INFORMER" envDefault:"true"`
}

func getCdConfig() (*CdConfig, error) {
	cdConfig := &CdConfig{}
	err := env.Parse(cdConfig)
	return cdConfig, err
}

type AcdConfig struct {
	ACDNamespace string `env:"ACD_NAMESPACE" envDefault:"devtroncd"`
	ACDInformer  bool   `env:"ACD_INFORMER" envDefault:"true"`
}

func getAcdConfig() (*AcdConfig, error) {
	acdCfg := &AcdConfig{}
	err := env.Parse(acdCfg)
	return acdCfg, err
}

type Timeout struct {
	SleepTimeout int `env:"SLEEP_TIMEOUT" envDefault:"5"`
}

func getTimeout() (*Timeout, error) {
	cfg := &Timeout{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println(err)
	}
	return cfg, err
}
