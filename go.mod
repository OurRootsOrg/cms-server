module github.com/ourrootsorg/cms-server

go 1.19

replace github.com/awslabs/aws-lambda-go-api-proxy v0.8.0 => github.com/jancona/aws-lambda-go-api-proxy v0.6.1-0.20200804024701-b4721077da6b

require (
	github.com/DATA-DOG/go-sqlmock v1.4.1
	github.com/aws/aws-lambda-go v1.18.0
	github.com/aws/aws-sdk-go v1.34.0
	github.com/awslabs/aws-lambda-go-api-proxy v0.8.0
	github.com/cenkalti/backoff/v4 v4.0.2
	github.com/codingconcepts/env v0.0.0-20190614135724-bb4545dff6a4
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/disintegration/imaging v1.6.2
	github.com/elastic/go-elasticsearch/v7 v7.7.0
	github.com/go-playground/validator/v10 v10.2.0
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/schema v1.1.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hashicorp/logutils v1.0.0
	github.com/jriquelme/awsgosigv4 v0.0.0-20200515043227-0e5300b5f3e2
	github.com/lib/pq v1.3.0
	github.com/ourrootsorg/go-oidc v2.2.1-cognito2+incompatible
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/stretchr/testify v1.8.0
	github.com/swaggo/http-swagger v0.0.0-20200308142732-58ac5e232fba
	github.com/swaggo/swag v1.6.7
	gocloud.dev v0.19.0
	gocloud.dev/pubsub/rabbitpubsub v0.19.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/text v0.3.7
)

require (
	contrib.go.opencensus.io/integrations/ocsql v0.1.4 // indirect
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/aws/aws-sdk-go-v2 v0.23.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.7 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.4.0 // indirect
	github.com/google/wire v0.3.0 // indirect
	github.com/googleapis/gax-go v2.0.2+incompatible // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.3.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/stretchr/objx v0.4.0 // indirect
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14 // indirect
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073 // indirect
	golang.org/x/image v0.0.0-20191009234506-e7c1f5e7dbb8 // indirect
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/sys v0.1.0 // indirect
	golang.org/x/tools v0.0.0-20200213224642-88e652f7a869 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20200212174721-66ed5ce911ce // indirect
	google.golang.org/grpc v1.27.1 // indirect
	google.golang.org/protobuf v1.21.0 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
