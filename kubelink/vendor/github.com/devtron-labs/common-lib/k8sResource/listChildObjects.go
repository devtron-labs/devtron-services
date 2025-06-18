package k8sResource

import (
	"context"
	"errors"
	k8sUtils "github.com/devtron-labs/common-lib/utils/k8s"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeResource "k8s.io/cli-runtime/pkg/resource"
	dynamicClient "k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"time"
)

func (impl *K8sServiceImpl) GetChildObjectsV1(restConfig *rest.Config, namespace string, parentGvk schema.GroupVersionKind, parentName string, parentApiVersion string) ([]*unstructured.Unstructured, error) {
	impl.logger.Debugw("Getting child objects ", "namespace", namespace, "parentGvk", parentGvk, "parentName", parentName, "parentApiVersion", parentApiVersion)

	gvrAndScopes, ok := impl.getChildGvrFromParentGvk(parentGvk)
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

func (impl *K8sServiceImpl) GetChildObjectsV2(restConfig *rest.Config, parentIdentifier *Identifier) ([]*unstructured.Unstructured, error) {
	parentGvk := parentIdentifier.GetGvk()
	parentName := parentIdentifier.GetName()
	namespace := parentIdentifier.GetNamespace()
	startTime := time.Now()
	impl.logger.Debugw("Getting child listObjects", "namespace", namespace, "parentGvk", parentGvk, "parentName", parentName, "startTime", startTime)
	gvrAndScopes, ok := impl.getChildGvrFromParentGvk(parentGvk)
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
