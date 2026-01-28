package auth

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

func validateStandardClaims(claims map[string]interface{}, issuer string, audiences []string) error {
	now := time.Now().Unix()
	const skewSeconds int64 = 5

	if issuer != "" {
		iss, _ := claims["iss"].(string)
		if iss == "" || iss != issuer {
			return errors.New("issuer mismatch")
		}
	}

	if len(audiences) > 0 {
		got := extractAudience(claims)
		if !audienceMatch(got, audiences) {
			return errors.New("audience mismatch")
		}
	}

	if exp, ok := claimInt64(claims, "exp"); ok {
		if now > exp+skewSeconds {
			return errors.New("token expired")
		}
	}
	if nbf, ok := claimInt64(claims, "nbf"); ok {
		if now+skewSeconds < nbf {
			return errors.New("token not yet valid")
		}
	}
	return nil
}

func extractAudience(claims map[string]interface{}) []string {
	v, ok := claims["aud"]
	if !ok {
		return nil
	}
	switch t := v.(type) {
	case string:
		if t == "" {
			return nil
		}
		return []string{t}
	case []string:
		return cloneStrings(t)
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok {
				if trimmed := strings.TrimSpace(s); trimmed != "" {
					out = append(out, trimmed)
				}
			}
		}
		return out
	default:
		return nil
	}
}

func audienceMatch(got, required []string) bool {
	if len(required) == 0 {
		return true
	}
	if len(got) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(got))
	for _, a := range got {
		set[a] = struct{}{}
	}
	for _, need := range required {
		if _, ok := set[need]; ok {
			return true
		}
	}
	return false
}

func claimInt64(claims map[string]interface{}, key string) (int64, bool) {
	v, ok := claims[key]
	if !ok {
		return 0, false
	}
	switch t := v.(type) {
	case json.Number:
		i, err := t.Int64()
		return i, err == nil
	case float64:
		return int64(t), true
	case int64:
		return t, true
	case int:
		return int64(t), true
	case string:
		if t == "" {
			return 0, false
		}
		var n json.Number = json.Number(t)
		i, err := n.Int64()
		return i, err == nil
	default:
		return 0, false
	}
}

func hasAllScopes(required []string, claims map[string]interface{}) bool {
	if len(required) == 0 {
		return true
	}
	granted := extractScopes(claims)
	if len(granted) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(granted))
	for _, s := range granted {
		set[s] = struct{}{}
	}
	for _, req := range required {
		if _, ok := set[req]; !ok {
			return false
		}
	}
	return true
}

func extractScopes(claims map[string]interface{}) []string {
	for _, key := range []string{"scope", "scp", "scopes"} {
		if v, ok := claims[key]; ok {
			if scopes := parseScopeValue(v); len(scopes) > 0 {
				return scopes
			}
		}
	}
	return nil
}

func parseScopeValue(v interface{}) []string {
	switch t := v.(type) {
	case string:
		return strings.Fields(t)
	case []string:
		return cloneStrings(t)
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok {
				if trimmed := strings.TrimSpace(s); trimmed != "" {
					out = append(out, trimmed)
				}
			}
		}
		return out
	default:
		return nil
	}
}
