package config

import "github.com/caarlos0/env"

// CATEGORY=VELERO_INFORMER
type VeleroConfig struct {
	// VeleroInformer is used to determine whether Velero informer is enabled or not
	VeleroInformer bool `env:"VELERO_INFORMER" envDefault:"false" description:"Used to determine whether Velero informer is enabled or not" deprecated:"false"`

	// VeleroNamespace is the namespace where all the Velero backup objects are published
	VeleroNamespace string `env:"VELERO_NAMESPACE" envDefault:"velero" description:"Namespace where all the Velero backup objects are published" deprecated:"false"`
}

func getVeleroConfig() (*VeleroConfig, error) {
	veleroConfig := &VeleroConfig{}
	err := env.Parse(veleroConfig)
	return veleroConfig, err
}

func (v *VeleroConfig) GetVeleroNamespace() string {
	return v.VeleroNamespace
}
