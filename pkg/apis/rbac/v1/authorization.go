package v1

import (
	"net/http"
	"url-shortener/pkg/apis/rbac"

	"github.com/gin-gonic/gin"
)

type AuthRequest struct {
	Name      string `json:"name" validate:"required"`
	Verb      string `json:"verb" validate:"required"`
	Resource  string `json:"resource" validate:"required"`
	Namespace string `json:"namespace" validate:"required"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func SuccessResponse(c *gin.Context, status int, message string, data any) {
	c.JSON(status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, status int, message string, err error) {
	c.JSON(status, APIResponse{
		Success: false,
		Message: message,
		Error:   err.Error(),
	})
}

func (r *RBACSystem) HandleRBACAuthCheck(c *gin.Context) {
	var authReq AuthRequest
	if err := c.ShouldBindJSON(&authReq); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	allowed, err := r.Authorize(authReq)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "authorization failed", err)
		return
	}

	result := map[string]bool{"authorized": allowed}
	if allowed {
		SuccessResponse(c, http.StatusOK, "authorization successful", result)
	} else {
		ErrorResponse(c, http.StatusForbidden, "authorization failed", nil)
	}
}

func (r *RBACSystem) Authorize(authReq AuthRequest) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roles := r.findRolesForUser(authReq.Name, authReq.Namespace)
	if len(roles) == 0 {
		return false, nil
	}

	for _, role := range roles {
		if r.roleHasPermission(role, authReq.Verb, authReq.Resource) {
			return true, nil
		}
	}

	return false, nil
}

func (r *RBACSystem) findRolesForUser(name, namespace string) []rbac.Role {
	var roles []rbac.Role

	for _, rolebinding := range r.roleBindings {
		if rolebinding.Namespace != namespace {
			continue
		}

		for _, subject := range rolebinding.Subjects {
			// Only check subject.Name.
			// Kind checking is done in Authorize method.
			if subject.Name == name {
				roleKey := r.generateKey(namespace, rolebinding.RoleRef.Name)
				if role, exists := r.roles[roleKey]; exists {
					roles = append(roles, role)
				}
				break
			}
		}
	}

	return roles
}

func (r *RBACSystem) roleHasPermission(role rbac.Role, verb, resource string) bool {
	for _, rule := range role.Rules {
		if ruleMatches(rule, verb, resource) {
			return true
		}
	}
	return false
}

func ruleMatches(rule rbac.PolicyRule, verb, resource string) bool {
	if contains(rule.Verbs, verb) && contains(rule.Resources, resource) {
		return true
	}

	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == "*" || s == item {
			return true
		}
	}
	return false
}
