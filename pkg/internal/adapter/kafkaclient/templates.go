package kafkaclient

import (
	"encoding/json"
	"fmt"
	"strings"
)

func renderFieldFromValue[T any](v T, placeholder string) (string, bool) {
	ph := strings.TrimSpace(placeholder)
	if !strings.HasPrefix(ph, "{") || !strings.HasSuffix(ph, "}") {
		return "", false
	}
	field := strings.TrimSpace(ph[1 : len(ph)-1])
	if field == "" {
		return "", false
	}

	b, err := json.Marshal(v)
	if err != nil {
		return "", false
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return "", false
	}
	if raw, ok := m[field]; ok && raw != nil {
		return fmt.Sprint(raw), true
	}
	return "", false
}

func renderKeyFromTemplate[T any](tmpl string, v T) []byte {
	t := strings.TrimSpace(tmpl)
	if t == "" {
		return nil
	}
	if val, ok := renderFieldFromValue(v, t); ok {
		return []byte(val)
	}
	return []byte(t)
}

func renderHeadersFromTemplates[T any](tmpls map[string]string, v T) []struct{ Key, Value string } {
	if len(tmpls) == 0 {
		return nil
	}
	out := make([]struct{ Key, Value string }, 0, len(tmpls))
	for k, t := range tmpls {
		val := t
		if vv, ok := renderFieldFromValue(v, t); ok {
			val = vv
		}
		out = append(out, struct{ Key, Value string }{Key: k, Value: val})
	}
	return out
}
