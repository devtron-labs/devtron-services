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

import "github.com/caarlos0/env"

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
