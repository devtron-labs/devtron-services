package k8sResource

import "github.com/google/wire"

var WireSet = wire.NewSet(
	GetK8sResourceConfig,
	NewK8sServiceImpl,
	wire.Bind(new(K8sService), new(*K8sServiceImpl)),
)
