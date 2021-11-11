package templates

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

var (
	//go:embed jira-down.template
	jiraDown string
	//go:embed mark-down.template
	markDown string
	//go:embed text.template
	textDown string
)

type Format int

const (
	Markdown Format = iota
	Jira
	Text
)

type OutputTemplate struct {
	Type           Format
	Template       string
	ParsedTemplate string
}

func getTemplate(name string) (out OutputTemplate) {
	sanitized := strings.ToLower(strings.TrimSpace(name))
	switch sanitized {
	case "markdown", "mark":
		out.Template = markDown
		out.Type = Markdown
	case "jiradown", "jira":
		out.Template = jiraDown
		out.Type = Jira
	default:
		out.Template = textDown
		out.Type = Text
	}
	return
}

func ParseWithTemplateFile(templateFile string, data interface{}) (OutputTemplate, error) {
	o := OutputTemplate{}
	if strings.ToLower(filepath.Ext(templateFile)) != ".md" {
		o.Type = Text
	}

	templateData, err := os.ReadFile(templateFile)
	if err != nil {
		return o, err
	}

	o.Template = string(templateData)
	return parse(o, data)
}

func Parse(templateFormat string, data interface{}) (OutputTemplate, error) {
	o := getTemplate(templateFormat)
	return parse(o, data)
}

func parse(templateFormat OutputTemplate, data interface{}) (OutputTemplate, error) {
	t, err := template.New("temp").Funcs(template.FuncMap{
		"diff": PercentageChange,
	}).Parse(templateFormat.Template)
	if err != nil {
		return templateFormat, err
	}

	buf := &bytes.Buffer{}
	if err = t.Execute(buf, data); err != nil {
		return templateFormat, err
	}
	templateFormat.ParsedTemplate = buf.String()
	return templateFormat, nil
}

func PercentageChange(o, n string) string {
	old, _ := strconv.ParseFloat(o, 64)
	new, _ := strconv.ParseFloat(n, 64)
	diff := float64(new - old)
	delta := (diff / float64(old)) * 100
	if delta > 0 {
		return fmt.Sprintf("*%0.2f%%*", delta)
	}
	return fmt.Sprintf("`%0.2f%%`", delta)
}
