package builder

import (
	"os"
	"strconv"
	"strings"
)

// EnvOr returns the trimmed env value or def when empty.
func EnvOr(key, def string) string {
	v := strings.TrimSpace(strings.Trim(os.Getenv(key), `"`))
	if v == "" {
		return def
	}
	return v
}

// EnvIntOr returns the parsed int env value or def on empty/parse failure.
func EnvIntOr(key string, def int) int {
	v := EnvOr(key, "")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
