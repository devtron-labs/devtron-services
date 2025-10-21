package veleroVSL

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/utils/storage"
	"github.com/pkg/errors"
	veleroVslBean "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

func (impl *InformerImpl) sendVslUpdate(vslChangeObj *storage.VeleroStorageEvent[storage.LocationsStatus]) error {
	if impl.client == nil {
		impl.logger.Errorw("pubsub client is nil - STORAGE_MODULE_TOPIC, skipping the publish")
		return errors.New("pubsub client is nil - STORAGE_MODULE_TOPIC, skipping the publish")
	}
	vslChangeObjByte, err := json.Marshal(vslChangeObj)
	if err != nil {
		impl.logger.Errorw("error in marshalling velero status update", "err", err)
		return err
	}
	err = impl.client.Publish(pubsub.STORAGE_MODULE_TOPIC, string(vslChangeObjByte))
	if err != nil {
		impl.logger.Errorw("error in publishing velero status update", "err", err)
		return err
	} else {
		impl.logger.Info("velero status update sent", "veleroStatusUpdate:", string(vslChangeObjByte))
		return nil
	}
}

func isChangeInVslObject(oldObj, newObj *veleroVslBean.VolumeSnapshotLocation, vslChangeObj *storage.VeleroStorageEvent[storage.LocationsStatus]) bool {
	if oldObj.Spec.Provider == newObj.Spec.Provider {
		return false
	}
	vslChangeObj.Data.Provider = newObj.Spec.Provider
	return true
}
