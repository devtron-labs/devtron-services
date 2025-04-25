package adapter

import (
	"github.com/devtron-labs/common-lib/helmLib/registry"
	remoteConnection "github.com/devtron-labs/common-lib/utils/remoteConnection/bean"
	client "github.com/devtron-labs/kubelink/grpc"
	"github.com/devtron-labs/kubelink/pkg/service/helmApplicationService/release"
	"github.com/devtron-labs/kubelink/pkg/util"
	"google.golang.org/protobuf/types/known/timestamppb"
	helmChart "helm.sh/helm/v3/pkg/chart"
	helmRelease "helm.sh/helm/v3/pkg/release"
)

func NewRegistryConfig(credential *client.RegistryCredential) (*registry.Configuration, error) {
	var registryConfig *registry.Configuration
	if credential != nil {
		registryConfig = &registry.Configuration{
			RegistryId:                credential.RegistryName,
			RegistryUrl:               credential.RegistryUrl,
			Username:                  credential.Username,
			Password:                  credential.Password,
			AwsAccessKey:              credential.AccessKey,
			AwsSecretKey:              credential.SecretKey,
			AwsRegion:                 credential.AwsRegion,
			RegistryConnectionType:    credential.Connection,
			RegistryCertificateString: credential.RegistryCertificate,
			RegistryType:              credential.RegistryType,
			IsPublicRegistry:          credential.IsPublic,
			CredentialsType:           credential.CredentialsType,
		}

		if credential.Connection == registry.SECURE_WITH_CERT {
			certificatePath, err := registry.CreateCertificateFile(credential.RegistryName, credential.RegistryCertificate)
			if err != nil {
				return nil, err
			}
			registryConfig.RegistryCAFilePath = certificatePath
		}

		connectionConfig := credential.RemoteConnectionConfig
		if connectionConfig != nil {
			registryConfig.RemoteConnectionConfig = &remoteConnection.RemoteConnectionConfigBean{}
			switch connectionConfig.RemoteConnectionMethod {
			case client.RemoteConnectionMethod_DIRECT:
				registryConfig.RemoteConnectionConfig.ConnectionMethod = remoteConnection.RemoteConnectionMethodDirect
			case client.RemoteConnectionMethod_PROXY:
				registryConfig.RemoteConnectionConfig.ConnectionMethod = remoteConnection.RemoteConnectionMethodProxy
				registryConfig.RemoteConnectionConfig.ProxyConfig = ConvertConfigToProxyConfig(connectionConfig)
			case client.RemoteConnectionMethod_SSH:
				registryConfig.RemoteConnectionConfig.ConnectionMethod = remoteConnection.RemoteConnectionMethodSSH
				registryConfig.RemoteConnectionConfig.SSHTunnelConfig = ConvertConfigToSSHTunnelConfig(connectionConfig)
			}
		}
	}
	return registryConfig, nil
}

func ConvertConfigToProxyConfig(config *client.RemoteConnectionConfig) *remoteConnection.ProxyConfig {
	var proxyConfig *remoteConnection.ProxyConfig
	if config.ProxyConfig != nil {
		proxyConfig = &remoteConnection.ProxyConfig{
			ProxyUrl: config.ProxyConfig.ProxyUrl,
		}
	}
	return proxyConfig
}

func ConvertConfigToSSHTunnelConfig(config *client.RemoteConnectionConfig) *remoteConnection.SSHTunnelConfig {
	var sshConfig *remoteConnection.SSHTunnelConfig
	if config.SSHTunnelConfig != nil {
		sshConfig = &remoteConnection.SSHTunnelConfig{
			SSHUsername:      config.SSHTunnelConfig.SSHUsername,
			SSHPassword:      config.SSHTunnelConfig.SSHPassword,
			SSHAuthKey:       config.SSHTunnelConfig.SSHAuthKey,
			SSHServerAddress: config.SSHTunnelConfig.SSHServerAddress,
		}
	}
	return sshConfig
}

func NewDeployedAppDetail(config *client.ClusterConfig, release *release.Release) *client.DeployedAppDetail {
	return &client.DeployedAppDetail{
		AppId:        util.GetAppId(config.ClusterId, release),
		AppName:      release.Name,
		ChartName:    release.Chart.Name(),
		ChartAvatar:  release.Chart.Metadata.Icon,
		LastDeployed: timestamppb.New(release.Info.LastDeployed.Time),
		EnvironmentDetail: &client.EnvironmentDetails{
			ClusterName: config.ClusterName,
			ClusterId:   config.ClusterId,
			Namespace:   release.Namespace,
		},
	}
}

func ParseDeployedAppDetail(clusterId int32, clusterName string, helmRelease *release.Release) *client.DeployedAppDetail {
	appDetail := &client.DeployedAppDetail{
		AppId:        util.GetAppId(clusterId, helmRelease),
		AppName:      helmRelease.Name,
		ChartName:    helmRelease.Chart.Name(),
		ChartAvatar:  helmRelease.Chart.Metadata.Icon,
		LastDeployed: timestamppb.New(helmRelease.Info.LastDeployed.Time),
		ChartVersion: helmRelease.Chart.Metadata.Version,
		EnvironmentDetail: &client.EnvironmentDetails{
			ClusterName: clusterName,
			ClusterId:   clusterId,
			Namespace:   helmRelease.Namespace,
		},
		ReleaseStatus: helmRelease.Info.Status.String(),
		Home:          helmRelease.Chart.Metadata.Home,
	}
	return appDetail
}

func GetAppDetailRequestFromGetResourceTreeRequest(req *client.GetResourceTreeRequest) *client.AppDetailRequest {
	return &client.AppDetailRequest{
		ClusterConfig: req.ClusterConfig,
		Namespace:     req.Namespace,
		ReleaseName:   req.GetReleaseName(),
		PreferCache:   req.PreferCache,
		UseFallBack:   req.UseFallBack,
		CacheConfig:   req.CacheConfig,
	}
}

func NewRelease(helmRelease *helmRelease.Release) *release.Release {
	if helmRelease == nil {
		return nil
	}
	return &release.Release{
		Name:      helmRelease.Name,
		Namespace: helmRelease.Namespace,
		Info:      NewReleaseInfo(helmRelease.Info),
		Chart:     NewChart(helmRelease.Chart),
	}
}

func NewReleaseInfo(helmReleaseInfo *helmRelease.Info) *release.Info {
	if helmReleaseInfo == nil {
		return nil
	}
	return &release.Info{
		LastDeployed: helmReleaseInfo.LastDeployed,
		Status:       helmReleaseInfo.Status,
	}
}

func NewChart(helmChart *helmChart.Chart) *release.Chart {
	if helmChart == nil {
		return nil
	}
	return &release.Chart{
		Metadata: NewChartMetadata(helmChart.Metadata),
	}
}

func NewChartMetadata(helmChartMetadata *helmChart.Metadata) *release.Metadata {
	if helmChartMetadata == nil {
		return nil
	}
	return &release.Metadata{
		Name:    helmChartMetadata.Name,
		Version: helmChartMetadata.Version,
		Icon:    helmChartMetadata.Icon,
		Home:    helmChartMetadata.Home,
	}
}
