package veleroBSL

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/utils/storage"
	"github.com/pkg/errors"
	veleroBslBean "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

func (impl *InformerImpl) sendBslUpdate(bslChangeObj *storage.VeleroStorageEvent[storage.LocationsStatus]) error {
	if impl.client == nil {
		impl.logger.Errorw("pubsub client is nil - STORAGE_MODULE_TOPIC, skipping the publish")
		return errors.New("pubsub client is nil - STORAGE_MODULE_TOPIC, skipping the publish")
	}
	//bslChangeobjJson, err := json.Marshal(veleroStatusUpdate)
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

func isChangeInBslObject(oldObj, newObj *veleroBslBean.BackupStorageLocation, bslChangeObj *storage.VeleroStorageEvent[storage.LocationsStatus]) bool {
	if oldObj.Spec.Provider == newObj.Spec.Provider && oldObj.Status.Phase == newObj.Status.Phase {
		return false
	}
	if oldObj.Spec.Provider != newObj.Spec.Provider {
		bslChangeObj.Data.Provider = newObj.Spec.Provider
	}
	if oldObj.Status.Phase != newObj.Status.Phase {
		bslChangeObj.Data.Status = newObj.Status
	}
	return true
}
