package v1

import (
	"fmt"
	"strings"
	"url-shortener/pkg/apis/rbac"
)

func formatRules(rules []rbac.PolicyRule) string {
	var result []string
	for _, rule := range rules {
		resources := "*"
		if len(rule.Resources) > 0 {
			resources = strings.Join(rule.Resources, ",")
		}

		verbs := "*"
		if len(rule.Verbs) > 0 {
			verbs = strings.Join(rule.Verbs, ",")
		}

		result = append(result, fmt.Sprintf("%s on %s", verbs, resources))
	}
	return strings.Join(result, "; ")
}
