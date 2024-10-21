package commonHelmService

import (
	client "github.com/devtron-labs/kubelink/grpc"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func GetObjectIdentifierFromHelmManifest(manifest *unstructured.Unstructured, namespace string) *client.ObjectIdentifier {
	gvk := manifest.GroupVersionKind()
	namespaceManifest := manifest.GetNamespace()
	if namespaceManifest == "" {
		namespaceManifest = namespace
	}
	return &client.ObjectIdentifier{
		Group:     gvk.Group,
		Version:   gvk.Version,
		Kind:      gvk.Kind,
		Name:      manifest.GetName(),
		Namespace: namespaceManifest,
	}
}

func GetObjectIdentifierFromExternalResource(externalResourceList []*client.ExternalResourceDetail) []*client.ObjectIdentifier {
	if len(externalResourceList) == 0 {
		return []*client.ObjectIdentifier{}
	}
	resp := make([]*client.ObjectIdentifier, len(externalResourceList))
	for _, externalResource := range externalResourceList {
		objectIdentifier := &client.ObjectIdentifier{
			Group:     externalResource.Group,
			Kind:      externalResource.Kind,
			Version:   externalResource.Version,
			Name:      externalResource.Name,
			Namespace: externalResource.Namespace,
		}
		resp = append(resp, objectIdentifier)
	}
	return resp
}
