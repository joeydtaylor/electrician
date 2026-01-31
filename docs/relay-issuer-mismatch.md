# JWT Issuer Mismatch

If the relay logs or responds with:

```
[unauthenticated] auth validation failed: issuer mismatch
```

Your token's `iss` claim does not match the receiver's configured issuer.

## How to fix
1) Decode the token and read `iss`.
2) Configure the receiver to use the same value.

### Default secure gRPC-web receiver
`example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb`

Defaults to:
- `OAUTH_ISSUER_BASE = auth-service`

### If your token has a different issuer
Example: token has `iss = https://auth.example.com`

```bash
export OAUTH_ISSUER_BASE=https://auth.example.com
```

Also ensure JWKS points to the same auth system:

```bash
export OAUTH_JWKS_URL=https://auth.example.com/api/auth/oauth/jwks.json
```

## Related checks
- `aud` must include `your-api`
- `scope` must include `write:data`
- TLS must be enabled for the secure examples
