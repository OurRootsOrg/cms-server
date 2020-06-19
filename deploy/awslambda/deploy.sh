#!/bin/bash
set -e
ENVIRONMENT_NAME="${ENVIRONMENT_NAME:?}"
DOMAIN_NAME="${DOMAIN_NAME:?}"
CERTIFICATE_ARN="${CERTIFICATE_ARN:?}"
script="$0"
scriptdir="$(dirname $script)"
# set -x
s3_bucket_name="${ENVIRONMENT_NAME}-deploy"
echo Processing CloudFormation...
aws cloudformation package --template-file cms.cf.yaml --output-template-file output-cms.cf.yaml --s3-bucket $s3_bucket_name
aws cloudformation deploy --template-file output-cms.cf.yaml --stack-name "${ENVIRONMENT_NAME}-deploy" --parameter-overrides "EnvironmentName=${ENVIRONMENT_NAME}" "DomainName=${DOMAIN_NAME}" "CertificateArn=${CERTIFICATE_ARN}" "CMSSiteBucketURL=${ENVIRONMENT_NAME}-CMSSiteBucketURL" "CMSPostgresAddress=${ENVIRONMENT_NAME}-CMSPostgresAddress" "CMSPostgresPort=${ENVIRONMENT_NAME}-CMSPostgresPort" "AuroraMasterSecretARN=${ENVIRONMENT_NAME}-AuroraMasterSecretARN" "AuroraAppSecretARN=${ENVIRONMENT_NAME}-AuroraAppSecretARN" "CMSBlobStoreBucketName=${ENVIRONMENT_NAME}-CMSBlobStoreBucketName" "CMSRecordsWriterQueueURL=${ENVIRONMENT_NAME}-CMSRecordsWriterQueueURL" "CMSRecordsWriterQueueARN=${ENVIRONMENT_NAME}-CMSRecordsWriterQueueARN" "CMSPublisherQueueURL=${ENVIRONMENT_NAME}-CMSPublisherQueueURL" "CMSPublisherQueueARN=${ENVIRONMENT_NAME}-CMSPublisherQueueARN" "ElasticsearchDomainARN=${ENVIRONMENT_NAME}-ElasticsearchDomainARN" "LambdaFunctionSecurityGroup=${ENVIRONMENT_NAME}-LambdaFunctionSecurityGroup" "PrivateSubnet1=${ENVIRONMENT_NAME}-PrivateSubnet1" "PrivateSubnet2=${ENVIRONMENT_NAME}-PrivateSubnet2" "ElasticsearchDomainEndpoint=${ENVIRONMENT_NAME}-ElasticsearchDomainEndpoint" --capabilities CAPABILITY_IAM

# echo Uploading static site content...
pushd "${scriptdir}/../../uglyui/"
# This should probably be done in the Makefile
npm run build
popd
aws s3 sync "${scriptdir}/../../uglyui/dist/" "s3://${ENVIRONMENT_NAME}-site/" --delete
echo Done