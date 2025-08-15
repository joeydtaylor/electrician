#!/usr/bin/env bash
set -euo pipefail

REGION=us-east-1
ALIAS_NAME="alias/electrician-dev"
ROLE_NAME="exodus-dev-role"

# Create a CMK and capture KeyId without jq
KEY_ID="$(awslocal kms create-key \
  --region "${REGION}" \
  --description "Electrician dev key" \
  --query 'KeyMetadata.KeyId' \
  --output text)"

# Idempotent alias create
EXISTING_ALIAS="$(awslocal kms list-aliases \
  --region "${REGION}" \
  --query "Aliases[?AliasName=='${ALIAS_NAME}'].AliasName" \
  --output text || true)"

if [[ -z "${EXISTING_ALIAS}" ]]; then
  awslocal kms create-alias \
    --region "${REGION}" \
    --alias-name "${ALIAS_NAME}" \
    --target-key-id "${KEY_ID}"
else
  # Optional: point alias to latest key if you recreate keys often
  awslocal kms update-alias \
    --region "${REGION}" \
    --alias-name "${ALIAS_NAME}" \
    --target-key-id "${KEY_ID}" || true
fi

# Ensure IAM role exists
if ! awslocal iam get-role --role-name "${ROLE_NAME}" >/dev/null 2>&1; then
  awslocal iam create-role \
    --role-name "${ROLE_NAME}" \
    --assume-role-policy-document '{
      "Version":"2012-10-17",
      "Statement":[{"Effect":"Allow","Principal":{"AWS":"*"},"Action":"sts:AssumeRole"}]
    }' >/dev/null
fi

# Grant broad KMS perms to the role (fine for LocalStack)
awslocal iam put-role-policy \
  --role-name "${ROLE_NAME}" \
  --policy-name "electrician-kms-access" \
  --policy-document '{
    "Version":"2012-10-17",
    "Statement":[
      {"Effect":"Allow","Action":["kms:*"],"Resource":"*"}
    ]
  }' >/dev/null

echo "KMS KeyId: ${KEY_ID}"
echo "KMS Alias: ${ALIAS_NAME}"
