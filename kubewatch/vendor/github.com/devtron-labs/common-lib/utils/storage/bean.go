package storage

import (
	"encoding/json"
	veleroBean "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

type EventType string
type ResourceKind string

const (
	EventTypeAdded   EventType = "ADDED"
	EventTypeUpdated EventType = "UPDATED"
	EventTypeDeleted EventType = "DELETED"

	ResourceBackup                 ResourceKind = "Backup"
	ResourceRestore                ResourceKind = "Restore"
	ResourceBackupStorageLocation  ResourceKind = "BackupStorageLocation"
	ResourceVolumeSnapshotLocation ResourceKind = "VolumeSnapshotLocation"
)

type VeleoroBslStatusUpdate struct {
	ClusterId int    `json:"clusterId"`
	BslName   string `json:"bslName"`
	Status    string `json:"status"`
}

// NOTE: status is only available in case of BSL
type LocationsStatus struct {
	Provider string                                 `json:"provider,omitempty"`
	Status   veleroBean.BackupStorageLocationStatus `json:"status,omitempty"`
}

type VeleroStorageEvent[T any] struct {
	EventType    EventType    `json:"eventType"`
	ResourceKind ResourceKind `json:"kind"`
	ClusterId    int          `json:"clusterId"`
	ResourceName string       `json:"resourceName"`
	Data         T            `json:"data,omitempty"`
}

// Getters

// GetEventType returns the EventType
func (e *VeleroStorageEvent[T]) GetEventType() any {
	return e.EventType
}

// GetResourceKind returns the ResourceKind
func (e *VeleroStorageEvent[T]) GetResourceKind() ResourceKind {
	return e.ResourceKind
}

// GetClusterId returns the ClusterId
func (e *VeleroStorageEvent[T]) GetClusterId() int {
	return e.ClusterId
}

// GetResourceName returns the ResourceName
func (e *VeleroStorageEvent[T]) GetResourceName() string {
	return e.ResourceName
}

// GetData returns the Data
func (e *VeleroStorageEvent[T]) GetData() T {
	return e.Data
}

// Setters

// SetEventType sets the EventType
func (e *VeleroStorageEvent[T]) SetEventType(eventType EventType) {
	e.EventType = eventType
}

// SetClusterId sets the ClusterId
func (e *VeleroStorageEvent[T]) SetClusterId(clusterId int) {
	e.ClusterId = clusterId
}

// SetResourceName sets the ResourceName
func (e *VeleroStorageEvent[T]) SetResourceName(resourceName string) {
	e.ResourceName = resourceName
}

// SetData sets the Data
func (e *VeleroStorageEvent[T]) SetData(data T) {
	e.Data = data
}

// SetResourceKind sets the ResourceKind
func (e *VeleroStorageEvent[T]) SetResourceKind(resourceKind ResourceKind) {
	e.ResourceKind = resourceKind
}

// JSON unmarshalling and marshalling
func (e *VeleroStorageEvent[T]) UnmarshalJSON(data []byte) error {
	var event VeleroStorageEvent[T]
	err := json.Unmarshal(data, &event)
	if err != nil {
		return err
	}
	*e = event
	return nil
}
func (e *VeleroStorageEvent[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(*e)
}
