// The writing habits of all fields in package meta are the same as Kubernets.
package v1

import "time"

type TypeMeta struct {
	// Available APIVersion: v1, rbac/v1
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`

	// Available Kind: role, rolebinding
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

type ObjectMeta struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// If Namespace is empty, it means "default".
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// Generate with UUID when an obeject is be created automatically.
	UID string `json:"uid,omitempty" yaml:"uid,omitempty"`

	// Generate when an obeject is be created automatically.
	// In particular, CreationTimestamp field can not be changed once it is created.
	CreationTimestamp time.Time `json:"creationTimestamp,omitempty,omitzero" yaml:"creationTimestamp,omitempty"`

	// Generate when an obeject is be created automatically.
	// CreationTimestamp is nil once it is created. After deletion, this field must have a value.
	DeletionTimestamp *time.Time `json:"deletionTimestamp,omitempty" yaml:"deletionTimestamp,omitempty"`
}
