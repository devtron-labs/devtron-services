package storage

import (
	veleroBean "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

// EventType represents the type of event
type EventType string

func (e EventType) String() string {
	return string(e)
}

func (e EventType) IsCreated() bool {
	return e == EventTypeAdded
}

func (e EventType) IsUpdated() bool {
	return e == EventTypeUpdated
}

func (e EventType) IsDeleted() bool {
	return e == EventTypeDeleted
}

const (
	EventTypeAdded   EventType = "ADDED"
	EventTypeUpdated EventType = "UPDATED"
	EventTypeDeleted EventType = "DELETED"
)

// ResourceKind represents the kind of resource
type ResourceKind string

func (r ResourceKind) String() string {
	return string(r)
}

func (r ResourceKind) IsBackup() bool {
	return r == ResourceBackup
}

func (r ResourceKind) IsRestore() bool {
	return r == ResourceRestore
}

func (r ResourceKind) IsBackupStorageLocation() bool {
	return r == ResourceBackupStorageLocation
}

func (r ResourceKind) IsVolumeSnapshotLocation() bool {
	return r == ResourceVolumeSnapshotLocation
}

func (r ResourceKind) IsBackupSchedule() bool {
	return r == ResourceBackupSchedule
}

const (
	ResourceBackup                 ResourceKind = "Backup"
	ResourceRestore                ResourceKind = "Restore"
	ResourceBackupStorageLocation  ResourceKind = "BackupStorageLocation"
	ResourceVolumeSnapshotLocation ResourceKind = "VolumeSnapshotLocation"
	ResourceBackupSchedule         ResourceKind = "BackupSchedule"
)

// LocationsStatus represents the status of a location
// NOTE: status is only available in case of BSL
type LocationsStatus struct {
	*veleroBean.BackupStorageLocationStatus
}

// BackupStatus represents the status of a backup
type BackupStatus struct {
	*veleroBean.BackupStatus
}

// RestoreStatus represents the status of a restore
type RestoreStatus struct {
	*veleroBean.RestoreStatus
}

// BackupScheduleStatus represents the status of a backup schedule
type BackupScheduleStatus struct {
	*veleroBean.ScheduleStatus
}

// VeleroResourceEvent represents the event sent by velero
type VeleroResourceEvent struct {
	EventType    EventType    `json:"eventType"`
	ResourceKind ResourceKind `json:"kind"`
	ClusterId    int          `json:"clusterId"`
	ResourceName string       `json:"resourceName"`
}

func NewVeleroResourceEvent() *VeleroResourceEvent {
	return &VeleroResourceEvent{}
}

// Getters

// GetEventType returns the EventType
func (e *VeleroResourceEvent) GetEventType() any {
	return e.EventType
}

// GetResourceKind returns the ResourceKind
func (e *VeleroResourceEvent) GetResourceKind() ResourceKind {
	return e.ResourceKind
}

// GetClusterId returns the ClusterId
func (e *VeleroResourceEvent) GetClusterId() int {
	return e.ClusterId
}

// GetResourceName returns the ResourceName
func (e *VeleroResourceEvent) GetResourceName() string {
	return e.ResourceName
}

// Setters

// SetEventType sets the EventType
func (e *VeleroResourceEvent) SetEventType(eventType EventType) *VeleroResourceEvent {
	e.EventType = eventType
	return e
}

// SetClusterId sets the ClusterId
func (e *VeleroResourceEvent) SetClusterId(clusterId int) *VeleroResourceEvent {
	e.ClusterId = clusterId
	return e
}

// SetResourceKind sets the ResourceKind
func (e *VeleroResourceEvent) SetResourceKind(resourceKind ResourceKind) *VeleroResourceEvent {
	e.ResourceKind = resourceKind
	return e
}

// SetResourceName sets the ResourceName
func (e *VeleroResourceEvent) SetResourceName(resourceName string) *VeleroResourceEvent {
	e.ResourceName = resourceName
	return e
}
