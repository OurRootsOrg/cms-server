# Prerequisites
To deploy to AWS, you must have:

* A successful build of the project, as described in the top-level [README](../../README.md).
* An AWS account.
* Account credentials with the proper permissions. In general, you [should not use the AWS root credentials](https://docs.aws.amazon.com/general/latest/gr/root-vs-iam.html) for tasks like this. An IAM user with either the `Administrator` policy or the `PowerUserAccess` policy plus the ability to manage IAM roles should work.
* A working [AWS CLI installation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html).
It should be [configured](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) so that you can run AWS CLI commands in that account, e.g. if you run `aws s3 ls` you get a list of the buckets in the AWS account you want to use. You should also be sure that the CLI is configured to use the region where you want the app to run.

Right now the deploy scripts don't manage DNS and don't create a TLS certificate. So you will need the ability to configure a CNAME pointing to the OurRoots application's AWS domain. (You can manage DNS at AWS or at some other provider.)

Once you decide on the domain where the application will be deployed (e.g. `app.ourroots.org`), you will need to use [AWS Certificate Manager](https://console.aws.amazon.com/acm/home) to create and validate a certificate. Once you have done so, make a record of the ARN (Amazon Resource Name) of your certificate. It can be found on the details page in the console and will look something like this: `arn:aws:acm:us-east-1:123456789012:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`.
- Note, you must request the certificate in the correct region!

Then run the following commands from the `deploy/awslambda` directory:

* `<env-name>` is the name of this deployment. It is used for internal names of AWS resources and won't be visible to users. Example: `ourroots-preprod`
* `<aws-region>` is the AWS region where the app is deployed. Note that it should match the region for the AWS CLI. Example: `us-east-1`
* `<domain-name>` is the domain name where the application will run. (See above.) Example: `app.ourroots.org`
* `<cert-arn>` is the ARN for the certificate you configured above. Example: `arn:aws:acm:us-east-1:123456789012:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`.

To deploy to AWS using Serverless RDS Postgres as a database, execute the following commands:
```
ENVIRONMENT_NAME=<env-name> ./deploy-infra.sh
AWS_REGION=<aws-region> go run dbconfig/dbconfig.go <env-name>
ENVIRONMENT_NAME=<env-name> DOMAIN_NAME="<domain-name>" CERTIFICATE_ARN="<cert-arn>" ./deploy.sh
```

* If `deploy-infra` fails, run `aws cloudformation delete-stack --stack-name ENVIRONMENT_NAME-infra` and try again
* Modify OpenSearch in the console to be 2-AZ, t3.small.search, 2 nodes, gp2 storage
* if you update the client or server code, re-build and rerun the deploy.sh script

Load the database for db_load_full.sh against the RDS database (just need to do this once)

* Launch an EC2 micro instance (make sure you are in the correct region)
  * make sure you are in the correct region
  * make sure you launch it in the same VPC as the RDS Postgres database
  * make sure you select a public subnet
* Add the instance's security group to the security group for RDS: AuroraClusterSecurityGroup
  * Allow PostgresQL traffic 
* copy the load scripts: `scp -i sbcgs.pem db/db_load*.sh ec2-user@<IP_ADDR>:/home/ec2-user`
* copy the migrations directory: `scp -i sbcgs.pem -r cms-server/db/migrations ec2-user@<IP_ADDR>:/home/ec2-user`
* grab the password and the host for the ourroots_schema user from AWS Secrets manager postgres/secrets/master
* ssh to your instance
* install postgres: `sudo dnf update && sudo dnf install postgresql15`
* install migrate `curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz`
* run migrate: `./migrate -database "postgres://ourroots_schema:<password>@<host>:5432/cms?sslmode=disable" -path migrations up` 
* load the database `./db_load_full.sh ourroots_schema <password> <host>`
* stop the EC2 micro instance


<!--- The DynamoDB code hasn't been updated to work with the latest features

To deploy to AWS using DynamoDB as a database, execute the following commands:
```
USE_POSTGRES=false ENVIRONMENT_NAME=<env-name> ./deploy-infra.sh
USE_POSTGRES=false ENVIRONMENT_NAME=<env-name> DOMAIN_NAME="<domain-name>" CERTIFICATE_ARN="<cert-arn>" ./deploy.sh
cd ../../db/dynamo/ddbloader
AWS_REGION=<aws-region> ENVIRONMENT_NAME=<env-name> ./ddb_load_full.sh
```
-->

After those commands complete without errors, you will need to configure a DNS CNAME. 
Go to [https://console.aws.amazon.com/apigateway/main/publish/domain-names] and select the entry for the domain name you selected above. 
Make a record of the "API Gateway domain name" on the details page. 
It should look like `x-xxxxxxxxxx.execute-api.us-east-1.amazonaws.com`. 
At your DNS provider, create a CNAME record pointing your domain name (i.e. `app.ourroots.org`) to the API Gateway domain name. 
Once that is done and the DNS has propagated, you should be able to see the home page of the app at your domain. 
(Example: `https://app.ourroots.org`).

The final step is to configure authentication at Auth0.

- Create an App
  - Application Type: Single page application
  - Allowed Callback URLs: http://localhost:8080, http://localhost:3000/swagger/oauth2-redirect.html, http://localhost:8080/callback.html, https://app.sbgen-ourroots.com/, https://app.sbgen-ourroots.com/callback.html, https://app.sbgen-ourroots.com/api/swagger/oauth2-redirect.html
  - Allowed Logout URLs: http://localhost:8080, http://localhost:3000, https://app.sbgen-ourroots.com
  - Allowed Web Origins: http://localhost:8080, http://localhost:3000, https://app.sbgen-ourroots.com
- Create an API
  - Audience Identifier: OIDC_AUDIENCE from cms-aurora.cf.yaml
  - Signing Algorithm: RS256
  - Permission: CMS, Grands read and write access to the CMS
- Set the default audience in Settings to OIDC_AUDIENCE
