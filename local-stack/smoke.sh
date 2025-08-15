#!/usr/bin/env bash
set -euo pipefail

# Base creds for LocalStack bootstrap
export AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID:-test}
export AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY:-test}
export AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION:-us-east-1}
EP=${EP:-http://localhost:4566}

ROLE_ARN="arn:aws:iam::000000000000:role/exodus-dev-role"
BUCKET="steeze-dev"

# Sanity: role & bucket exist
aws --endpoint-url "$EP" iam get-role --role-name exodus-dev-role >/dev/null
aws --endpoint-url "$EP" s3 ls "s3://$BUCKET" >/dev/null 2>&1 || aws --endpoint-url "$EP" s3 mb "s3://$BUCKET"

# Assume role and extract creds via --query
read AK SK TOK <<<"$(aws --endpoint-url "$EP" sts assume-role \
  --role-arn "$ROLE_ARN" \
  --role-session-name exodus-smoke \
  --query 'Credentials.[AccessKeyId,SecretAccessKey,SessionToken]' \
  --output text)"

# Ensure we actually got values
if [ -z "${AK:-}" ] || [ -z "${SK:-}" ] || [ -z "${TOK:-}" ]; then
  echo "assume-role failed: empty credentials"
  exit 1
fi

# Export session creds
export AWS_ACCESS_KEY_ID="$AK"
export AWS_SECRET_ACCESS_KEY="$SK"
export AWS_SESSION_TOKEN="$TOK"

# Write & read test object
KEY="assume-role-smoke.txt"
printf 'ok\n' | aws --endpoint-url "$EP" s3 cp - "s3://$BUCKET/$KEY"
aws --endpoint-url "$EP" s3 cp "s3://$BUCKET/$KEY" -

echo "[smoke] success"
