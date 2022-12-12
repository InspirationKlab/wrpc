package templating

import "strings"

type TemplateData struct {
	Entries map[string]any
}

func (data *TemplateData) ExtendWith(name string, value any) *TemplateData {
	resultMap := make(map[string]any)
	for eName, eValue := range data.Entries {
		resultMap[eName] = eValue
	}
	resultMap[name] = value
	return &TemplateData{
		resultMap,
	}
}

type templateNode interface {
	Render(data *TemplateData) string
}

type templateText struct {
	text string
}

func (templateText *templateText) Render(data *TemplateData) string {
	return templateText.text
}

type templateMarkerInsert struct {
	name string
}

func (templateMarkerInsert *templateMarkerInsert) Render(data *TemplateData) string {
	return data.Entries[templateMarkerInsert.name].(string)
}

type ParsedTemplate struct {
	nodes []templateNode
}

func (tmpl *ParsedTemplate) Render(data *TemplateData) string {
	builder := strings.Builder{}

	for _, node := range tmpl.nodes {
		builder.WriteString(node.Render(data))
	}

	return builder.String()
}