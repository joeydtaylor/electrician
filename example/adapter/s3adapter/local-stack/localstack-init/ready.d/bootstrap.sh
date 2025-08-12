#!/usr/bin/env bash
set -euo pipefail

# awslocal is preconfigured to hit the LocalStack endpoint.
ROLE_NAME="exodus-dev-role"
ROLE_POLICY_NAME="exodus-dev-s3"
BUCKET="steeze-dev"

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

cat >/tmp/s3-policy.json <<'JSON'
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["s3:*"],
    "Resource": ["*"]
  }]
}
JSON

echo "[init] provisioning IAM role + S3 bucket"

# idempotent role creation
if ! awslocal iam get-role --role-name "$ROLE_NAME" >/dev/null 2>&1; then
  awslocal iam create-role \
    --role-name "$ROLE_NAME" \
    --assume-role-policy-document file:///tmp/trust.json >/dev/null
fi

# attach inline policy
awslocal iam put-role-policy \
  --role-name "$ROLE_NAME" \
  --policy-name "$ROLE_POLICY_NAME" \
  --policy-document file:///tmp/s3-policy.json >/dev/null

# create bucket if missing
awslocal s3 ls "s3://$BUCKET" >/dev/null 2>&1 || awslocal s3 mb "s3://$BUCKET" >/dev/null

echo "[init] ready: role '$ROLE_NAME', bucket '$BUCKET'"
