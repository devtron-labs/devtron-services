package commonHelmService

import (
	"context"
	k8sUtils "github.com/devtron-labs/common-lib/utils/k8s"
	yamlUtil "github.com/devtron-labs/common-lib/utils/yaml"
	"github.com/devtron-labs/kubelink/bean"
	globalConfig "github.com/devtron-labs/kubelink/config"
	"github.com/devtron-labs/kubelink/converter"
	client "github.com/devtron-labs/kubelink/grpc"
	"github.com/devtron-labs/kubelink/pkg/helmClient"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

type CommonHelmService interface {
	GetHelmRelease(clusterConfig *client.ClusterConfig, namespace string, releaseName string) (*release.Release, error)
	BuildResourceTreeForHelmRelease(ctx context.Context, appDetailRequest *client.AppDetailRequest, release *release.Release) (*bean.ResourceTreeResponse, error)
	GetResourceTreeForExternalResources(ctx context.Context, req *client.ExternalResourceTreeRequest) (*bean.ResourceTreeResponse, error)
	GetRestConfigForClusterConfig(clusterConfig *client.ClusterConfig) (*rest.Config, error)
}

type CommonHelmServiceImpl struct {
	k8sService          K8sService
	logger              *zap.SugaredLogger
	k8sUtil             k8sUtils.K8sService
	converter           converter.ClusterBeanConverter
	helmReleaseConfig   *globalConfig.HelmReleaseConfig
	resourceTreeService ResourceTreeService
}

func NewCommonHelmServiceImpl(logger *zap.SugaredLogger,
	k8sUtil k8sUtils.K8sService, converter converter.ClusterBeanConverter,
	k8sService K8sService, helmReleaseConfig *globalConfig.HelmReleaseConfig,
	resourceTreeService ResourceTreeService) *CommonHelmServiceImpl {
	return &CommonHelmServiceImpl{
		logger:              logger,
		k8sUtil:             k8sUtil,
		converter:           converter,
		k8sService:          k8sService,
		helmReleaseConfig:   helmReleaseConfig,
		resourceTreeService: resourceTreeService,
	}

}

func (impl *CommonHelmServiceImpl) GetHelmRelease(clusterConfig *client.ClusterConfig, namespace string, releaseName string) (*release.Release, error) {

	k8sClusterConfig := impl.converter.GetClusterConfigFromClientBean(clusterConfig)
	conf, err := impl.k8sUtil.GetRestConfigByCluster(k8sClusterConfig)
	if err != nil {
		return nil, err
	}
	opt := &helmClient.RestConfClientOptions{
		Options: &helmClient.Options{
			Namespace: namespace,
		},
		RestConfig: conf,
	}
	helmClient, err := helmClient.NewClientFromRestConf(opt)
	if err != nil {
		return nil, err
	}
	release, err := helmClient.GetRelease(releaseName)
	if err != nil {
		return nil, err
	}
	return release, nil
}

func (impl *CommonHelmServiceImpl) BuildResourceTreeForHelmRelease(ctx context.Context, appDetailRequest *client.AppDetailRequest, release *release.Release) (*bean.ResourceTreeResponse, error) {

	conf, err := impl.GetRestConfigForClusterConfig(appDetailRequest.ClusterConfig)
	if err != nil {
		return nil, err
	}

	parentObjects, err := impl.getObjectIdentifiersFromHelmRelease(release, appDetailRequest.Namespace)
	if err != nil {
		impl.logger.Errorw("Error in getting helm release", "appDetailRequest", appDetailRequest, "err", err)
		return nil, err
	}
	
	return impl.resourceTreeService.BuildResourceTreeUsingParentObjects(ctx, appDetailRequest, conf, parentObjects)

}

func (impl *CommonHelmServiceImpl) GetRestConfigForClusterConfig(clusterConfig *client.ClusterConfig) (*rest.Config, error) {
	k8sClusterConfig := impl.converter.GetClusterConfigFromClientBean(clusterConfig)
	conf, err := impl.k8sUtil.GetRestConfigByCluster(k8sClusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster", "clusterName", k8sClusterConfig.ClusterName)
		return nil, err
	}
	return conf, nil
}

func (impl *CommonHelmServiceImpl) getObjectIdentifiersFromHelmRelease(helmRelease *release.Release, namespace string) ([]*client.ObjectIdentifier, error) {
	manifests, err := impl.getManifestsFromHelmRelease(helmRelease)
	if err != nil {
		impl.logger.Errorw("Error in getting helm release", "helmRelease", helmRelease, "err", err)
		return nil, err
	}

	objectIdentifiers := make([]*client.ObjectIdentifier, 0)
	for _, manifest := range manifests {
		objectIdentifier := GetObjectIdentifierFromHelmManifest(&manifest, namespace)
		if objectIdentifier != nil && len(objectIdentifier.Kind) > 0 && len(objectIdentifier.Name) > 0 {
			objectIdentifiers = append(objectIdentifiers, objectIdentifier)
		}
	}

	return objectIdentifiers, nil
}

func (impl *CommonHelmServiceImpl) getManifestsFromHelmRelease(helmRelease *release.Release) ([]unstructured.Unstructured, error) {
	manifests, err := yamlUtil.SplitYAMLs([]byte(helmRelease.Manifest))
	if err != nil {
		return nil, err
	}
	manifests = impl.addHookResourcesInManifest(helmRelease, manifests)
	return manifests, nil
}

func (impl *CommonHelmServiceImpl) addHookResourcesInManifest(helmRelease *release.Release, manifests []unstructured.Unstructured) []unstructured.Unstructured {
	for _, helmHook := range helmRelease.Hooks {
		var hook unstructured.Unstructured
		err := yaml.Unmarshal([]byte(helmHook.Manifest), &hook)
		if err != nil {
			impl.logger.Errorw("error in converting string manifest into unstructured obj", "hookName", helmHook.Name, "releaseName", helmRelease.Name, "err", err)
			continue
		}
		manifests = append(manifests, hook)
	}
	return manifests
}

func (impl *CommonHelmServiceImpl) GetResourceTreeForExternalResources(ctx context.Context, req *client.ExternalResourceTreeRequest) (*bean.ResourceTreeResponse, error) {
	k8sClusterConfig := impl.converter.GetClusterConfigFromClientBean(req.ClusterConfig)
	restConfig, err := impl.k8sUtil.GetRestConfigByCluster(k8sClusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting restConfig", "err", err)
		return nil, err
	}

	return impl.resourceTreeService.BuildResourceTreeUsingParentObjects(ctx, &client.AppDetailRequest{
		ClusterConfig: req.ClusterConfig,
		PreferCache:   req.PreferCache,
		UseFallBack:   req.UseFallBack,
		CacheConfig:   req.CacheConfig,
	}, restConfig, GetObjectIdentifierFromExternalResource(req.ExternalResourceDetail))
}
