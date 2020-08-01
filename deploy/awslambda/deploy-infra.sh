#!/bin/bash
set -e
ENVIRONMENT_NAME="${ENVIRONMENT_NAME:?}"
USE_POSTGRES="${USE_POSTGRES:-true}"
ES_ADMIN_CIDR="${ES_ADMIN_CIDR:-0.0.0.0/0}"
# DOMAIN_NAME="${DOMAIN_NAME:?}"
# set -x
s3_bucket_name="${ENVIRONMENT_NAME}-deploy"
if aws s3 ls "s3://${s3_bucket_name}" 2>&1 | grep -q 'NoSuchBucket'
then
  echo Creating deploy bucket...
  aws s3 mb "s3://${s3_bucket_name}"
fi
if [ "${USE_POSTGRES}" == "true" ]
then
  template_file="cms-infra-aurora.cf.yaml"
else
  template_file="cms-infra-dynamodb.cf.yaml"
fi
echo Processing CloudFormation...
aws cloudformation package --template-file "${template_file}" --output-template-file output-cms-infra.cf.yaml --s3-bucket $s3_bucket_name
aws cloudformation deploy --template-file output-cms-infra.cf.yaml --stack-name "${ENVIRONMENT_NAME}-infra" --parameter-overrides "EnvironmentName=${ENVIRONMENT_NAME}" "ESAdminCIDR=${ES_ADMIN_CIDR}" --capabilities CAPABILITY_IAM

if [ "${USE_POSTGRES}" == "true" ]
then
  # Enable Data API
  aws rds modify-db-cluster --db-cluster-identifier "${ENVIRONMENT_NAME}-cms" --enable-http-endpoint --apply-immediately
fi
echo Done