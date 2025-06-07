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

package k8sResource

import (
	"context"
	"encoding/json"
	"errors"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/meta"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	dynamicClient "k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type ClusterConfig struct {
	Host        string
	BearerToken string
}

type K8sService interface {
	CanHaveChild(gvk schema.GroupVersionKind) bool
	GetLiveManifest(restConfig *rest.Config, namespace string, gvk *schema.GroupVersionKind, name string) (*unstructured.Unstructured, *schema.GroupVersionResource, error)

	// GetChildObjectsV1 is the old implementation for getting children resource manifests for a parent GVK.
	// It doesn't support paginated listing,
	// which can cause high memory consumption if the child GVK are present in a large number.
	// But as it fetches all the data in a single call, it saves up multiple round trips cost.
	// This is the deprecated way to get child resources.
	// Use GetChildObjectsV2 instead.
	GetChildObjectsV1(restConfig *rest.Config, namespace string, parentGvk schema.GroupVersionKind, parentName string, parentApiVersion string) ([]*unstructured.Unstructured, error)

	// GetChildObjectsV2 is the new implementation for getting children resource manifests for a parent GVK.
	// It supports paginated listing,
	// which will optimize memory consumption if the child GVK are present in a large number.
	// But as it fetches all the data in multiple calls, it will cost multiple round trips.
	// This is the recommended way to get child resources.
	GetChildObjectsV2(restConfig *rest.Config, parentIdentifier *Identifier) ([]*unstructured.Unstructured, error)
	GetChildObjectsV2WithFilter(restConfig *rest.Config, parentIdentifier *Identifier, filterOpt ChildObjectFilterOpt) ([]*unstructured.Unstructured, error)
	PatchResource(ctx context.Context, restConfig *rest.Config, r *PatchRequest) error
}

type K8sServiceImpl struct {
	logger                       *zap.SugaredLogger
	k8sResourceConfig            *ServiceConfig
	gvkVsChildGvrAndScope        map[schema.GroupVersionKind][]*k8sCommonBean.GvrAndScope
	defaultGvkVsChildGvrAndScope map[schema.GroupVersionKind][]*k8sCommonBean.GvrAndScope
}

func NewK8sServiceImpl(
	logger *zap.SugaredLogger,
	k8sResourceConfig *ServiceConfig,
	defaultGvkVsChildGvrAndScope map[schema.GroupVersionKind][]*k8sCommonBean.GvrAndScope,
) (*K8sServiceImpl, error) {
	gvkVsChildGvrAndScope := make(map[schema.GroupVersionKind][]*k8sCommonBean.GvrAndScope)
	k8sServiceImpl := &K8sServiceImpl{
		logger:                       logger,
		k8sResourceConfig:            k8sResourceConfig,
		gvkVsChildGvrAndScope:        gvkVsChildGvrAndScope,
		defaultGvkVsChildGvrAndScope: defaultGvkVsChildGvrAndScope,
	}
	if len(k8sResourceConfig.ParentChildGvkMapping) > 0 {
		k8sServiceImpl.logger.Infow("caching parent gvk to child gvr and scope mapping")
		_, err := k8sServiceImpl.cacheParentChildGvkMapping(gvkVsChildGvrAndScope)
		if err != nil {
			k8sServiceImpl.logger.Errorw("error in caching parent gvk to child gvr and scope mapping", "err", err)
			return nil, err
		}
	}
	return k8sServiceImpl, nil
}

func (impl *K8sServiceImpl) cacheParentChildGvkMapping(gvkVsChildGvrAndScope map[schema.GroupVersionKind][]*k8sCommonBean.GvrAndScope) (map[schema.GroupVersionKind][]*k8sCommonBean.GvrAndScope, error) {
	var gvkChildMappings []ParentChildGvkMapping
	parentChildGvkMapping := impl.k8sResourceConfig.ParentChildGvkMapping
	err := json.Unmarshal([]byte(parentChildGvkMapping), &gvkChildMappings)
	if err != nil {
		impl.logger.Errorw("error in unmarshalling ParentChildGvkMapping", "parentChildGvkMapping", parentChildGvkMapping, "err", err)
		return gvkVsChildGvrAndScope, err
	}
	for _, parent := range gvkChildMappings {
		childGvrAndScopes := make([]*k8sCommonBean.GvrAndScope, len(parent.ChildObjects))
		for i, childObj := range parent.ChildObjects {
			childGvrAndScopes[i] = childObj.GetGvrAndScopeForChildObject()
		}
		gvkVsChildGvrAndScope[parent.GetParentGvk()] = childGvrAndScopes
	}
	return gvkVsChildGvrAndScope, nil
}

func (impl *K8sServiceImpl) getChildGvrFromParentGvk(parentGvk schema.GroupVersionKind) ([]*k8sCommonBean.GvrAndScope, bool) {
	var gvrAndScopes []*k8sCommonBean.GvrAndScope
	var ok bool
	//if parent child gvk mapping found from CM override it over local hardcoded gvk mapping
	if len(impl.k8sResourceConfig.ParentChildGvkMapping) > 0 && len(impl.gvkVsChildGvrAndScope) > 0 {
		gvrAndScopes, ok = impl.gvkVsChildGvrAndScope[parentGvk]
	} else {
		gvrAndScopes, ok = impl.defaultGvkVsChildGvrAndScope[parentGvk]
	}
	return gvrAndScopes, ok
}

func (impl *K8sServiceImpl) CanHaveChild(gvk schema.GroupVersionKind) bool {
	_, ok := impl.getChildGvrFromParentGvk(gvk)
	return ok
}

func (impl *K8sServiceImpl) GetLiveManifest(restConfig *rest.Config, namespace string, gvk *schema.GroupVersionKind, name string) (*unstructured.Unstructured, *schema.GroupVersionResource, error) {
	impl.logger.Debugw("Getting live manifest ", "namespace", namespace, "gvk", gvk, "name", name)

	gvr, scope, err := impl.getGvrAndScopeFromGvk(gvk, restConfig)
	if err != nil {
		return nil, nil, err
	}

	dynamicClient, err := dynamicClient.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
	}
	if scope.Name() != meta.RESTScopeNameNamespace {
		manifest, err := dynamicClient.Resource(*gvr).Get(context.Background(), name, metaV1.GetOptions{})
		return manifest, gvr, err
	} else {
		manifest, err := dynamicClient.Resource(*gvr).Namespace(namespace).Get(context.Background(), name, metaV1.GetOptions{})
		return manifest, gvr, err
	}
}

func (impl *K8sServiceImpl) PatchResource(ctx context.Context, restConfig *rest.Config, r *PatchRequest) error {
	impl.logger.Debugw("Patching resource ", "namespace", r.Namespace, "name", r.Name)

	client, err := dynamicClient.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	gvr, scope, err := impl.getGvrAndScopeFromGvk(r.Gvk, restConfig)
	if err != nil {
		return err
	}

	if scope.Name() != meta.RESTScopeNameNamespace {
		_, err = client.Resource(*gvr).Patch(ctx, r.Name, types.PatchType(r.PatchType), []byte(r.Patch), metaV1.PatchOptions{})
	} else {
		_, err = client.Resource(*gvr).Namespace(r.Namespace).Patch(ctx, r.Name, types.PatchType(r.PatchType), []byte(r.Patch), metaV1.PatchOptions{})
	}

	if err != nil {
		return err
	}

	return nil
}

func (impl *K8sServiceImpl) getGvrAndScopeFromGvk(gvk *schema.GroupVersionKind, restConfig *rest.Config) (*schema.GroupVersionResource, meta.RESTScope, error) {
	descoClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(descoClient))
	restMapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, nil, err
	}
	if restMapping == nil {
		return nil, nil, errors.New("gvr not found for given gvk")
	}
	return &restMapping.Resource, restMapping.Scope, nil
}
