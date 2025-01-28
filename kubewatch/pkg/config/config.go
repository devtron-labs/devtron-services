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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (app *AppConfig) GetCiWfNamespace() string {
	if app.GetExternalConfig().External {
		return app.GetExternalConfig().Namespace
	} else {
		return app.GetCiConfig().DefaultNamespace
	}
}

func (app *AppConfig) GetCdWfNamespace() string {
	if app.GetExternalConfig().External {
		return app.GetExternalConfig().Namespace
	} else {
		return app.GetCdConfig().DefaultNamespace
	}
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

func (app *AppConfig) GetACDNamespace() string {
	if app.IsMultiClusterArgoCD() {
		return metav1.NamespaceAll
	} else {
		return app.acdConfig.ACDNamespace
	}
}

func (app *AppConfig) GetTimeout() *Timeout {
	return app.timeout
}

func (app *AppConfig) IsDBAvailable() bool {
	return !app.GetExternalConfig().External
}

func (app *AppConfig) IsMultiClusterArgoCD() bool {
	return app.GetClusterConfig().ClusterArgoCDType == AllClusterType && !app.GetExternalConfig().External
}

func (app *AppConfig) IsMultiClusterArgoWfType() bool {
	return app.GetClusterConfig().ClusterCdArgoWfType == AllClusterType && !app.GetExternalConfig().External
}

func (app *AppConfig) IsMultiClusterSystemExec() bool {
	return app.GetClusterConfig().SystemExecClusterType == AllClusterType && !app.GetExternalConfig().External
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

// ExternalConfig is used to determine whether it's an external kubewatch or internal kubewatch
// For external kubewatch, it will sit at an external namespace and publish events to the orchestrator using REST
// It is used by CI workflow as well as CD workflow
type ExternalConfig struct {
	// External is used to determine whether it's an external kubewatch or internal kubewatch
	External bool `env:"CD_EXTERNAL_REST_LISTENER" envDefault:"false"`

	// Token is the token used to authenticate with the orchestrator
	Token string `env:"CD_EXTERNAL_ORCHESTRATOR_TOKEN" envDefault:""`

	// ListenerUrl is the URL of the orchestrator
	ListenerUrl string `env:"CD_EXTERNAL_LISTENER_URL" envDefault:"http://devtroncd-orchestrator-service-prod.devtroncd:80"`

	// Namespace is the namespace where the external kubewatch is set up
	Namespace string `env:"CD_EXTERNAL_NAMESPACE" envDefault:""`
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
	// SystemExecClusterType uses CLUSTER_TYPE env variable; for backward compatibility
	//	- AllClusterType: All clusters are enabled for SystemExec informer
	//	- InClusterType: Only default cluster is enabled for SystemExec informer
	SystemExecClusterType string `env:"CLUSTER_TYPE" envDefault:"IN_CLUSTER"`

	// ClusterArgoCDType defines whether all clusters are enabled for ArgoCD informer
	//	- AllClusterType: All clusters are enabled for ArgoCD informer
	//	- InClusterType: Only default cluster is enabled for ArgoCD informer
	ClusterArgoCDType string `env:"CLUSTER_ARGO_CD_TYPE" envDefault:"IN_CLUSTER"`

	// ClusterCiArgoWfType defines whether all clusters are enabled for CI ArgoWorkflow informer
	//	- AllClusterType: All clusters are enabled for CI ArgoWorkflow informer
	//	- InClusterType: Only default cluster is enabled for CI ArgoWorkflow informer
	ClusterCiArgoWfType string `env:"CLUSTER_CI_ARGO_WF_TYPE" envDefault:"IN_CLUSTER"`

	// ClusterCdArgoWfType defines whether all clusters are enabled for CD ArgoWorkflow informer
	//	- AllClusterType: All clusters are enabled for CD ArgoWorkflow informer
	//	- InClusterType: Only default cluster is enabled for CD ArgoWorkflow informer
	ClusterCdArgoWfType string `env:"CLUSTER_CD_ARGO_WF_TYPE" envDefault:"IN_CLUSTER"`
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
	// DefaultNamespace is the namespace where all CI workflows are scheduled
	DefaultNamespace string `env:"DEFAULT_NAMESPACE" envDefault:"devtron-ci"`

	// CiInformer is used to determine whether CI informer is enabled or not
	CiInformer bool `env:"CI_INFORMER" envDefault:"true"`
}

func getCiConfig() (*CiConfig, error) {
	ciConfig := &CiConfig{}
	err := env.Parse(ciConfig)
	return ciConfig, err
}

type CdConfig struct {
	// DefaultNamespace is the namespace where all CD workflows are scheduled
	DefaultNamespace string `env:"CD_DEFAULT_NAMESPACE" envDefault:"devtron-cd"`

	// CdInformer is used to determine whether CD informer is enabled or not
	CdInformer bool `env:"CD_INFORMER" envDefault:"true"`
}

func getCdConfig() (*CdConfig, error) {
	cdConfig := &CdConfig{}
	err := env.Parse(cdConfig)
	return cdConfig, err
}

type AcdConfig struct {
	// ACDNamespace is the namespace where all the ArgoCD application objects are published
	// For multi-cluster mode, it will be set to v1.NamespaceAll
	ACDNamespace string `env:"ACD_NAMESPACE" envDefault:"devtroncd"`

	// ACDInformer is used to determine whether ArgoCD informer is enabled or not
	ACDInformer bool `env:"ACD_INFORMER" envDefault:"true"`
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
