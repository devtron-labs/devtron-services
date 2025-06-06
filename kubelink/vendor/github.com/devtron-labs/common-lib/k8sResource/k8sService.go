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
	k8sUtils "github.com/devtron-labs/common-lib/utils/k8s"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"go.uber.org/zap"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	runtimeResource "k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	dynamicClient "k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"time"
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
	GetChildObjectsV2(restConfig *rest.Config, namespace string, parentGvk schema.GroupVersionKind, parentName string) ([]*unstructured.Unstructured, error)
	PatchResource(ctx context.Context, restConfig *rest.Config, r *PatchRequest) error
}

type K8sServiceImpl struct {
	logger                *zap.SugaredLogger
	k8sResourceConfig     *ServiceConfig
	gvkVsChildGvrAndScope map[schema.GroupVersionKind][]*k8sCommonBean.GvrAndScope
}

func NewK8sServiceImpl(
	logger *zap.SugaredLogger,
	k8sResourceConfig *ServiceConfig,
) (*K8sServiceImpl, error) {
	gvkVsChildGvrAndScope := make(map[schema.GroupVersionKind][]*k8sCommonBean.GvrAndScope)
	k8sServiceImpl := &K8sServiceImpl{
		logger:                logger,
		k8sResourceConfig:     k8sResourceConfig,
		gvkVsChildGvrAndScope: gvkVsChildGvrAndScope,
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

func (impl *K8sServiceImpl) GetChildGvrFromParentGvk(parentGvk schema.GroupVersionKind) ([]*k8sCommonBean.GvrAndScope, bool) {
	var gvrAndScopes []*k8sCommonBean.GvrAndScope
	var ok bool
	//if parent child gvk mapping found from CM override it over local hardcoded gvk mapping
	if len(impl.k8sResourceConfig.ParentChildGvkMapping) > 0 && len(impl.gvkVsChildGvrAndScope) > 0 {
		gvrAndScopes, ok = impl.gvkVsChildGvrAndScope[parentGvk]
	} else {
		gvrAndScopes, ok = k8sCommonBean.GetGvkVsChildGvrAndScope()[parentGvk]
	}
	return gvrAndScopes, ok
}

func (impl *K8sServiceImpl) CanHaveChild(gvk schema.GroupVersionKind) bool {
	_, ok := impl.GetChildGvrFromParentGvk(gvk)
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

func (impl *K8sServiceImpl) filterChildrenFromListObjects(request *FilterChildrenObjectsRequest) (*FilterChildrenObjectsResponse, error) {
	response := NewFilterChildrenObjectsResponse()
	if request.GetListObjects() == nil {
		impl.logger.Debugw("filter children objects is empty. skipping...", request.GetLoggerMetadata()...)
		return response, nil
	} else if request.IsChildResourceTypePVC() {
		impl.logger.Debugw("filter children objects is of type pvc. updating pvc list...", request.GetLoggerMetadata()...)
		response.WithPVCs(request.GetListObjects().Items)
		return response, nil
	} else {
		startTime := time.Now()
		for _, item := range request.GetListObjects().Items {
			// special handling for pvcs created via statefulsets
			ownerRefs, isInferredParentOf := k8sUtils.ResolveResourceReferences(&item)
			if request.GetChildGvk().Resource == k8sCommonBean.StatefulSetsResourceType && isInferredParentOf != nil {
				for _, pvc := range request.GetPvcs() {
					var pvcClaim coreV1.PersistentVolumeClaim
					err := runtime.DefaultUnstructuredConverter.FromUnstructured(pvc.Object, &pvcClaim)
					if err != nil {
						impl.logger.Errorw("error in converting unstructured to pvc", request.GetLoggerMetadata("timeTaken", time.Since(startTime).Seconds(), "err", err)...)
						return response, err
					}
					isCurrentStsParentOfPvc := isInferredParentOf(k8sUtils.ResourceKey{
						Group:     "",
						Kind:      pvcClaim.Kind,
						Namespace: request.GetNamespace(),
						Name:      pvcClaim.Name,
					})
					if isCurrentStsParentOfPvc && item.GetName() == request.GetParentName() {
						response = response.WithManifest(pvc.DeepCopy())
					}
				}
			}
			item.SetOwnerReferences(ownerRefs)
			for _, ownerRef := range item.GetOwnerReferences() {
				parentApiVersion, parentKind := request.GetParentGvk().ToAPIVersionAndKind()
				if ownerRef.Name == request.GetParentName() && ownerRef.APIVersion == parentApiVersion && ownerRef.Kind == parentKind {
					// using deep copy as it replaces item in manifest in loop
					response = response.WithManifest(item.DeepCopy())
				}
			}
		}
		impl.logger.Debugw("filtered children objects", request.GetLoggerMetadata("timeTaken", time.Since(startTime).Seconds())...)
		return response, nil
	}
}

func (impl *K8sServiceImpl) getK8sResourceClient(k8sResource dynamicClient.NamespaceableResourceInterface, scope meta.RESTScopeName, namespace string) dynamicClient.ResourceInterface {
	if scope != meta.RESTScopeNameNamespace {
		return k8sResource
	}
	return k8sResource.Namespace(namespace)
}

func (impl *K8sServiceImpl) getChildObject(client *dynamicClient.DynamicClient, pvcs []unstructured.Unstructured,
	gvrAndScope *k8sCommonBean.GvrAndScope, namespace string, parentGvk schema.GroupVersionKind, parentName string) ([]unstructured.Unstructured, []*unstructured.Unstructured, error) {
	startTime := time.Now()
	var manifests []*unstructured.Unstructured
	childGvk := gvrAndScope.Gvr
	childScope := gvrAndScope.Scope
	childResourceClient := impl.getK8sResourceClient(client.Resource(childGvk), childScope, namespace)
	listOptions := metaV1.ListOptions{
		Limit: impl.k8sResourceConfig.ChildObjectListingPageSize,
	}
	filterObjRequest := NewFilterChildrenObjectsRequest().
		WithChildGvk(childGvk).
		WithNamespace(namespace).
		WithParentGvk(parentGvk).
		WithParentName(parentName).
		WithPvcs(pvcs)
	counter := 1
	err := runtimeResource.FollowContinue(&listOptions,
		func(options metaV1.ListOptions) (runtime.Object, error) {
			filterListStartTime := time.Now()
			childrenObjectsList, k8sErr := childResourceClient.List(context.Background(), options)
			if k8sErr != nil {
				impl.logger.Errorw("error in getting child listObjects", filterObjRequest.GetLoggerMetadata("counter", counter, "timeTaken", time.Since(filterListStartTime).Seconds(), "err", k8sErr)...)
				return nil, k8sErr
			}
			impl.logger.Debugw("listing child objects", filterObjRequest.GetLoggerMetadata("counter", counter, "timeTaken", time.Since(filterListStartTime).Seconds())...)
			filterObjRequest = filterObjRequest.WithListObjects(childrenObjectsList)
			response, filterErr := impl.filterChildrenFromListObjects(filterObjRequest)
			if filterErr != nil {
				impl.logger.Errorw("error in filtering child listObjects", filterObjRequest.GetLoggerMetadata("counter", counter, "timeTaken", time.Since(filterListStartTime).Seconds(), "err", filterErr)...)
				return nil, filterErr
			}
			pvcs = response.GetPvcs()
			manifests = append(manifests, response.GetManifests()...)
			if childrenObjectsList == nil {
				return childrenObjectsList.NewEmptyInstance(), nil
			}
			return childrenObjectsList, nil
		})
	if err != nil {
		impl.logger.Errorw("error in getting child listObjects", filterObjRequest.GetLoggerMetadata("timeTaken", time.Since(startTime).Seconds(), "err", err)...)
		return pvcs, manifests, err
	}
	return pvcs, manifests, nil
}

func (impl *K8sServiceImpl) GetChildObjectsV1(restConfig *rest.Config, namespace string, parentGvk schema.GroupVersionKind, parentName string, parentApiVersion string) ([]*unstructured.Unstructured, error) {
	impl.logger.Debugw("Getting child objects ", "namespace", namespace, "parentGvk", parentGvk, "parentName", parentName, "parentApiVersion", parentApiVersion)

	gvrAndScopes, ok := impl.GetChildGvrFromParentGvk(parentGvk)
	if !ok {
		impl.logger.Errorw("gvr not found for given kind", "parentGvk", parentGvk)
		return nil, errors.New("grv not found for given kind")
	}
	client, err := dynamicClient.NewForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in creating dynamic client", "host", restConfig.Host, "namespace", namespace, "err", err)
		return nil, err
	}
	var pvcs []unstructured.Unstructured
	var manifests []*unstructured.Unstructured
	for _, gvrAndScope := range gvrAndScopes {
		gvr := gvrAndScope.Gvr
		scope := gvrAndScope.Scope

		var objects *unstructured.UnstructuredList
		if scope != meta.RESTScopeNameNamespace {
			objects, err = client.Resource(gvr).List(context.Background(), metaV1.ListOptions{})
		} else {
			objects, err = client.Resource(gvr).Namespace(namespace).List(context.Background(), metaV1.ListOptions{})
		}

		if err != nil {
			impl.logger.Errorw("error in getting child objects", "namespace", namespace, "gvr", gvr, "parentGvk", parentGvk, "err", err)
			return nil, err
		}

		if objects != nil {
			for _, item := range objects.Items {
				ownerRefs, isInferredParentOf := k8sUtils.ResolveResourceReferences(&item)
				if parentGvk.Kind == k8sCommonBean.StatefulSetKind && gvr.Resource == k8sCommonBean.PersistentVolumeClaimsResourceType {
					pvcs = append(pvcs, item)
					continue
				}
				// special handling for pvcs created via statefulsets
				if gvr.Resource == k8sCommonBean.StatefulSetsResourceType && isInferredParentOf != nil {
					for _, pvc := range pvcs {
						var pvcClaim coreV1.PersistentVolumeClaim
						err := runtime.DefaultUnstructuredConverter.FromUnstructured(pvc.Object, &pvcClaim)
						if err != nil {
							impl.logger.Errorw("error in converting unstructured to pvc", "namespace", namespace, "gvr", gvr, "err", err)
							return manifests, err
						}
						isCurrentStsParentOfPvc := isInferredParentOf(k8sUtils.ResourceKey{
							Group:     "",
							Kind:      pvcClaim.Kind,
							Namespace: namespace,
							Name:      pvcClaim.Name,
						})
						if isCurrentStsParentOfPvc && item.GetName() == parentName {
							manifests = append(manifests, pvc.DeepCopy())
						}
					}
				}
				item.SetOwnerReferences(ownerRefs)
				for _, ownerRef := range item.GetOwnerReferences() {
					if ownerRef.Name == parentName && ownerRef.Kind == parentGvk.Kind && ownerRef.APIVersion == parentApiVersion {
						// using deep copy as it replaces item in manifest in loop
						manifests = append(manifests, item.DeepCopy())
					}
				}
			}
		}

	}

	return manifests, nil
}

func (impl *K8sServiceImpl) GetChildObjectsV2(restConfig *rest.Config, namespace string, parentGvk schema.GroupVersionKind, parentName string) ([]*unstructured.Unstructured, error) {
	startTime := time.Now()
	impl.logger.Debugw("Getting child listObjects", "namespace", namespace, "parentGvk", parentGvk, "parentName", parentName, "startTime", startTime)
	gvrAndScopes, ok := impl.GetChildGvrFromParentGvk(parentGvk)
	if !ok {
		impl.logger.Errorw("gvr not found for given kind", "parentGvk", parentGvk, "timeTaken", time.Since(startTime).Seconds())
		return nil, errors.New("grv not found for given kind")
	}
	client, err := dynamicClient.NewForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in creating dynamic client", "host", restConfig.Host, "namespace", namespace, "timeTaken", time.Since(startTime).Seconds(), "err", err)
		return nil, err
	}
	var pvcs []unstructured.Unstructured
	var manifests []*unstructured.Unstructured
	for _, gvrAndScope := range gvrAndScopes {
		childrenPVCs, childObjManifests, err := impl.getChildObject(client, pvcs, gvrAndScope, namespace, parentGvk, parentName)
		if err != nil {
			impl.logger.Errorw("error in getting child listObjects", "namespace", namespace, "childGvk", gvrAndScope.Gvr, "parentGvk", parentGvk, "timeTaken", time.Since(startTime).Seconds(), "err", err)
			return manifests, err
		}
		pvcs = append(pvcs, childrenPVCs...)
		manifests = append(manifests, childObjManifests...)
	}
	return manifests, nil
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
