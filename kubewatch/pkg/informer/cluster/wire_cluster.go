package cluster

import (
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster/argoCD"
	cdWf "github.com/devtron-labs/kubewatch/pkg/informer/cluster/argoWf/cd"
	ciWf "github.com/devtron-labs/kubewatch/pkg/informer/cluster/argoWf/ci"
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster/systemExec"
	veleroBackupInformer "github.com/devtron-labs/kubewatch/pkg/informer/cluster/velero/backup"
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster/velero/backupStorageLocation"
	veleroRestoreInformer "github.com/devtron-labs/kubewatch/pkg/informer/cluster/velero/restore"
	"github.com/devtron-labs/kubewatch/pkg/informer/cluster/velero/volumeSnapshotLocation"
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	argoCD.NewInformerImpl,
	cdWf.NewInformerImpl,
	ciWf.NewInformerImpl,
	systemExec.NewInformerImpl,
	veleroBslInformer.NewInformerImpl,
	veleroVslInformer.NewInformerImpl,
	veleroBackupInformer.NewInformerImpl,
	veleroRestoreInformer.NewInformerImpl,

	NewInformerImpl,
	wire.Bind(new(Informer), new(*InformerImpl)),
)
