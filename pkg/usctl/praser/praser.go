// Praser prases TypeMeta from yaml data,
// then register use this message to create a specific resource
// and uses serializer to serialize it to JSON.
package praser

import (
	"encoding/json"
	"fmt"
	metav1 "url-shortener/pkg/apis/meta/v1"
	"url-shortener/pkg/usctl"
	"url-shortener/pkg/usctl/register"

	"gopkg.in/yaml.v3"
)

type ResourceParser struct{}

// Parse parses the given YAML data into a specific resource type based on its TypeMeta.
func (p *ResourceParser) Parse(data []byte, marshalType string) (name string, jsonData []byte, printData []byte, err error) {
	typemeta := metav1.TypeMeta{}
	if err := yaml.Unmarshal(data, &typemeta); err != nil {
		return "", nil, nil, usctl.ErrInvalidMeta
	}

	register := register.NewRegister(typemeta, data)
	name, resource, err := register.Register()
	if err != nil {
		return "", nil, nil, err
	}

	jsonData, err = json.Marshal(resource)
	if err != nil {
		return "", nil, nil, usctl.ErrSerialized
	}

	if marshalType == "yaml" {
		printData, err = yaml.Marshal(resource)
	}

	return fmt.Sprintf("%s %s", typemeta.Kind, name), jsonData, printData, err
}
