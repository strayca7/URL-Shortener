package register

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	metav1 "url-shortener/pkg/apis/meta/v1"
	rbac "url-shortener/pkg/apis/rbac"
	"url-shortener/pkg/usctl"

	"gopkg.in/yaml.v3"
)

const (
	httpUrlHead  = "http://127.0.0.1:8080/"
	httpsUrlHead = "https://127.0.0.1:8080/"

	roleAPI        = "rbac/v1/role"
	roleBindingAPI = "rbac/v1/rolebinding"
)

type Register struct {
	_type metav1.TypeMeta
	// file data
	data    []byte
	client  *http.Client
	requset *http.Request
}

func NewRegister(_type metav1.TypeMeta, data []byte) *Register {
	return &Register{
		_type: _type,
		data:  data,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		requset: &http.Request{},
	}
}

// Register creates a resource based on the TypeMeta information.
// It will return the resource's name, serialized resource list and an error.
// It only checks the necessary fields for the resource type.
// More complex validation and registration logic will be handled in API Server.
func (r *Register) Register() (name string, resource any, err error) {
	if r._type.APIVersion == "" || r._type.Kind == "" {
		return "", nil, usctl.ErrInvalidMeta
	}

	// here error all returns usctl.ErrRegiterKind
	switch r._type.Kind {
	case "Role":
		role, err := r.registerRole()
		return role.Name, role, err
	case "RoleBinding":
		roleBinding, err := r.registerRoleBinding()
		return roleBinding.Name, roleBinding, err
	default:
		return "", nil, usctl.ErrRegiterKind // TODO: handle unsupported types
	}
}

// registerRole is a method of Register.
// It will only check two fileds APIVersion and Namespace.
// More fields will be check in API Server.
// If the yaml data can not bind to a rbac.Role struct, It will return an error in terminal.
func (r *Register) registerRole() (*rbac.Role, error) {
	role := &rbac.Role{}

	if err := yaml.Unmarshal(r.data, role); err != nil {
		return nil, usctl.ErrRegiterKind
	}

	role.SetAPIVersion("rbac/v1")
	if role.Namespace == "" {
		role.SetNamespace("default")
	}

	jsonData, err := json.Marshal(role)
	if err != nil {
		return nil, usctl.ErrRegiterKind
	}
	status := r.Requset("POST", httpUrlHead+roleAPI, jsonData)
	if status != http.StatusCreated {
		switch status {
		case http.StatusBadRequest:
			fmt.Fprintln(os.Stderr, "invalid role")
		case http.StatusConflict:
			fmt.Fprintln(os.Stderr, "role already exists")
		case http.StatusInternalServerError:
			fmt.Fprintln(os.Stderr, "internal server error")
		default:
			fmt.Fprintln(os.Stderr, "register role error, please check your role yaml file")
		}
		os.Exit(1)
	}

	return role, nil
}

func (r *Register) registerRoleBinding() (*rbac.RoleBinding, error) {
	roleBinding := &rbac.RoleBinding{}

	if err := yaml.Unmarshal(r.data, roleBinding); err != nil {
		return nil, usctl.ErrRegiterKind
	}

	roleBinding.SetAPIVersion("rbac/v1")
	if roleBinding.Namespace == "" {
		roleBinding.SetNamespace("default")
	}

	jsonData, err := json.Marshal(roleBinding)
	if err != nil {
		return nil, usctl.ErrRegiterKind
	}
	status := r.Requset("POST", httpUrlHead+roleBindingAPI, jsonData)
	if status != http.StatusCreated {
		switch status {
		case http.StatusBadRequest:
			fmt.Fprintln(os.Stderr, "invalid roleBinding")
		case http.StatusConflict:
			fmt.Fprintln(os.Stderr, "roleBinding already exists")
		case http.StatusInternalServerError:
			fmt.Fprintln(os.Stderr, "internal server error")
		default:
			fmt.Fprintln(os.Stderr, "register roleBinding error, please check your roleBinding yaml file")
		}
		os.Exit(1)
	}

	return roleBinding, nil
}

func (r *Register) Requset(method string, url string, jsonData []byte) int {
	var err error
	r.requset, err = http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	r.requset.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(r.requset)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}
