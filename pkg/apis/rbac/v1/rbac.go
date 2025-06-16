package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
	metav1 "url-shortener/pkg/apis/meta/v1"
	"url-shortener/pkg/apis/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	rolePrefix        = "/rbac/v1/role/"
	rolebindingPrefix = "/rbac/v1/rolebindings/"
)

type RBACSystem struct {
	etcdClient   *clientv3.Client
	roles        map[string]rbac.Role
	roleBindings map[string]rbac.RoleBinding
	mu           sync.RWMutex
}

// NewRBACSystem initializes a new RBAC system with an etcd client.
//
//	type RBACSystem struct {
//			etcdClient 		*clientv3.Client
//			roles      		map[string]rbac.Role
//			roleBindings   	map[string]rbac.RoleBinding
//			mu         		sync.RWMutex
//		}
func NewRBACSystem(etcdClient *clientv3.Client) *RBACSystem {
	return &RBACSystem{
		etcdClient:   etcdClient,
		roles:        make(map[string]rbac.Role),
		roleBindings: make(map[string]rbac.RoleBinding, 100),
		mu:           sync.RWMutex{},
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
		r.roleBindings[rolebindingKey] = rolebinding
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

// generateKey generates a key which will be stored in RBACSystem.roles or RBACSystem.roleBindings.
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
		ErrorResponse(c, http.StatusBadRequest, "Invalid role data", err)
		return
	}
	role.SetUID(uuid.NewString())
	role.SetCreationTimestamp(time.Now())

	roleKey := r.generateKey(role.Namespace, role.Name)

	r.mu.Lock()
	defer r.mu.Unlock()

	if role, exists := r.roles[roleKey]; exists {
		ErrorResponse(c, http.StatusConflict, fmt.Sprintf("Role %s already exists", role), errors.New("role already exists"))
		return
	}

	etcdKey := r.generateEtcdKey(rolePrefix, role.Namespace, role.Name)
	if err := r.saveToEtcd(etcdKey, role); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to save role", err)
		return
	}

	// avoid conflicts when saveToEtcd is unsuccessful
	r.roles[roleKey] = role

	SuccessResponse(c, http.StatusCreated, "Role created successfully", role)
}

func (r *RBACSystem) HandleCreateRoleBinding(c *gin.Context) {
	var rolebinding rbac.RoleBinding
	if err := c.ShouldBindJSON(&rolebinding); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid role binding data", err)
		return
	}
	rolebinding.SetUID(uuid.NewString())
	rolebinding.SetCreationTimestamp(time.Now())

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if the role exists before creating the rolebinding
	roleKey := r.generateKey(rolebinding.Namespace, rolebinding.RoleRef.Name)
	if role, exists := r.roles[roleKey]; !exists {
		ErrorResponse(c, http.StatusNotFound, fmt.Sprintf("Role %s not found", role), errors.New("role not found"))
		return
	}

	rolebindingkey := r.generateKey(rolebinding.Namespace, rolebinding.Name)
	if roleBinding, exists := r.roleBindings[rolebindingkey]; exists {
		ErrorResponse(c, http.StatusConflict, fmt.Sprintf("Rolebinding %s already exists", roleBinding), errors.New("rolebinding already exists"))
		return
	}

	etcdKey := r.generateEtcdKey(rolebindingPrefix, rolebinding.Namespace, rolebinding.Name)
	if err := r.saveToEtcd(etcdKey, rolebinding); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to save role binding", err)
		return
	}

	// avoid conflicts when saveToEtcd is unsuccessful
	r.roleBindings[rolebindingkey] = rolebinding

	SuccessResponse(c, http.StatusCreated, "Role binding created successfully", rolebinding)
}

// HandleListRoles is a list function for retrieving all roles.
func (r *RBACSystem) HandleListRoles(c *gin.Context) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roles := make([]rbac.Role, 0, len(r.roles))
	for _, role := range r.roles {
		roles = append(roles, role)
	}

	SuccessResponse(c, http.StatusOK, "Roles retrieved successfully", roles)
}

// HandleListRoleBindings is a list function for retrieving all role bindings.
func (r *RBACSystem) HandleListRoleBindings(c *gin.Context) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roleBindings := make([]rbac.RoleBinding, 0, len(r.roleBindings))
	for _, rolebinding := range r.roleBindings {
		roleBindings = append(roleBindings, rolebinding)
	}

	SuccessResponse(c, http.StatusOK, "Role bindings retrieved successfully", roleBindings)
}

func (r *RBACSystem) HandleHealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_, err := r.etcdClient.Get(ctx, "health-check-key")
	cancel()
	if err == nil || err == rpctypes.ErrPermissionDenied {
		// 连接正常（即使权限错误也说明连接ok）
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "rbac system is healthy",
		})
	}

	c.JSON(500, gin.H{
		"status":  "error",
		"message": "rbac system is not healthy",
	})
}

// InitRegister method create default roles and roleBidings
// when rbac system is initialized
func (r *RBACSystem) InitRegister() {
	roles := []rbac.Role{
		// admin is a super user, it has all permissions for any resource
		rbac.Role{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac/v1",
				Kind:       "Role",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              "admin",
				Namespace:         "default",
				UID:               uuid.NewString(),
				CreationTimestamp: time.Now(),
			},
			Rules: []rbac.PolicyRule{
				rbac.PolicyRule{
					APIGroups: []string{""},
					Resources: []string{"*"},
					Verbs:     []string{"*"},
				},
			},
		},
		// view role is a read-only role, it can only get all URLs, roles and rolebindings
		rbac.Role{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac/v1",
				Kind:       "Role",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              "view",
				Namespace:         "default",
				UID:               uuid.NewString(),
				CreationTimestamp: time.Now(),
			},
			Rules: []rbac.PolicyRule{
				rbac.PolicyRule{
					APIGroups: []string{""},
					Resources: []string{"*"},
					Verbs:     []string{"get"},
				},
			},
		},
		// edit role is a read-write role, it can create, get, delete any resource
		// except roles and rolebindings
		rbac.Role{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac/v1",
				Kind:       "Role",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              "edit",
				Namespace:         "default",
				UID:               uuid.NewString(),
				CreationTimestamp: time.Now(),
			},
			Rules: []rbac.PolicyRule{
				rbac.PolicyRule{
					APIGroups: []string{""},
					Resources: []string{"urls", "users"},
					Verbs:     []string{"get", "create", "delete", "list"},
				},
			},
		},
		// user is a commom role, it only can read URLs
		rbac.Role{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac/v1",
				Kind:       "Role",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              "user",
				Namespace:         "default",
				UID:               uuid.NewString(),
				CreationTimestamp: time.Now(),
			},
			Rules: []rbac.PolicyRule{
				rbac.PolicyRule{
					APIGroups: []string{""},
					Resources: []string{"urls"},
					Verbs:     []string{"get", "list"},
				},
			},
		},
	}

	roleBindings := []rbac.RoleBinding{
		rbac.RoleBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac/v1",
				Kind:       "RoleBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              "admin-binding",
				Namespace:         "default",
				UID:               uuid.NewString(),
				CreationTimestamp: time.Now(),
			},
			RoleRef: rbac.RoleRef{
				APIGroup: "rbac/v1",
				Kind:     "Role",
				Name:     "admin",
			},
			Subjects: []rbac.Subject{
				rbac.Subject{
					Kind: "User",
					Name: "admin",
				},
			},
		},
		rbac.RoleBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac/v1",
				Kind:       "RoleBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              "view-binding",
				Namespace:         "default",
				UID:               uuid.NewString(),
				CreationTimestamp: time.Now(),
			},
			RoleRef: rbac.RoleRef{
				APIGroup: "rbac/v1",
				Kind:     "Role",
				Name:     "view",
			},
			Subjects: []rbac.Subject{
				rbac.Subject{
					Kind: "User",
					Name: "view",
				},
			},
		},
		rbac.RoleBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac/v1",
				Kind:       "RoleBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              "edit-binding",
				Namespace:         "default",
				UID:               uuid.NewString(),
				CreationTimestamp: time.Now(),
			},
			RoleRef: rbac.RoleRef{
				APIGroup: "rbac/v1",
				Kind:     "Role",
				Name:     "edit",
			},
			Subjects: []rbac.Subject{
				rbac.Subject{
					Kind: "User",
					Name: "edit",
				},
			},
		},
		rbac.RoleBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac/v1",
				Kind:       "RoleBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              "user-binding",
				Namespace:         "default",
				UID:               uuid.NewString(),
				CreationTimestamp: time.Now(),
			},
			RoleRef: rbac.RoleRef{
				APIGroup: "rbac/v1",
				Kind:     "Role",
				Name:     "user",
			},
			Subjects: []rbac.Subject{
				rbac.Subject{
					Kind: "User",
					Name: "user",
				},
			},
		},
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, role := range roles {
		etcdKey := r.generateEtcdKey(rolePrefix, role.Namespace, role.Name)
		if err := r.saveToEtcd(etcdKey, role); err != nil {
			log.Err(err).Msg("failed to save role to etcd")
			return
		}
		roleKey := r.generateKey(role.Namespace, role.Name)
		r.roles[roleKey] = role
	}

	for _, roleBinding := range roleBindings {
		etcdKey := r.generateEtcdKey(rolebindingPrefix, roleBinding.Namespace, roleBinding.Name)
		if err := r.saveToEtcd(etcdKey, roleBinding); err != nil {
			log.Err(err).Msg("failed to save rolebinding to etcd")
			return
		}
		roleBindingKey := r.generateKey(roleBinding.Namespace, roleBinding.Name)
		r.roleBindings[roleBindingKey] = roleBinding
	}
	log.Info().Msg("Loaded etcd initial data.")
}
