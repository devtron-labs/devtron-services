package registry

type registry string

func (r registry) String() string {
	return string(r)
}

const (
	DOCKER_REGISTRY_TYPE_ECR        registry = "ecr"
	DOCKER_REGISTRY_TYPE_ACR        registry = "acr"
	DOCKER_REGISTRY_TYPE_DOCKERHUB  registry = "docker-hub"
	DOCKER_REGISTRY_TYPE_OTHER      registry = "other"
	REGISTRY_TYPE_ARTIFACT_REGISTRY registry = "artifact-registry"
	REGISTRY_TYPE_GCR               registry = "gcr"
)

const JSON_KEY_USERNAME = "_json_key"

type RegistryCredential struct {
	RegistryType       registry `json:"registryType"`
	RegistryURL        string   `json:"registryURL"`
	Username           string   `json:"username"`
	Password           string   `json:"password"`
	AWSAccessKeyId     string   `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey string   `json:"awsSecretAccessKey,omitempty"`
	AWSRegion          string   `json:"awsRegion,omitempty"`
}
