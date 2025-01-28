package cluster

import (
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster/argoCD"
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster/argoWf"
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster/systemExec"
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	argoCD.NewInformerImpl,
	argoWf.NewInformerImpl,
	systemExec.NewInformerImpl,

	NewInformerImpl,
	wire.Bind(new(Informer), new(*InformerImpl)),
)
