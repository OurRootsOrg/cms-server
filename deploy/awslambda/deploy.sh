#!/bin/bash
set -e
STACK_NAME="${STACK_NAME:?}"
DOMAIN_NAME="${DOMAIN_NAME:?}"
CERTIFICATE_ARN="${CERTIFICATE_ARN:?}"
script="$0"
scriptdir="$(dirname $script)"
# set -x
s3_bucket_name="${STACK_NAME}-deploy"
if aws s3 ls "s3://${s3_bucket_name}" 2>&1 | grep -q 'NoSuchBucket'
then
  echo Creating deploy bucket...
  aws s3 mb "s3://${s3_bucket_name}"
fi
echo Processing CloudFormation...
aws cloudformation package --template-file ourroots.cf.yaml --output-template-file output-ourroots.cf.yaml --s3-bucket $s3_bucket_name
aws cloudformation deploy --template-file output-ourroots.cf.yaml --stack-name $STACK_NAME --parameter-overrides "DomainName=${DOMAIN_NAME}" "CertificateArn=${CERTIFICATE_ARN}" --capabilities CAPABILITY_IAM
# echo Uploading static site content...
# aws s3 sync "${scriptdir}/../../vectyui/build/web/" "s3://${STACK_NAME}-site/wasm/"
# aws s3 sync "${scriptdir}/../../flutterui/build/web/" "s3://${STACK_NAME}-site/flutter/"
echo Done