package rbac

import (
	metav1 "url-shortener/pkg/apis/meta/v1"
)

type PolicyRule struct {
	// Available APIGroups: rbac/v1, ""
	APIGroups []string `json:"apiGroups,omitempty" yaml:"apiGroups,omitempty"`
	// Available Verbs: get, list, create, update, delete
	Verbs []string `json:"verbs" yaml:"verbs"`
	// Available Resources: role, rolebinding, user, urls
	Resources []string `json:"resources,omitempty" yaml:"resources"`
}

type Role struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Rules holds all the PolicyRules for this Role
	Rules []PolicyRule `json:"rules" yaml:"rules"`
}

type RoleRef struct {
	// Available APIGroup: "rbac/v1", "", "v1"
	APIGroup string `json:"apiGroup"`

	// Available Kind: User
	Kind string `json:"kind"`

	Name string `json:"name"`
}

// Same Sunject can be used in multiple RoleBindings.
type Subject struct {
	// Available Kind: User, Admin
	Kind string `json:"kind"`
	// APIGroup holds the API group of the referenced subject.
	// Subject name must be unique within the namespace.
	Name string `json:"name"`
	// Namespace of the referenced object.
	Namespace string `json:"namespace,omitempty"`
}

type RoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	RoleRef  RoleRef   `json:"roleRef"`
	Subjects []Subject `json:"subjects,omitempty"`
}
