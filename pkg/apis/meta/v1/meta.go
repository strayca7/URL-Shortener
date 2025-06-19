package v1

import "time"

type Type interface {
	GetAPIVersion() string
	SetAPIVersion(version string)
	GetKind() string
	SetKind(kind string)
}

type Object interface {
	GetName() string
	SetName(name string)
	GetNamespace() string
	SetNamespace(namespace string)
	GetUID() string
	SetUID(uid string)
	GetCreationTimestamp() string
	SetCreationTimestamp(timestamp string)
	GetDeletionTimestamp() string
	SetDeletionTimestamp(timestamp string)
}

// TypeMeta contains the type information of an object.
func (meta *TypeMeta) GetAPIVersion() string {
	return meta.APIVersion
}
func (meta *TypeMeta) SetAPIVersion(version string) {
	meta.APIVersion = version
}
func (meta *TypeMeta) GetKind() string {
	return meta.Kind
}
func (meta *TypeMeta) SetKind(kind string) {
	meta.Kind = kind
}

// ObjectMeta contains the metadata of an object.
func (meta *ObjectMeta) GetName() string {
	return meta.Name
}
func (meta *ObjectMeta) SetName(name string) {
	meta.Name = name
}
func (meta *ObjectMeta) GetNamespace() string {
	return meta.Namespace
}
func (meta *ObjectMeta) SetNamespace(namespace string) {
	meta.Namespace = namespace
}
func (meta *ObjectMeta) GetUID() string {
	return meta.UID
}
func (meta *ObjectMeta) SetUID(uid string) {
	meta.UID = uid
}
func (meta *ObjectMeta) GetCreationTimestamp() time.Time {
	return meta.CreationTimestamp
}
func (meta *ObjectMeta) SetCreationTimestamp(timestamp time.Time) {
	meta.CreationTimestamp = timestamp
}
func (meta *ObjectMeta) GetDeletionTimestamp() string {
	if meta.DeletionTimestamp.IsZero() {
		return ""
	}
	return meta.DeletionTimestamp.Format(time.RFC3339)
}
func (meta *ObjectMeta) SetDeletionTimestamp(timestamp *time.Time) {
	meta.DeletionTimestamp = timestamp
}
