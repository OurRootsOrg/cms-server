#!/bin/bash
set -e
ENVIRONMENT_NAME="${ENVIRONMENT_NAME:?}"
DOMAIN_NAME="${DOMAIN_NAME:?}"
# set -x
s3_bucket_name="${ENVIRONMENT_NAME}-deploy"
if aws s3 ls "s3://${s3_bucket_name}" 2>&1 | grep -q 'NoSuchBucket'
then
  echo Creating deploy bucket...
  aws s3 mb "s3://${s3_bucket_name}"
fi
echo Processing CloudFormation...
aws cloudformation package --template-file cms-infra.cf.yaml --output-template-file output-cms-infra.cf.yaml --s3-bucket $s3_bucket_name
aws cloudformation deploy --template-file output-cms-infra.cf.yaml --stack-name "${ENVIRONMENT_NAME}-infra" --parameter-overrides "EnvironmentName=${ENVIRONMENT_NAME}" "DomainName=${DOMAIN_NAME}" --capabilities CAPABILITY_IAM
# Enable Data API
aws rds modify-db-cluster --db-cluster-identifier "${ENVIRONMENT_NAME}-cms" --enable-http-endpoint --apply-immediately
# Get the secret ARNs from CF
# Get the app username and password from Secrets Manager
# Create the app user
echo Done