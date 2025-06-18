package k8sResource

import "github.com/caarlos0/env"

// CATEGORY=K8S_RESOURCE_SERVICE_CONFIG
type ServiceConfig struct {
	ParentChildGvkMapping      string `env:"PARENT_CHILD_GVK_MAPPING" envDefault:"" description:"Parent child GVK mapping for resource tree" deprecated:"false" example:""`
	ChildObjectListingPageSize int64  `env:"CHILD_OBJECT_LISTING_PAGE_SIZE" envDefault:"1000" description:"Resource tree child object listing page size" deprecated:"false" example:"100"`
}

func GetK8sResourceConfig() (*ServiceConfig, error) {
	cfg := &ServiceConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return cfg, err
	}
	if cfg.ChildObjectListingPageSize <= 10 {
		// set the default value for invalid values
		cfg.ChildObjectListingPageSize = 1000
	}
	return cfg, nil
}
