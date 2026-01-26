package internallogger

import "go.uber.org/zap"

func fieldsFromMap(fields map[string]interface{}) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	out := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		if key == "" {
			continue
		}
		out = append(out, zap.Any(key, value))
	}
	return out
}
