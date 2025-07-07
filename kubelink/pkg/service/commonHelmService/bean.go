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

package commonHelmService

import (
	"errors"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/common-lib/workerPool"
	"github.com/devtron-labs/kubelink/bean"
	"github.com/devtron-labs/kubelink/pkg/asyncProvider"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

type HelmReleaseStatusConfig struct {
	InstallAppVersionHistoryId int
	Message                    string
	IsReleaseInstalled         bool
	ErrorInInstallation        bool
}

type BuildNodesConfig struct {
	DesiredOrLiveManifests []*bean.DesiredOrLiveManifest
	batchWorker            *workerPool.WorkerPool[*BuildNodeResponse]
	BuildNodesRequest
}

type GetNodeFromManifestRequest struct {
	DesiredOrLiveManifest *bean.DesiredOrLiveManifest
	BuildNodesRequest
}

type BuildNodesRequest struct {
	RestConfig        *rest.Config
	ReleaseNamespace  string
	ParentResourceRef *commonBean.ResourceRef
}

func NewBuildNodesRequest(buildNodesConfig *BuildNodesRequest) *BuildNodesConfig {
	if buildNodesConfig == nil {
		return &BuildNodesConfig{}
	}
	req := &BuildNodesConfig{
		BuildNodesRequest: *buildNodesConfig,
	}
	return req
}

func NewGetNodesFromManifest(buildNodesConfig *BuildNodesRequest) *GetNodeFromManifestRequest {
	if buildNodesConfig == nil {
		return &GetNodeFromManifestRequest{}
	}
	req := &GetNodeFromManifestRequest{
		BuildNodesRequest: *buildNodesConfig,
	}
	return req
}

func (req *BuildNodesConfig) WithDesiredOrLiveManifests(desiredOrLiveManifests ...*bean.DesiredOrLiveManifest) *BuildNodesConfig {
	if len(desiredOrLiveManifests) == 0 {
		return req
	}
	req.DesiredOrLiveManifests = append(req.DesiredOrLiveManifests, desiredOrLiveManifests...)
	return req
}

func (req *BuildNodesConfig) WithBatchWorker(buildNodesBatchSize int, logger *zap.SugaredLogger) *BuildNodesConfig {
	if buildNodesBatchSize <= 0 {
		buildNodesBatchSize = 1
	}
	// for parallel processing of Nodes
	req.batchWorker = asyncProvider.NewBatchWorker[*BuildNodeResponse](buildNodesBatchSize, logger)
	return req
}

func (req *GetNodeFromManifestRequest) WithDesiredOrLiveManifest(desiredOrLiveManifest *bean.DesiredOrLiveManifest) *GetNodeFromManifestRequest {
	if desiredOrLiveManifest == nil {
		return req
	}
	req.DesiredOrLiveManifest = desiredOrLiveManifest
	return req
}

func NewBuildNodesConfig(restConfig *rest.Config) *BuildNodesRequest {
	return &BuildNodesRequest{
		RestConfig: restConfig,
	}
}

func (req *BuildNodesRequest) WithReleaseNamespace(releaseNamespace string) *BuildNodesRequest {
	if releaseNamespace == "" {
		return req
	}
	req.ReleaseNamespace = releaseNamespace
	return req
}

func (req *BuildNodesRequest) WithParentResourceRef(parentResourceRef *commonBean.ResourceRef) *BuildNodesRequest {
	if parentResourceRef == nil {
		return req
	}
	req.ParentResourceRef = parentResourceRef
	return req
}

type BuildNodeResponse struct {
	Nodes             []*commonBean.ResourceNode
	HealthStatusArray []*commonBean.HealthStatus
}

type GetNodeFromManifestResponse struct {
	Node                           *commonBean.ResourceNode
	HealthStatus                   *commonBean.HealthStatus
	ResourceRef                    *commonBean.ResourceRef
	DesiredOrLiveChildrenManifests []*bean.DesiredOrLiveManifest
}

func NewGetNodesFromManifestResponse() *GetNodeFromManifestResponse {
	return &GetNodeFromManifestResponse{}
}

func (resp *GetNodeFromManifestResponse) WithNode(node *commonBean.ResourceNode) *GetNodeFromManifestResponse {
	if node == nil {
		return resp
	}
	resp.Node = node
	return resp
}

func (resp *GetNodeFromManifestResponse) WithHealthStatus(healthStatus *commonBean.HealthStatus) *GetNodeFromManifestResponse {
	if healthStatus == nil {
		return resp
	}
	resp.HealthStatus = healthStatus
	return resp
}

func (resp *GetNodeFromManifestResponse) WithParentResourceRef(resourceRef *commonBean.ResourceRef) *GetNodeFromManifestResponse {
	if resourceRef == nil {
		return resp
	}
	resp.ResourceRef = resourceRef
	return resp
}

func (resp *GetNodeFromManifestResponse) WithDesiredOrLiveManifests(desiredOrLiveManifests ...*bean.DesiredOrLiveManifest) *GetNodeFromManifestResponse {
	if len(desiredOrLiveManifests) == 0 {
		return resp
	}
	resp.DesiredOrLiveChildrenManifests = append(resp.DesiredOrLiveChildrenManifests, desiredOrLiveManifests...)
	return resp
}

func NewBuildNodeResponse() *BuildNodeResponse {
	return &BuildNodeResponse{}
}

func (resp *BuildNodeResponse) WithNodes(nodes []*commonBean.ResourceNode) *BuildNodeResponse {
	if len(nodes) == 0 {
		return resp
	}
	resp.Nodes = append(resp.Nodes, nodes...)
	return resp
}

func (resp *BuildNodeResponse) WithHealthStatusArray(healthStatusArray []*commonBean.HealthStatus) *BuildNodeResponse {
	if len(healthStatusArray) == 0 {
		return resp
	}
	resp.HealthStatusArray = append(resp.HealthStatusArray, healthStatusArray...)
	return resp
}

var (
	ErrorReleaseNotFoundOnCluster = errors.New("release not found")
)
