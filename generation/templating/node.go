package templating

import (
	"errors"
	"fmt"
	"strings"
)

type Marker struct {
	Name       string
	Attributes map[string]string
}

func (m *Marker) ToString(builder *strings.Builder, needsAttrs bool) {
	builder.WriteString("/*")
	builder.WriteString("marker:")
	builder.WriteString(m.Name)
	if needsAttrs {
		for name, value := range m.Attributes {
			builder.WriteString(fmt.Sprintf(", %s=%s", name, value))
		}
	}
	builder.WriteString("*/")
}

func (m *Marker) Parse(source string) error {
	if !strings.HasPrefix(source, "/*marker:") {
		return errors.New("malformed marker")
	}
	end := strings.Index(source, "*/")
	if end == -1 {
		return errors.New("unclosed marker")
	}

	//entries := strings.Split(source[len("/*marker:"):end], ",")

	return nil
}

type iTextNode interface {
	Render(data map[string]any, preserveMarkers bool) string
}

type plainNode struct {
	value string
}

func (node *plainNode) Render(_ map[string]any, _ bool) string {
	return node.value
}

type insertNode struct {
	Value  []iTextNode
	Marker Marker
}

func (node *insertNode) Render(data map[string]any, preserveMarkers bool) string {
	builder := strings.Builder{}

	if preserveMarkers {
		node.Marker.ToString(&builder, true)
	}

	for _, textNode := range node.Value {
		builder.WriteString(textNode.Render(data, preserveMarkers))
	}

	if preserveMarkers {
		node.Marker.ToString(&builder, false)
	}

	return builder.String()
}

type forNode struct {
	Children []iTextNode
	Marker   Marker
}

func extendMap[K comparable, V any](data map[K]V, key K, value V) map[K]V {
	mapCopy := make(map[K]V)
	for k, v := range data {
		mapCopy[k] = v
	}
	mapCopy[key] = value
	return mapCopy
}

func (node *forNode) Render(data map[string]any, preserveMarkers bool) string {
	builder := strings.Builder{}

	if preserveMarkers {
		node.Marker.ToString(&builder, true)
	}

	iterName := node.Marker.Attributes["iter"]

	collection := data[node.Marker.Name].([]any)

	for _, element := range collection {
		extended := extendMap(data, iterName, element)
		for _, child := range node.Children {
			builder.WriteString(child.Render(extended, preserveMarkers))
		}
	}

	if preserveMarkers {
		node.Marker.ToString(&builder, false)
	}

	return builder.String()
}
