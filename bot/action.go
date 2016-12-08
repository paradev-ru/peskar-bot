package bot

import (
	"bytes"
	"text/template"
)

type Action struct {
	ChatId   string `toml:"chat_id"`
	Message  string `toml:"message"`
	JobState string `toml:"job_state"`
}

func (a *Action) Template(object interface{}) (string, error) {
	var doc bytes.Buffer
	t, err := template.New("text").Parse(a.Message)
	if err != nil {
		return "", err
	}
	err = t.Execute(&doc, object)
	if err != nil {
		return "", err
	}
	return doc.String(), nil
}
