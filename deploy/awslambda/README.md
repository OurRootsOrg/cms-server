# Prerequisites
To deploy to AWS, you must have:

* An AWS account.
* Account credentials with the proper permissions. In general, you [should not use the AWS root credentials](https://docs.aws.amazon.com/general/latest/gr/root-vs-iam.html) for tasks like this. An IAM user with either the `Administrator` policy or the `PowerUserAccess` policy plus the ability to manage IAM roles should work.
* A working [AWS CLI installation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html).
It should be [configured](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) so that you can run AWS CLI commands in that account, e.g. if you run `aws s3 ls` you get a list of the buckets in the AWS account you want to use. You should also be sure that the CLI is configured to use the region where you want the app to run.

Right now the deploy scripts don't manage DNS and don't create a TLS certificate. So you will need the ability to configure a CNAME pointing to the OurRoots application's AWS domain. (You can manage DNS at AWS or at some other provider.)

Once you decide on the domain where the application will be deployed (e.g. `app.ourroots.org`), you will need to use [AWS Certificate Manager](https://console.aws.amazon.com/acm/home) to create and validate a certificate. Once you have done so, make a record of the ARN (Amazon Resource Name) of your certificate. It can be found on the details page in the console and will look something like this: `arn:aws:acm:us-east-1:123456789012:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`.

Then run the following commands from the `deploy/awslambda` directory:

* `<env-name>` is the name of this deployment. It is used for internal names of AWS resources and won't be visible to users. Example: `ourroots-preprod`
* `<aws-region>` is the AWS region where the app is deployed. Note that it should match the region for the AWS CLI. Example: `us-east-1`
* `<domain-name>` is the domain name where the application will run. (See above.) Example: `app.ourroots.org`
* `<cert-arn>` is the ARN for the certificate you configured above. Example: `arn:aws:acm:us-east-1:123456789012:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`.

```
ENVIRONMENT_NAME=<env-name> ./deploy-infra.sh
AWS_REGION=<aws-region> go run dbconfig/dbconfig.go <env-name>
ENVIRONMENT_NAME=<env-name> DOMAIN_NAME="<domain-name>" CERTIFICATE_ARN="<cert-arn>" ./deploy.sh
```

After those commands complete without errors, you will need to configure a DNS CNAME. Go to [https://console.aws.amazon.com/apigateway/main/publish/domain-names] and select the entry for the domain name you selected above. Make a record of the "API Gateway domain name" on the details page. It should look like `x-xxxxxxxxxx.execute-api.us-east-1.amazonaws.com`. At your DNS provider, create a CNAME record pointing your domain name (i.e. `app.ourroots.org`) to the API Gateway domain name. Once that is done and the DNS has propagated, you should be able to see the home page of the app at your domain. (Example: `https://app.ourroots.org`).

The final step is to configure authentication at Auth0.

*Details to follow*
