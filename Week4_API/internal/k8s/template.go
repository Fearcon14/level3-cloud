package k8s

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// RenderRedisFailoverTemplate reads the template file, executes it with data, and returns the rendered YAML.
func RenderRedisFailoverTemplate(templatePath string, data RedisFailoverTemplateData) ([]byte, error) {
	tplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("read template %q: %w", templatePath, err)
	}

	tpl, err := template.New("redis-failover").Parse(string(tplBytes))
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// DecodeYAMLToUnstructured decodes YAML bytes into an unstructured Kubernetes object.
// The result can be passed to the dynamic client for create/update.
func DecodeYAMLToUnstructured(yamlBytes []byte) (*unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(yamlBytes, &obj.Object); err != nil {
		return nil, fmt.Errorf("unmarshal yaml to unstructured: %w", err)
	}
	
	return obj, nil
}
