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
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/common-lib/workerPool"
	"github.com/devtron-labs/kubelink/bean"
	"github.com/devtron-labs/kubelink/pkg/asyncProvider"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type HelmReleaseStatusConfig struct {
	InstallAppVersionHistoryId int
	Message                    string
	IsReleaseInstalled         bool
	ErrorInInstallation        bool
}

type ParentChildGvkMapping struct {
	Group        string         `json:"group"`
	Version      string         `json:"version"`
	Kind         string         `json:"kind"`
	ChildObjects []ChildObjects `json:"childObjects"`
}

func (r ParentChildGvkMapping) GetParentGvk() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   r.Group,
		Version: r.Version,
		Kind:    r.Kind,
	}
}

type ChildObjects struct {
	Group    string             `json:"group"`
	Version  string             `json:"version"`
	Resource string             `json:"resource"`
	Scope    meta.RESTScopeName `json:"scope"`
}

func (r ChildObjects) GetGvrAndScopeForChildObject() *commonBean.GvrAndScope {
	return &commonBean.GvrAndScope{
		Gvr: schema.GroupVersionResource{
			Group:    r.Group,
			Version:  r.Version,
			Resource: r.Resource,
		},
		Scope: r.Scope,
	}
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

type filterChildrenObjectsResponse struct {
	pvcs      []unstructured.Unstructured
	manifests []*unstructured.Unstructured
}

func newFilterChildrenObjectsResponse() *filterChildrenObjectsResponse {
	return &filterChildrenObjectsResponse{}
}

func (resp *filterChildrenObjectsResponse) GetPvcs() []unstructured.Unstructured {
	return resp.pvcs
}

func (resp *filterChildrenObjectsResponse) GetManifests() []*unstructured.Unstructured {
	return resp.manifests
}

func (resp *filterChildrenObjectsResponse) WithPVCs(pvcs []unstructured.Unstructured) *filterChildrenObjectsResponse {
	resp.pvcs = append(resp.pvcs, pvcs...)
	return resp
}

func (resp *filterChildrenObjectsResponse) WithManifest(manifest *unstructured.Unstructured) *filterChildrenObjectsResponse {
	if manifest == nil {
		return resp
	}
	resp.manifests = append(resp.manifests, manifest)
	return resp
}

type filterChildrenObjectsRequest struct {
	childGvk    schema.GroupVersionResource
	pvcs        []unstructured.Unstructured
	listObjects *unstructured.UnstructuredList
	namespace   string
	parentGvk   schema.GroupVersionKind
	parentName  string
}

func (req *filterChildrenObjectsRequest) IsChildResourceTypePVC() bool {
	return req.GetParentGvk().Kind == k8sCommonBean.StatefulSetKind && req.GetChildGvk().Resource == k8sCommonBean.PersistentVolumeClaimsResourceType
}

func (req *filterChildrenObjectsRequest) GetLoggerMetadata(keysAndValues ...any) []any {
	metaData := []any{
		"namespace", req.namespace,
		"childGvk", req.childGvk,
		"parentGvk", req.parentGvk,
		"parentName", req.parentName,
	}
	return append(metaData, keysAndValues...)
}

func (req *filterChildrenObjectsRequest) GetChildGvk() schema.GroupVersionResource {
	return req.childGvk
}

func (req *filterChildrenObjectsRequest) GetPvcs() []unstructured.Unstructured {
	return req.pvcs
}

func (req *filterChildrenObjectsRequest) GetListObjects() *unstructured.UnstructuredList {
	return req.listObjects
}

func (req *filterChildrenObjectsRequest) GetNamespace() string {
	return req.namespace
}

func (req *filterChildrenObjectsRequest) GetParentGvk() schema.GroupVersionKind {
	return req.parentGvk
}

func (req *filterChildrenObjectsRequest) GetParentName() string {
	return req.parentName
}

func newFilterChildrenObjectsRequest() *filterChildrenObjectsRequest {
	return &filterChildrenObjectsRequest{}
}

func (req *filterChildrenObjectsRequest) WithChildGvk(gvr schema.GroupVersionResource) *filterChildrenObjectsRequest {
	req.childGvk = gvr
	return req
}

func (req *filterChildrenObjectsRequest) WithPvcs(pvcs []unstructured.Unstructured) *filterChildrenObjectsRequest {
	req.pvcs = pvcs
	return req
}

func (req *filterChildrenObjectsRequest) WithListObjects(objects *unstructured.UnstructuredList) *filterChildrenObjectsRequest {
	req.listObjects = objects
	return req
}

func (req *filterChildrenObjectsRequest) WithNamespace(namespace string) *filterChildrenObjectsRequest {
	req.namespace = namespace
	return req
}

func (req *filterChildrenObjectsRequest) WithParentGvk(parentGvk schema.GroupVersionKind) *filterChildrenObjectsRequest {
	req.parentGvk = parentGvk
	return req
}

func (req *filterChildrenObjectsRequest) WithParentName(parentName string) *filterChildrenObjectsRequest {
	req.parentName = parentName
	return req
}
