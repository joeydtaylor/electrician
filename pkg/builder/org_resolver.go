package builder

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// OrgResolverConfig controls how org IDs are discovered from env and JWTs.
type OrgResolverConfig struct {
	DefaultOrgID string
	OrgIDEnvName string
	AssertJWTEnv string
	BearerJWTEnv string
	ClaimKeys    []string
}

// DefaultOrgResolverConfig returns a config with standard env and claim keys.
func DefaultOrgResolverConfig(defaultOrgID string) OrgResolverConfig {
	return OrgResolverConfig{
		DefaultOrgID: defaultOrgID,
		OrgIDEnvName: "ORG_ID",
		AssertJWTEnv: "ASSERT_JWT",
		BearerJWTEnv: "BEARER_JWT",
		ClaimKeys:    []string{"org", "org_id", "orgId", "organization"},
	}
}

// ResolveOrgIDFromEnv checks ORG_ID, then JWTs (ASSERT_JWT, BEARER_JWT), then default.
func ResolveOrgIDFromEnv(cfg OrgResolverConfig) string {
	if cfg.OrgIDEnvName == "" {
		cfg.OrgIDEnvName = "ORG_ID"
	}
	if cfg.AssertJWTEnv == "" {
		cfg.AssertJWTEnv = "ASSERT_JWT"
	}
	if cfg.BearerJWTEnv == "" {
		cfg.BearerJWTEnv = "BEARER_JWT"
	}
	if len(cfg.ClaimKeys) == 0 {
		cfg.ClaimKeys = []string{"org", "org_id", "orgId", "organization"}
	}

	if v := EnvOr(cfg.OrgIDEnvName, ""); v != "" {
		return v
	}
	if v := EnvOr(cfg.AssertJWTEnv, ""); v != "" {
		if org, ok := OrgIDFromJWT(v, cfg.ClaimKeys); ok {
			return org
		}
	}
	if v := EnvOr(cfg.BearerJWTEnv, ""); v != "" {
		if org, ok := OrgIDFromJWT(v, cfg.ClaimKeys); ok {
			return org
		}
	}
	return cfg.DefaultOrgID
}

// OrgIDFromJWT extracts an org-like claim from a JWT payload without verifying.
func OrgIDFromJWT(token string, claimKeys []string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return "", false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", false
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", false
	}
	for _, k := range claimKeys {
		if v, ok := claims[k]; ok && v != nil {
			s := strings.TrimSpace(fmt.Sprint(v))
			if s != "" {
				return s, true
			}
		}
	}
	return "", false
}
