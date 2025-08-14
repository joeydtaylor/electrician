#!/usr/bin/env bash
set -euo pipefail

# ---------- inputs (override via env if you want) ----------
ORG_ID="${ORG_ID:-4d948fa0-084e-490b-aad5-cfd01eeab79a}"

BUCKET="${BUCKET:-steeze-dev}"
ROLE_NAME="${ROLE_NAME:-exodus-dev-role}"
ROLE_POLICY_NAME="${ROLE_POLICY_NAME:-exodus-dev-s3}"

# Create a KMS key for SSE-KMS? (1=yes, 0=no)
CREATE_KMS="${CREATE_KMS:-1}"
KMS_ALIAS="${KMS_ALIAS:-alias/electrician-dev}"   # match your app’s config/alias

# A few “datasets” to pre-create under your org (purely cosmetic; S3 has no real folders)
DATASETS=("feedback/demo" "dlq" "audit/events")

# ---------- derived ----------
PREFIX_BASE="org=${ORG_ID}"

echo "[init] provisioning IAM role + S3 bucket (org=${ORG_ID})"

# ---------- Trust policy (dev-friendly; anyone can assume) ----------
cat >/tmp/trust.json <<'JSON'
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": { "AWS": "*" },
    "Action": "sts:AssumeRole"
  }]
}
JSON

# ---------- S3 policy: scope to your org prefix ----------
cat >/tmp/s3-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ListOwnPrefix",
      "Effect": "Allow",
      "Action": ["s3:ListBucket"],
      "Resource": ["arn:aws:s3:::${BUCKET}"],
      "Condition": {
        "StringLike": { "s3:prefix": ["${PREFIX_BASE}/*"] }
      }
    },
    {
      "Sid": "RWObjectsUnderPrefix",
      "Effect": "Allow",
      "Action": [
        "s3:GetObject","s3:PutObject","s3:DeleteObject",
        "s3:AbortMultipartUpload","s3:ListMultipartUploadParts"
      ],
      "Resource": ["arn:aws:s3:::${BUCKET}/${PREFIX_BASE}/*"]
    }
  ]
}
EOF

# ---------- idempotent role ----------
if ! awslocal iam get-role --role-name "$ROLE_NAME" >/dev/null 2>&1; then
  awslocal iam create-role \
    --role-name "$ROLE_NAME" \
    --assume-role-policy-document file:///tmp/trust.json >/dev/null
fi

# (re)attach inline S3 policy
awslocal iam put-role-policy \
  --role-name "$ROLE_NAME" \
  --policy-name "$ROLE_POLICY_NAME" \
  --policy-document file:///tmp/s3-policy.json >/dev/null

# ---------- bucket ----------
awslocal s3 ls "s3://${BUCKET}" >/dev/null 2>&1 || awslocal s3 mb "s3://${BUCKET}" >/dev/null

# ---------- optional KMS key + alias, plus permissive KMS policy for the role ----------
if [[ "${CREATE_KMS}" == "1" ]]; then
  KEY_ID="$(awslocal kms create-key --query 'KeyMetadata.KeyId' --output text)"
  # create or update alias
  if awslocal kms list-aliases --query "Aliases[?AliasName=='${KMS_ALIAS}']" --output text | grep -q "${KMS_ALIAS}" ; then
    TARGET="$(awslocal kms list-aliases --query "Aliases[?AliasName=='${KMS_ALIAS}'].TargetKeyId" --output text || true)"
    if [[ -n "${TARGET}" && "${TARGET}" != "${KEY_ID}" ]]; then
      awslocal kms update-alias --alias-name "${KMS_ALIAS}" --target-key-id "${KEY_ID}" >/dev/null
    fi
  else
    awslocal kms create-alias --alias-name "${KMS_ALIAS}" --target-key-id "${KEY_ID}" >/dev/null
  fi

  # Give the role basic permissions to use the KMS key (dev-friendly)
  cat >/tmp/kms-policy.json <<'JSON'
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "kms:Encrypt","kms:Decrypt","kms:ReEncrypt*",
      "kms:GenerateDataKey*","kms:DescribeKey"
    ],
    "Resource": "*"
  }]
}
JSON
  awslocal iam put-role-policy \
    --role-name "$ROLE_NAME" \
    --policy-name "${ROLE_POLICY_NAME}-kms" \
    --policy-document file:///tmp/kms-policy.json >/dev/null
fi

# ---------- cosmetic "folders" so structure is visible ----------
for ds in "${DATASETS[@]}"; do
  awslocal s3api put-object --bucket "${BUCKET}" --key "${PREFIX_BASE}/${ds}/" >/dev/null
done

echo "[init] ready:"
echo "  role:   ${ROLE_NAME}"
echo "  bucket: s3://${BUCKET}/${PREFIX_BASE}/"
echo "  kms:    ${KMS_ALIAS} (created=${CREATE_KMS})"
echo
echo "Try listing:"
echo "  awslocal s3 ls s3://${BUCKET}/${PREFIX_BASE}/ --recursive --human-readable --summarize"
echo
echo "Writer prefix to use in your app:"
echo "  ${PREFIX_BASE}/feedback/demo/{yyyy}/{MM}/{dd}/{HH}/{mm}/"
