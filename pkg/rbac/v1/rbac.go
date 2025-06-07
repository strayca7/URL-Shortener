package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"url-shortener/pkg/rbac"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	rolePrefix    = "/rbac/v1/role/"
	rolebindingPrefix = "/rbac/v1/rolebindings/"
)

type RBACSystem struct {
	etcdClient *clientv3.Client
	roles      map[string]rbac.Role
	rolebindings   map[string]rbac.RoleBinding
	mu         sync.RWMutex
}

// NewRBACSystem initializes a new RBAC system with an etcd client.
//
//	type RBACSystem struct {
//		etcdClient *clientv3.Client
//		roles      map[string]rbac.Role
//		rolebindings   map[string]rbac.RoleBinding
//		mu         sync.RWMutex
//	}
func NewRBACSystem(etcdClient *clientv3.Client) *RBACSystem {
	return &RBACSystem{
		etcdClient: etcdClient,
		roles:      make(map[string]rbac.Role),
		rolebindings:   make(map[string]rbac.RoleBinding),
		mu:         sync.RWMutex{},
	}
}

func (r *RBACSystem) Close() error {
	if r.etcdClient != nil {
		if err := r.etcdClient.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close etcd client")
			return err
		}
	}
	return nil
}

func (r *RBACSystem) LoadInitialData() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	log.Debug().Msg("Loading initial data into RBAC system")

	if err := r.LoadData(rolePrefix, func(key string, value []byte) error {
		var role rbac.Role
		if err := json.Unmarshal(value, &role); err != nil {
			log.Error().Err(err).Str("key", key).Msg("Failed to unmarshal role data")
			return fmt.Errorf("failed to unmarshal role data: %w", err)
		}
		roleKey := r.generateKey(role.Namespace, role.Name)
		r.roles[roleKey] = role
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("Failed to load roles from etcd")
		return err
	}

	if err := r.LoadData(rolebindingPrefix, func(key string, value []byte) error {
		var rolebinding rbac.RoleBinding
		if err := json.Unmarshal(value, &rolebinding); err != nil {
			log.Error().Err(err).Str("key", key).Msg("Failed to unmarshal rolebinding data")
			return fmt.Errorf("failed to unmarshal role binding data: %w", err)
		}
		rolebindingKey := r.generateKey(rolebinding.Namespace, rolebinding.Name)
		r.rolebindings[rolebindingKey] = rolebinding
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("Failed to load role bindings from etcd")
		return err
	}

	return nil
}

// LoadData is a tipical method to load data from etcd into the RBAC system.
// Func processor is used to process each key-value pair with custom logic.
func (r *RBACSystem) LoadData(prefix string, processor func(string, []byte) error) error {
	resp, err := r.etcdClient.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get data from etcd")
		return fmt.Errorf("failed to get data from etcd: %w", err)
	}

	for _, kv := range resp.Kvs {
		if err := processor(string(kv.Key), kv.Value); err != nil {
			log.Error().Err(err).Str("key", string(kv.Key)).Msg("Failed to process key-value pair")
			return err
		}
	}
	return nil
}

// generateKey generates a key which will be stored in RBACSystem.roles or RBACSystem.rolebindings.
func (r *RBACSystem) generateKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

// generateEtcdKey generates a key which will be stored in etcd.
func (r *RBACSystem) generateEtcdKey(prefix, namespace, name string) string {
	return fmt.Sprintf("%s%s/%s", prefix, namespace, name)
}

// saveToEtcd marshals the provided data into JSON and saves it to etcd under the specified key.
func (r *RBACSystem) saveToEtcd(key string, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal value to JSON")
		return err
	}

	if _, err := r.etcdClient.Put(context.Background(), key, string(jsonData)); err != nil {
		log.Error().Err(err).Str("key", key).Msg("Failed to put data into etcd")
		return err
	}
	return nil
}

func (r *RBACSystem) HandleCreateRole(c *gin.Context) {
	var role rbac.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid role data", err)
		return
	}

	// Validation move to marshal yaml to json pkg/marshal/
	// When yaml is marshaled to JSON, the fields like "CreationTimestamo" and
	// "UID" are be set automatically.

	roleKey := r.generateKey(role.Namespace, role.Name)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.roles[roleKey]; exists {
		errorResponse(c, http.StatusConflict, "Role already exists", nil)
		return
	}

	etcdKey := r.generateEtcdKey(rolePrefix, role.Namespace, role.Name)
	if err := r.saveToEtcd(etcdKey, role); err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to save role", err)
		return
	}

	// avoid conflicts when saveToEtcd is unsuccessful
	r.roles[roleKey] = role

	successResponse(c, http.StatusCreated, "Role created successfully", role)
}

func (r *RBACSystem) HandleCreateRoleBinding(c *gin.Context) {
	var rolebinding rbac.RoleBinding
	if err := c.ShouldBindJSON(&rolebinding); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid role binding data", err)
		return
	}

	// Validation move to marshal yaml to json pkg/marshal/
	// When yaml is marshaled to JSON, the fields like "CreationTimestamo" and
	// "UID" are be set automatically.

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if the role exists before creating the rolebinding
	roleKey := r.generateKey(rolebinding.Namespace, rolebinding.RoleRef.Name)
	if _, exists := r.roles[roleKey]; !exists {
		errorResponse(c, http.StatusNotFound, "Role not found", nil)
		return
	}

	rolebindingkey := r.generateKey(rolebinding.Namespace, rolebinding.Name)
	if _, exists := r.rolebindings[rolebindingkey]; exists {
		errorResponse(c, http.StatusConflict, "Role binding already exists", nil)
		return
	}

	etcdKey := r.generateEtcdKey(rolebindingPrefix, rolebinding.Namespace, rolebinding.Name)
	if err := r.saveToEtcd(etcdKey, rolebinding); err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to save role binding", err)
		return
	}

	// avoid conflicts when saveToEtcd is unsuccessful
	r.rolebindings[rolebindingkey] = rolebinding

	successResponse(c, http.StatusCreated, "Role binding created successfully", rolebinding)
}

// HandleListRoles is a list function for retrieving all roles.
func (r *RBACSystem) HandleListRoles(c *gin.Context) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roles := make([]rbac.Role, 0, len(r.roles))
	for _, role := range r.roles {
		roles = append(roles, role)
	}

	successResponse(c, http.StatusOK, "Roles retrieved successfully", roles)
}

// HandleListRoleBindings is a list function for retrieving all role bindings.
func (r *RBACSystem) HandleListRoleBindings(c *gin.Context) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rolebindings := make([]rbac.RoleBinding, 0, len(r.rolebindings))
	for _, rolebinding := range r.rolebindings {
		rolebindings = append(rolebindings, rolebinding)
	}

	successResponse(c, http.StatusOK, "Role bindings retrieved successfully", rolebindings)
}
