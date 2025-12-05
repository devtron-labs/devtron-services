package veleroBSL

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/utils/storage"
	"github.com/pkg/errors"
	veleroBslBean "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

func (impl *InformerImpl) sendBslUpdate(bslChangeObj *storage.VeleroResourceEvent) error {
	if impl.client == nil {
		impl.logger.Errorw("pubsub client is nil - STORAGE_MODULE_TOPIC, skipping the publish")
		return errors.New("pubsub client is nil - STORAGE_MODULE_TOPIC, skipping the publish")
	}
	bslChangeObjByte, err := json.Marshal(bslChangeObj)
	if err != nil {
		impl.logger.Errorw("error in marshalling velero status update", "err", err)
		return err
	}
	err = impl.client.Publish(pubsub.STORAGE_MODULE_TOPIC, string(bslChangeObjByte))
	if err != nil {
		impl.logger.Errorw("error in publishing velero status update", "err", err)
		return err
	} else {
		impl.logger.Info("velero status update sent", "veleroStatusUpdate:", string(bslChangeObjByte))
		return nil
	}
}

func isChangeInBslObject(oldObj, newObj *veleroBslBean.BackupStorageLocation) bool {
	return oldObj.Status.Phase != newObj.Status.Phase ||
		!oldObj.Status.LastSyncedTime.Equal(newObj.Status.LastSyncedTime) ||
		!oldObj.Status.LastValidationTime.Equal(newObj.Status.LastValidationTime) ||
		oldObj.Status.Message != newObj.Status.Message ||
		oldObj.Status.LastSyncedRevision != newObj.Status.LastSyncedRevision ||
		oldObj.Status.AccessMode != newObj.Status.AccessMode
}
