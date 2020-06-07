To deploy to AWS, you must have an AWS account and credentials set up so that you can run AWS CLI commands in that account. Then run the following commands from the `deploy/awslambda` directory.
```
ENVIRONMENT_NAME=ourroots-cms-dev ./deploy-infra.sh
AWS_REGION=us-east-1 go run dbconfig/dbconfig.go ourroots-cms-dev
ENVIRONMENT_NAME=ourroots-cms-dev DOMAIN_NAME="ourroots.anconafamily.com" CERTIFICATE_ARN="arn:aws:acm:us-east-1:481386943213:certificate/0aa37842-490f-47a3-a54e-397e25108a40" ./deploy.sh
```