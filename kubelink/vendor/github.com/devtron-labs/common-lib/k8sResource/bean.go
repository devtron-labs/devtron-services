package k8sResource

import (
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Identifier struct {
	gvk       schema.GroupVersionKind
	name      string
	namespace string
}

func NewIdentifier(name, namespace string, gvk schema.GroupVersionKind) *Identifier {
	return &Identifier{
		gvk:       gvk,
		name:      name,
		namespace: namespace,
	}
}

func (identifier *Identifier) GetGvk() schema.GroupVersionKind {
	if identifier == nil {
		return schema.GroupVersionKind{}
	}
	return identifier.gvk
}

func (identifier *Identifier) GetName() string {
	if identifier == nil {
		return ""
	}
	return identifier.name
}

func (identifier *Identifier) GetNamespace() string {
	if identifier == nil {
		return ""
	}
	return identifier.namespace
}

type PatchRequest struct {
	// TODO: Use Identifier instead of Name, Namespace, Gvk
	Name      string
	Namespace string
	Gvk       *schema.GroupVersionKind
	Patch     string
	PatchType string
}

type ParentChildGvkMapping struct {
	Group        string         `json:"group"`
	Version      string         `json:"version"`
	Kind         string         `json:"kind"`
	ChildObjects []ChildObjects `json:"childObjects"`
}

type ChildObjects struct {
	Group    string             `json:"group"`
	Version  string             `json:"version"`
	Resource string             `json:"resource"`
	Scope    meta.RESTScopeName `json:"scope"`
}

func (r ChildObjects) GetGvrAndScopeForChildObject() *k8sCommonBean.GvrAndScope {
	return &k8sCommonBean.GvrAndScope{
		Gvr: schema.GroupVersionResource{
			Group:    r.Group,
			Version:  r.Version,
			Resource: r.Resource,
		},
		Scope: r.Scope,
	}
}

func (r ParentChildGvkMapping) GetParentGvk() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   r.Group,
		Version: r.Version,
		Kind:    r.Kind,
	}
}

type FilterChildrenObjectsResponse struct {
	pvcs      []unstructured.Unstructured
	manifests []*unstructured.Unstructured
}

func NewFilterChildrenObjectsResponse() *FilterChildrenObjectsResponse {
	return &FilterChildrenObjectsResponse{}
}

func (resp *FilterChildrenObjectsResponse) GetPvcs() []unstructured.Unstructured {
	return resp.pvcs
}

func (resp *FilterChildrenObjectsResponse) GetManifests() []*unstructured.Unstructured {
	return resp.manifests
}

func (resp *FilterChildrenObjectsResponse) WithPVCs(pvcs []unstructured.Unstructured) *FilterChildrenObjectsResponse {
	resp.pvcs = append(resp.pvcs, pvcs...)
	return resp
}

func (resp *FilterChildrenObjectsResponse) WithManifest(manifest *unstructured.Unstructured) *FilterChildrenObjectsResponse {
	if manifest == nil {
		return resp
	}
	resp.manifests = append(resp.manifests, manifest)
	return resp
}

type FilterChildrenObjectsRequest struct {
	childGvk    schema.GroupVersionResource
	pvcs        []unstructured.Unstructured
	listObjects *unstructured.UnstructuredList
	// TODO: Use Identifier instead of Name, Namespace, Gvk
	namespace  string
	parentGvk  schema.GroupVersionKind
	parentName string
}

func (req *FilterChildrenObjectsRequest) IsChildResourceTypePVC() bool {
	return req.GetParentGvk().Kind == k8sCommonBean.StatefulSetKind && req.GetChildGvk().Resource == k8sCommonBean.PersistentVolumeClaimsResourceType
}

func (req *FilterChildrenObjectsRequest) GetLoggerMetadata(keysAndValues ...any) []any {
	metaData := []any{
		"namespace", req.namespace,
		"childGvk", req.childGvk,
		"parentGvk", req.parentGvk,
		"parentName", req.parentName,
	}
	return append(metaData, keysAndValues...)
}

func (req *FilterChildrenObjectsRequest) GetChildGvk() schema.GroupVersionResource {
	return req.childGvk
}

func (req *FilterChildrenObjectsRequest) GetPvcs() []unstructured.Unstructured {
	return req.pvcs
}

func (req *FilterChildrenObjectsRequest) GetListObjects() *unstructured.UnstructuredList {
	return req.listObjects
}

func (req *FilterChildrenObjectsRequest) GetNamespace() string {
	return req.namespace
}

func (req *FilterChildrenObjectsRequest) GetParentGvk() schema.GroupVersionKind {
	return req.parentGvk
}

func (req *FilterChildrenObjectsRequest) GetParentName() string {
	return req.parentName
}

func NewFilterChildrenObjectsRequest() *FilterChildrenObjectsRequest {
	return &FilterChildrenObjectsRequest{}
}

func (req *FilterChildrenObjectsRequest) WithChildGvk(gvr schema.GroupVersionResource) *FilterChildrenObjectsRequest {
	req.childGvk = gvr
	return req
}

func (req *FilterChildrenObjectsRequest) WithPvcs(pvcs []unstructured.Unstructured) *FilterChildrenObjectsRequest {
	req.pvcs = pvcs
	return req
}

func (req *FilterChildrenObjectsRequest) WithListObjects(objects *unstructured.UnstructuredList) *FilterChildrenObjectsRequest {
	req.listObjects = objects
	return req
}

func (req *FilterChildrenObjectsRequest) WithNamespace(namespace string) *FilterChildrenObjectsRequest {
	req.namespace = namespace
	return req
}

func (req *FilterChildrenObjectsRequest) WithParentGvk(parentGvk schema.GroupVersionKind) *FilterChildrenObjectsRequest {
	req.parentGvk = parentGvk
	return req
}

func (req *FilterChildrenObjectsRequest) WithParentName(parentName string) *FilterChildrenObjectsRequest {
	req.parentName = parentName
	return req
}
