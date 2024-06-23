package templating

import (
	"bytes"
	"text/template"
)

func PrepareTeamplate(pattern string, varsTemplate map[string]interface{}) (string, error) {
	tmpl, err := template.New("person_template").Parse(pattern)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, varsTemplate)
	if err != nil {
		return "", err
	}

	patternTemplate := buf.String()
	return patternTemplate, nil
}
