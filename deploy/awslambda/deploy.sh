#!/bin/bash
set -e
ENVIRONMENT_NAME="${ENVIRONMENT_NAME:?}"
DOMAIN_NAME="${DOMAIN_NAME:?}"
CERTIFICATE_ARN="${CERTIFICATE_ARN:?}"
USE_POSTGRES="${USE_POSTGRES:-true}"
script="$0"
scriptdir="$(dirname $script)"
# set -x
s3_bucket_name="${ENVIRONMENT_NAME}-deploy"
echo Processing CloudFormation...
if [ "${USE_POSTGRES}" == "true" ]
then
  template_file="cms-aurora.cf.yaml"
else
  template_file="cms-dynamodb.cf.yaml"
fi
aws cloudformation package --template-file "${template_file}" --output-template-file output-cms.cf.yaml --s3-bucket $s3_bucket_name
aws cloudformation deploy --template-file output-cms.cf.yaml --stack-name "${ENVIRONMENT_NAME}-deploy" --parameter-overrides "EnvironmentName=${ENVIRONMENT_NAME}" "DomainName=${DOMAIN_NAME}" "CertificateArn=${CERTIFICATE_ARN}" "CMSSiteBucketURL=${ENVIRONMENT_NAME}-CMSSiteBucketURL" "CMSPostgresAddress=${ENVIRONMENT_NAME}-CMSPostgresAddress" "CMSPostgresPort=${ENVIRONMENT_NAME}-CMSPostgresPort" "AuroraMasterSecretARN=${ENVIRONMENT_NAME}-AuroraMasterSecretARN" "AuroraAppSecretARN=${ENVIRONMENT_NAME}-AuroraAppSecretARN" "CMSBlobStoreBucketName=${ENVIRONMENT_NAME}-CMSBlobStoreBucketName" "CMSRecordsWriterQueueURL=${ENVIRONMENT_NAME}-CMSRecordsWriterQueueURL" "CMSRecordsWriterQueueARN=${ENVIRONMENT_NAME}-CMSRecordsWriterQueueARN" "CMSImagesWriterQueueURL=${ENVIRONMENT_NAME}-CMSImagesWriterQueueURL" "CMSImagesWriterQueueARN=${ENVIRONMENT_NAME}-CMSImagesWriterQueueARN" "CMSPublisherQueueURL=${ENVIRONMENT_NAME}-CMSPublisherQueueURL" "CMSPublisherQueueARN=${ENVIRONMENT_NAME}-CMSPublisherQueueARN" "ElasticsearchDomainARN=${ENVIRONMENT_NAME}-ElasticsearchDomainARN" "LambdaFunctionSecurityGroup=${ENVIRONMENT_NAME}-LambdaFunctionSecurityGroup" "PrivateSubnet1=${ENVIRONMENT_NAME}-PrivateSubnet1" "PrivateSubnet2=${ENVIRONMENT_NAME}-PrivateSubnet2" "ElasticsearchDomainEndpoint=${ENVIRONMENT_NAME}-ElasticsearchDomainEndpoint" --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM

# echo Uploading static site content...
aws s3 sync "${scriptdir}/../../client/dist/" "s3://${ENVIRONMENT_NAME}-site/" --delete
aws s3 sync "${scriptdir}/../../search-client/dist/" "s3://${ENVIRONMENT_NAME}-site/search/" --delete
echo Done
