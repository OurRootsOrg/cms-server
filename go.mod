module github.com/ourrootsorg/cms-server

go 1.14

require (
	github.com/DATA-DOG/go-sqlmock v1.4.1
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/aws/aws-lambda-go v1.16.0
	github.com/aws/aws-sdk-go v1.19.45
	github.com/awslabs/aws-lambda-go-api-proxy v0.6.0
	github.com/cenkalti/backoff/v4 v4.0.2
	github.com/codingconcepts/env v0.0.0-20190614135724-bb4545dff6a4
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/elastic/go-elasticsearch/v7 v7.7.0
	github.com/go-openapi/spec v0.19.7 // indirect
	github.com/go-openapi/swag v0.19.9 // indirect
	github.com/go-playground/validator/v10 v10.2.0
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/golang/protobuf v1.4.0 // indirect
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/schema v1.1.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hashicorp/logutils v1.0.0
	github.com/lib/pq v1.3.0
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/stretchr/testify v1.5.1
	github.com/swaggo/http-swagger v0.0.0-20200308142732-58ac5e232fba
	github.com/swaggo/swag v1.6.5
	gocloud.dev v0.19.0
	gocloud.dev/pubsub/rabbitpubsub v0.19.0
	golang.org/x/net v0.0.0-20200501053045-e0ff5e5a1de5 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200420163511-1957bb5e6d1f // indirect
	golang.org/x/text v0.3.2
	golang.org/x/tools v0.0.0-20200501205727-542909fd9944 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

// replace github.com/awslabs/aws-lambda-go-api-proxy => /Users/jim/jimprojects/aws-lambda-go-api-proxy
