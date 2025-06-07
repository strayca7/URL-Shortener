package rbac

import (
	metav1 "url-shortener/pkg/meta/v1"
)

type PolicyRule struct {
	// Available APIGroups: rbac, ""
	APIGroups []string `json:"apiGroups,omitempty"`
	// Available Verbs: get, list, create, update, delete
	Verbs     []string `json:"verbs" validate:"required,min=1"`
	// Available Resources: role, rolebinding, user, group, admin
	Resources []string `json:"resources,omitempty" validate:"required,min=1"`
}

type Role struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Rules holds all the PolicyRules for this Role
	Rules []PolicyRule `json:"rules"`
}

type RoleRef struct {
	// Available APIGroup: rbac, ""
	APIGroup string `json:"apiGroup"`

	// Available Kind: Role, ClusterRole
	Kind string `json:"kind"`

	Name string `json:"name"`
}

// Same Sunject can be used in multiple RoleBindings.
type Subject struct {
	// Available Kind: User, Admin, ClusterAdmin
	Kind string `json:"kind"`
	// APIGroup holds the API group of the referenced subject.
	// Subject name must be unique within the namespace.
	Name string `json:"name"`
	// Namespace of the referenced object.  If the object kind is non-namespace, such as "User" or "Group", and this value is not empty
	// the Authorizer should report an error.
	Namespace string `json:"namespace,omitempty"`
}

type RoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	RoleRef  RoleRef   `json:"roleRef" validate:"required"`
	Subjects []Subject `json:"subjects,omitempty" validate:"required"`
}


