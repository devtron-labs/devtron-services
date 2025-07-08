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

// CATEGORY=HELM_RELEASE
type HelmReleaseConfig struct {
	EnableHelmReleaseCache                      bool   `env:"ENABLE_HELM_RELEASE_CACHE" envDefault:"true" description:"Enable helm releases list cache" deprecated:"false" example:"true"`
	MaxCountForHelmRelease                      int    `env:"MAX_COUNT_FOR_HELM_RELEASE" envDefault:"20" description:"Max count for helm release history list" deprecated:"false" example:"20"`
	ManifestFetchBatchSize                      int    `env:"MANIFEST_FETCH_BATCH_SIZE" envDefault:"2" description:"Manifest fetch parallelism batch size (applied only for parent objects)" deprecated:"false" example:"2"`
	RunHelmInstallInAsyncMode                   bool   `env:"RUN_HELM_INSTALL_IN_ASYNC_MODE" envDefault:"false" description:"Run helm install/ upgrade in async mode" deprecated:"false" example:"false"`
	ChartWorkingDirectory                       string `env:"CHART_WORKING_DIRECTORY" envDefault:"/home/devtron/devtroncd/charts/" description:"Helm charts working directory" deprecated:"false" example:"/home/devtron/devtroncd/charts/"`
	BuildNodesBatchSize                         int    `env:"BUILD_NODES_BATCH_SIZE" envDefault:"2" description:"Resource tree build nodes parallelism batch size (applied only for depth-1 child objects of a parent object)" deprecated:"false" example:"2"`
	FeatChildChildObjectListingPaginationEnable bool   `env:"FEAT_CHILD_OBJECT_LISTING_PAGINATION" envDefault:"true" description:"use pagination in listing all the dependent child objects. use 'CHILD_OBJECT_LISTING_PAGE_SIZE' to set the page size." deprecated:"false" example:"true"`
}

func GetHelmReleaseConfig() (*HelmReleaseConfig, error) {
	cfg := &HelmReleaseConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func (c *HelmReleaseConfig) IsHelmReleaseCachingEnabled() bool {
	return c.EnableHelmReleaseCache
}
