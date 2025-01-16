package commonHelmService

import (
	"context"
	"errors"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/common-lib/utils/k8sObjectsUtil"
	"github.com/devtron-labs/common-lib/workerPool"
	"github.com/devtron-labs/kubelink/bean"
	globalConfig "github.com/devtron-labs/kubelink/config"
	client "github.com/devtron-labs/kubelink/grpc"
	"github.com/devtron-labs/kubelink/pkg/util"
	"go.uber.org/zap"
	"k8s.io/api/extensions/v1beta1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"net/http"
)

type ResourceTreeServiceImpl struct {
	k8sService        K8sService
	logger            *zap.SugaredLogger
	helmReleaseConfig *globalConfig.HelmReleaseConfig
}

type ResourceTreeService interface {
	BuildNodes(request *BuildNodesConfig) (*BuildNodeResponse, error)
	BuildResourceTreeUsingParentObjects(ctx context.Context, appDetailRequest *client.AppDetailRequest, conf *rest.Config, parentObjects []*client.ObjectIdentifier) (*bean.ResourceTreeResponse, error)
	BuildResourceTreeUsingK8s(ctx context.Context, appDetailRequest *client.AppDetailRequest, conf *rest.Config, parentObjects []*client.ObjectIdentifier) (*bean.ResourceTreeResponse, error)
}

func NewResourceTreeServiceImpl(k8sService K8sService,
	logger *zap.SugaredLogger,
	helmReleaseConfig *globalConfig.HelmReleaseConfig) *ResourceTreeServiceImpl {
	return &ResourceTreeServiceImpl{
		k8sService:        k8sService,
		logger:            logger,
		helmReleaseConfig: helmReleaseConfig,
	}
}
func (impl *ResourceTreeServiceImpl) BuildResourceTreeUsingParentObjects(ctx context.Context, appDetailRequest *client.AppDetailRequest, conf *rest.Config, parentObjects []*client.ObjectIdentifier) (*bean.ResourceTreeResponse, error) {
	parentObjects = sanitizeParentObjects(parentObjects)
	if appDetailRequest.PreferCache && appDetailRequest.CacheConfig != nil {
		impl.logger.Infow("Cache is not supported in oss", "releaseName", appDetailRequest.ReleaseName)
		if !appDetailRequest.UseFallBack {
			impl.logger.Infow("Use fallback is false, hence returning with error", "appDetailRequest", appDetailRequest)
			return nil, errors.New("Cache is not supported in oss and use_fallback flag is false")
		}

	}
	//fallback
	return impl.BuildResourceTreeUsingK8s(ctx, appDetailRequest, conf, parentObjects)

}

func sanitizeParentObjects(parentObjects []*client.ObjectIdentifier) []*client.ObjectIdentifier {
	sanitizedParentObjects := make([]*client.ObjectIdentifier, 0)
	if len(parentObjects) > 0 {
		for _, parentObject := range parentObjects {
			if parentObject != nil {
				sanitizedParentObjects = append(sanitizedParentObjects, parentObject)
			}
		}
	}
	return sanitizedParentObjects

}

func (impl *ResourceTreeServiceImpl) BuildResourceTreeUsingK8s(ctx context.Context, appDetailRequest *client.AppDetailRequest, conf *rest.Config, parentObjects []*client.ObjectIdentifier) (*bean.ResourceTreeResponse, error) {
	liveManifests := impl.getLiveManifestsForGVKList(conf, parentObjects)

	// build resource Nodes
	req := NewBuildNodesRequest(NewBuildNodesConfig(conf).
		WithReleaseNamespace(appDetailRequest.Namespace)).
		WithDesiredOrLiveManifests(liveManifests...).
		WithBatchWorker(impl.helmReleaseConfig.BuildNodesBatchSize, impl.logger)
	buildNodesResponse, err := impl.BuildNodes(req)
	if err != nil {
		return nil, err
	}
	updateHookInfoForChildNodes(buildNodesResponse.Nodes)

	// filter Nodes based on ResourceTreeFilter
	resourceTreeFilter := appDetailRequest.ResourceTreeFilter
	if resourceTreeFilter != nil && len(buildNodesResponse.Nodes) > 0 {
		buildNodesResponse.Nodes = impl.filterNodes(resourceTreeFilter, buildNodesResponse.Nodes)
	}

	// build pods metadata
	podsMetadata, err := impl.buildPodMetadata(buildNodesResponse.Nodes, conf)
	if err != nil {
		return nil, err
	}
	resourceTreeResponse := &bean.ResourceTreeResponse{
		ApplicationTree: &bean.ApplicationTree{
			Nodes: buildNodesResponse.Nodes,
		},
		PodMetadata: podsMetadata,
	}
	return resourceTreeResponse, nil
}

func (impl *ResourceTreeServiceImpl) getLiveManifestsForGVKList(restConfig *rest.Config, gvkList []*client.ObjectIdentifier) []*bean.DesiredOrLiveManifest {
	var manifests []*bean.DesiredOrLiveManifest
	for _, resource := range gvkList {
		gvk := &schema.GroupVersionKind{
			Group:   resource.GetGroup(),
			Version: resource.GetVersion(),
			Kind:    resource.GetKind(),
		}
		manifest, _, err := impl.k8sService.GetLiveManifest(restConfig, resource.GetNamespace(), gvk, resource.GetName())
		if err != nil {
			impl.logger.Errorw("Error in getting live manifest", "err", err)
			statusError, _ := err.(*errors2.StatusError)
			desiredManifest := &unstructured.Unstructured{}
			desiredManifest.SetGroupVersionKind(*gvk)
			desiredManifest.SetName(resource.Name)
			desiredManifest.SetNamespace(resource.Namespace)
			desiredManifest.SetAnnotations(resource.Annotations)
			desiredOrLiveManifest := &bean.DesiredOrLiveManifest{
				Manifest: desiredManifest,
				// using deep copy as it replaces item in manifest in loop
				IsLiveManifestFetchError: true,
			}
			if statusError != nil {
				desiredOrLiveManifest.LiveManifestFetchErrorCode = statusError.Status().Code
			}
			manifests = append(manifests, desiredOrLiveManifest)
		} else {
			manifests = append(manifests, &bean.DesiredOrLiveManifest{
				Manifest: manifest,
			})
		}
	}
	return manifests
}

// BuildNodes builds Nodes from desired or live manifest.
//   - It uses recursive approach to build child Nodes.
//   - It uses batch worker to build child Nodes in parallel.
//   - Batch workers configuration is provided in BuildNodesConfig.
//   - NOTE: To avoid creating batch worker recursively, it does not use batch worker for child Nodes.
func (impl *ResourceTreeServiceImpl) BuildNodes(request *BuildNodesConfig) (*BuildNodeResponse, error) {
	var buildChildNodesRequests []*BuildNodesConfig
	response := NewBuildNodeResponse()
	for _, desiredOrLiveManifest := range request.DesiredOrLiveManifests {

		// build request to get Nodes from desired or live manifest
		getNodesFromManifest := NewGetNodesFromManifest(NewBuildNodesConfig(request.RestConfig).
			WithParentResourceRef(request.ParentResourceRef).
			WithReleaseNamespace(request.ReleaseNamespace)).
			WithDesiredOrLiveManifest(desiredOrLiveManifest)

		// get Node from desired or live manifest
		getNodesFromManifestResponse, err := impl.getNodeFromDesiredOrLiveManifest(getNodesFromManifest)
		if err != nil {
			return response, err
		}
		// add Node and health status
		if getNodesFromManifestResponse.Node != nil {
			response.Nodes = append(response.Nodes, getNodesFromManifestResponse.Node)
			response.HealthStatusArray = append(response.HealthStatusArray, getNodesFromManifestResponse.Node.Health)
		}

		// add child Nodes request
		if len(getNodesFromManifestResponse.DesiredOrLiveChildrenManifests) > 0 {
			req := NewBuildNodesRequest(NewBuildNodesConfig(request.RestConfig).
				WithReleaseNamespace(request.ReleaseNamespace).
				WithParentResourceRef(getNodesFromManifestResponse.ResourceRef)).
				WithDesiredOrLiveManifests(getNodesFromManifestResponse.DesiredOrLiveChildrenManifests...)
			// NOTE:  Do not use batch worker for child Nodes as it will create batch worker recursively
			buildChildNodesRequests = append(buildChildNodesRequests, req)
		}
	}
	// build child Nodes, if any.
	// NOTE: build child Nodes calls buildNodes recursively
	childNodeResponse, err := impl.buildChildNodesInBatch(request.batchWorker, buildChildNodesRequests)
	if err != nil {
		return response, err
	}
	// add child Nodes and health status to response
	response.WithNodes(childNodeResponse.Nodes).WithHealthStatusArray(childNodeResponse.HealthStatusArray)
	return response, nil
}

// buildChildNodes builds child Nodes sequentially from desired or live manifest.
func (impl *ResourceTreeServiceImpl) buildChildNodes(buildChildNodesRequests []*BuildNodesConfig) (*BuildNodeResponse, error) {
	response := NewBuildNodeResponse()
	// for recursive calls, build child Nodes sequentially
	for _, req := range buildChildNodesRequests {
		// build child Nodes
		childNodesResponse, err := impl.BuildNodes(req)
		if err != nil {
			impl.logger.Errorw("error in building child Nodes", "ReleaseNamespace", req.ReleaseNamespace, "parentResource", req.ParentResourceRef.GetGvk(), "err", err)
			return response, err
		}
		response.WithNodes(childNodesResponse.Nodes).WithHealthStatusArray(childNodesResponse.HealthStatusArray)
	}
	return response, nil
}

// buildChildNodesInBatch builds child Nodes in parallel from desired or live manifest.
//   - It uses batch workers workerPool.WorkerPool[*BuildNodeResponse] to build child Nodes in parallel.
//   - If workerPool is not defined, it builds child Nodes sequentially.
func (impl *ResourceTreeServiceImpl) buildChildNodesInBatch(wp *workerPool.WorkerPool[*BuildNodeResponse], buildChildNodesRequests []*BuildNodesConfig) (*BuildNodeResponse, error) {
	if wp == nil {
		// build child Nodes sequentially
		return impl.buildChildNodes(buildChildNodesRequests)
	}
	response := NewBuildNodeResponse()
	for index := range buildChildNodesRequests {
		// passing buildChildNodesRequests[index] to closure as it will be updated in next iteration and the func call is async
		func(req *BuildNodesConfig) {
			// submit child Nodes build request to workerPool
			wp.Submit(func() (*BuildNodeResponse, error) {
				// build child Nodes
				return impl.BuildNodes(req)
			})
		}(buildChildNodesRequests[index])
	}
	// wait for all child Nodes build requests to complete and return error from workerPool error channel
	err := wp.StopWait()
	if err != nil {
		return response, err
	}
	// extract the children nodes from workerPool response
	for _, childNode := range wp.GetResponse() {
		response.WithNodes(childNode.Nodes).WithHealthStatusArray(childNode.HealthStatusArray)
	}
	return response, nil
}

func (impl *ResourceTreeServiceImpl) getNodeFromDesiredOrLiveManifest(request *GetNodeFromManifestRequest) (*GetNodeFromManifestResponse, error) {
	response := NewGetNodesFromManifestResponse()
	manifest := request.DesiredOrLiveManifest.Manifest
	gvk := manifest.GroupVersionKind()
	_namespace := manifest.GetNamespace()
	if _namespace == "" {
		_namespace = request.ReleaseNamespace
	}
	ports := k8sObjectsUtil.GetPorts(manifest, gvk)
	resourceRef := k8sObjectsUtil.BuildResourceRef(gvk, *manifest, _namespace)

	if impl.k8sService.CanHaveChild(gvk) {
		children, err := impl.k8sService.GetChildObjects(request.RestConfig, _namespace, gvk, manifest.GetName(), manifest.GetAPIVersion())
		if err != nil {
			return response, err
		}
		desiredOrLiveManifestsChildren := make([]*bean.DesiredOrLiveManifest, 0, len(children))
		for _, child := range children {
			desiredOrLiveManifestsChildren = append(desiredOrLiveManifestsChildren, &bean.DesiredOrLiveManifest{
				Manifest: child,
			})
		}
		response.WithParentResourceRef(resourceRef).
			WithDesiredOrLiveManifests(desiredOrLiveManifestsChildren...)
	}

	creationTimeStamp := ""
	val, found, err := unstructured.NestedString(manifest.Object, "metadata", "creationTimestamp")
	if found && err == nil {
		creationTimeStamp = val
	}
	node := &k8sCommonBean.ResourceNode{
		ResourceRef:     resourceRef,
		ResourceVersion: manifest.GetResourceVersion(),
		NetworkingInfo: &k8sCommonBean.ResourceNetworkingInfo{
			Labels: manifest.GetLabels(),
		},
		CreatedAt: creationTimeStamp,
		Port:      ports,
	}
	node.IsHook, node.HookType = k8sObjectsUtil.GetHookMetadata(manifest)

	if request.ParentResourceRef != nil {
		node.ParentRefs = append(make([]*k8sCommonBean.ResourceRef, 0), request.ParentResourceRef)
	}

	// set health of Node
	if request.DesiredOrLiveManifest.IsLiveManifestFetchError {
		if request.DesiredOrLiveManifest.LiveManifestFetchErrorCode == http.StatusNotFound {
			node.Health = &k8sCommonBean.HealthStatus{
				Status:  k8sCommonBean.HealthStatusMissing,
				Message: "Resource missing as live manifest not found",
			}
		} else {
			node.Health = &k8sCommonBean.HealthStatus{
				Status:  k8sCommonBean.HealthStatusUnknown,
				Message: "Resource state unknown as error while fetching live manifest",
			}
		}
	} else {
		k8sObjectsUtil.SetHealthStatusForNode(node, manifest, gvk)
	}

	// hibernate set starts
	if request.ParentResourceRef == nil {

		// set CanBeHibernated
		k8sObjectsUtil.SetHibernationRules(node, &node.Manifest)
	}
	// hibernate set ends

	if k8sObjectsUtil.IsPod(gvk.Kind, gvk.Group) {
		infoItems, _ := k8sObjectsUtil.PopulatePodInfo(manifest)
		node.Info = infoItems
	}
	k8sObjectsUtil.AddSelectiveInfoInResourceNode(node, gvk, manifest.Object)

	response.WithNode(node).WithHealthStatus(node.Health)
	return response, nil
}

func (impl *ResourceTreeServiceImpl) filterNodes(resourceTreeFilter *client.ResourceTreeFilter, nodes []*k8sCommonBean.ResourceNode) []*k8sCommonBean.ResourceNode {
	resourceFilters := resourceTreeFilter.ResourceFilters
	globalFilter := resourceTreeFilter.GlobalFilter
	if globalFilter == nil && (resourceFilters == nil || len(resourceFilters) == 0) {
		return nodes
	}

	filteredNodes := make([]*k8sCommonBean.ResourceNode, 0, len(nodes))

	// handle global
	if globalFilter != nil && len(globalFilter.Labels) > 0 {
		globalLabels := globalFilter.Labels
		for _, node := range nodes {
			toAdd := util.IsMapSubset(node.NetworkingInfo.Labels, globalLabels)
			if toAdd {
				filteredNodes = append(filteredNodes, node)
			}
		}
		return filteredNodes
	}

	// handle gvk level
	var gvkVsLabels map[schema.GroupVersionKind]map[string]string
	for _, resourceFilter := range resourceTreeFilter.ResourceFilters {
		gvk := resourceFilter.Gvk
		gvkVsLabels[schema.GroupVersionKind{
			Group:   gvk.Group,
			Version: gvk.Version,
			Kind:    gvk.Kind,
		}] = resourceFilter.ResourceIdentifier.Labels
	}

	for _, node := range nodes {
		nodeGvk := node.Manifest.GroupVersionKind()
		if val, ok := gvkVsLabels[nodeGvk]; ok {
			toAdd := util.IsMapSubset(node.NetworkingInfo.Labels, val)
			if toAdd {
				filteredNodes = append(filteredNodes, node)
			}
		}
	}

	return filteredNodes
}

func (impl *ResourceTreeServiceImpl) buildPodMetadata(nodes []*k8sCommonBean.ResourceNode, restConfig *rest.Config) ([]*k8sCommonBean.PodMetadata, error) {
	podMetadatas, err := k8sObjectsUtil.BuildPodMetadata(nodes)
	if err != nil {
		impl.logger.Errorw("error in building pod metadata", "err", err)
		return nil, err
	}
	if len(podMetadatas) > 0 {
		for _, node := range nodes {
			var isNew bool
			if len(node.ParentRefs) > 0 {
				deploymentPodHashMap, rolloutMap, uidVsExtraNodeInfoMap := k8sObjectsUtil.GetExtraNodeInfoMappings(nodes)
				isNew, err = impl.isPodNew(nodes, node, deploymentPodHashMap, rolloutMap, uidVsExtraNodeInfoMap, restConfig)
				if err != nil {
					return podMetadatas, err
				}
			}
			podMetadata := k8sObjectsUtil.GetMatchingPodMetadataForUID(podMetadatas, node.UID)
			if podMetadata != nil {
				podMetadata.IsNew = isNew
			}
		}
	}
	return podMetadatas, nil

}

func (impl *ResourceTreeServiceImpl) isPodNew(nodes []*k8sCommonBean.ResourceNode, node *k8sCommonBean.ResourceNode, deploymentPodHashMap map[string]string, rolloutMap map[string]*k8sCommonBean.ExtraNodeInfo,
	uidVsExtraNodeInfoMap map[string]*k8sCommonBean.ExtraNodeInfo, restConfig *rest.Config) (bool, error) {

	isNew := false
	parentRef := node.ParentRefs[0]
	parentKind := parentRef.Kind

	// if parent is StatefulSet - then pod label controller-revision-hash should match StatefulSet's update revision
	if parentKind == k8sCommonBean.StatefulSetKind && node.NetworkingInfo != nil {
		isNew = uidVsExtraNodeInfoMap[parentRef.UID].UpdateRevision == node.NetworkingInfo.Labels["controller-revision-hash"]
	}

	// if parent is Job - then pod label controller-revision-hash should match StatefulSet's update revision
	if parentKind == k8sCommonBean.JobKind {
		// TODO - new or old logic not built in orchestrator for Job's pods. hence not implementing here. as don't know the logic :)
		isNew = true
	}

	// if parent kind is replica set then
	if parentKind == k8sCommonBean.ReplicaSetKind {
		replicaSetNode := k8sObjectsUtil.GetMatchingNode(nodes, parentKind, parentRef.Name)

		// if parent of replicaset is deployment, compare label pod-template-hash
		if replicaSetParent := replicaSetNode.ParentRefs[0]; replicaSetNode != nil && len(replicaSetNode.ParentRefs) > 0 && replicaSetParent.Kind == k8sCommonBean.DeploymentKind {
			deploymentPodHash := deploymentPodHashMap[replicaSetParent.Name]
			replicaSetObj, err := impl.getReplicaSetObject(restConfig, replicaSetNode)
			if err != nil {
				return isNew, err
			}
			deploymentNode := k8sObjectsUtil.GetMatchingNode(nodes, replicaSetParent.Kind, replicaSetParent.Name)
			// TODO: why do we need deployment object for collisionCount ??
			var deploymentCollisionCount *int32
			if deploymentNode != nil && deploymentNode.DeploymentCollisionCount != nil {
				deploymentCollisionCount = deploymentNode.DeploymentCollisionCount
			} else {
				deploymentCollisionCount, err = impl.getDeploymentCollisionCount(restConfig, replicaSetParent)
				if err != nil {
					return isNew, err
				}
			}
			replicaSetPodHash := k8sObjectsUtil.GetReplicaSetPodHash(replicaSetObj, deploymentCollisionCount)
			isNew = replicaSetPodHash == deploymentPodHash
		} else if replicaSetParent.Kind == k8sCommonBean.K8sClusterResourceRolloutKind {

			rolloutExtraInfo := rolloutMap[replicaSetParent.Name]
			rolloutPodHash := rolloutExtraInfo.RolloutCurrentPodHash
			replicasetPodHash := k8sObjectsUtil.GetRolloutPodTemplateHash(replicaSetNode)

			isNew = rolloutPodHash == replicasetPodHash

		}

	}

	// if parent kind is DaemonSet then compare DaemonSet's Child ControllerRevision's label controller-revision-hash with pod label controller-revision-hash
	if parentKind == k8sCommonBean.DaemonSetKind {
		controllerRevisionNodes := k8sObjectsUtil.GetMatchingNodes(nodes, "ControllerRevision")
		for _, controllerRevisionNode := range controllerRevisionNodes {
			if len(controllerRevisionNode.ParentRefs) > 0 && controllerRevisionNode.ParentRefs[0].Kind == parentKind &&
				controllerRevisionNode.ParentRefs[0].Name == parentRef.Name && uidVsExtraNodeInfoMap[parentRef.UID].ResourceNetworkingInfo != nil &&
				node.NetworkingInfo != nil {

				isNew = uidVsExtraNodeInfoMap[parentRef.UID].ResourceNetworkingInfo.Labels["controller-revision-hash"] == node.NetworkingInfo.Labels["controller-revision-hash"]
			}
		}
	}
	return isNew, nil
}

func (impl *ResourceTreeServiceImpl) getReplicaSetObject(restConfig *rest.Config, replicaSetNode *k8sCommonBean.ResourceNode) (*v1beta1.ReplicaSet, error) {
	gvk := &schema.GroupVersionKind{
		Group:   replicaSetNode.Group,
		Version: replicaSetNode.Version,
		Kind:    replicaSetNode.Kind,
	}
	var replicaSetNodeObj map[string]interface{}
	var err error
	if replicaSetNode.Manifest.Object == nil {
		replicaSetNodeManifest, _, err := impl.k8sService.GetLiveManifest(restConfig, replicaSetNode.Namespace, gvk, replicaSetNode.Name)
		if err != nil {
			impl.logger.Errorw("error in getting replicaSet live manifest", "clusterName", restConfig.ServerName, "replicaSetName", replicaSetNode.Name)
			return nil, err
		}
		if replicaSetNodeManifest != nil {
			replicaSetNodeObj = replicaSetNodeManifest.Object
		}
	} else {
		replicaSetNodeObj = replicaSetNode.Manifest.Object
	}

	replicaSetObj, err := k8sObjectsUtil.ConvertToV1ReplicaSet(replicaSetNodeObj)
	if err != nil {
		impl.logger.Errorw("error in converting replicaSet unstructured object to replicaSet object", "clusterName", restConfig.ServerName, "replicaSetName", replicaSetNode.Name)
		return nil, err
	}
	return replicaSetObj, nil
}

func (impl *ResourceTreeServiceImpl) getDeploymentCollisionCount(restConfig *rest.Config, deploymentInfo *k8sCommonBean.ResourceRef) (*int32, error) {
	parentGvk := &schema.GroupVersionKind{
		Group:   deploymentInfo.Group,
		Version: deploymentInfo.Version,
		Kind:    deploymentInfo.Kind,
	}
	var deploymentNodeObj map[string]interface{}
	var err error
	if deploymentInfo.Manifest.Object == nil {
		deploymentLiveManifest, _, err := impl.k8sService.GetLiveManifest(restConfig, deploymentInfo.Namespace, parentGvk, deploymentInfo.Name)
		if err != nil {
			impl.logger.Errorw("error in getting parent deployment live manifest", "clusterName", restConfig.ServerName, "deploymentName", deploymentInfo.Name)
			return nil, err
		}
		if deploymentLiveManifest != nil {
			deploymentNodeObj = deploymentLiveManifest.Object
		}
	} else {
		deploymentNodeObj = deploymentInfo.Manifest.Object
	}

	deploymentObj, err := k8sObjectsUtil.ConvertToV1Deployment(deploymentNodeObj)
	if err != nil {
		impl.logger.Errorw("error in converting parent deployment unstructured object to replicaSet object", "clusterName", restConfig.ServerName, "deploymentName", deploymentInfo.Name)
		return nil, err
	}
	return deploymentObj.Status.CollisionCount, nil
}

func updateHookInfoForChildNodes(nodes []*k8sCommonBean.ResourceNode) {
	hookUidToHookTypeMap := make(map[string]string)
	for _, node := range nodes {
		if node.IsHook {
			hookUidToHookTypeMap[node.UID] = node.HookType
		}
	}
	// if node's parentRef is a hook then add hook info in child node also
	if len(hookUidToHookTypeMap) > 0 {
		for _, node := range nodes {
			if node.ParentRefs != nil && len(node.ParentRefs) > 0 {
				if hookType, ok := hookUidToHookTypeMap[node.ParentRefs[0].UID]; ok {
					node.IsHook = true
					node.HookType = hookType
				}
			}
		}
	}
}
