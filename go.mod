module github.com/ourrootsorg/cms-server

go 1.17

replace github.com/awslabs/aws-lambda-go-api-proxy v0.8.0 => github.com/jancona/aws-lambda-go-api-proxy v0.6.1-0.20200804024701-b4721077da6b

require (
	github.com/DATA-DOG/go-sqlmock v1.4.1
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/aws/aws-lambda-go v1.28.0
	github.com/aws/aws-sdk-go v1.42.35
	github.com/awslabs/aws-lambda-go-api-proxy v0.12.0
	github.com/cenkalti/backoff/v4 v4.1.2
	github.com/codingconcepts/env v0.0.0-20200821220118-a8fbf8d84482
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/disintegration/imaging v1.6.2
	github.com/elastic/go-elasticsearch/v7 v7.16.0
	github.com/go-playground/validator/v10 v10.10.0
	github.com/golang-migrate/migrate/v4 v4.15.1
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.2.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hashicorp/logutils v1.0.0
	github.com/jriquelme/awsgosigv4 v0.0.0-20210621040219-41d17355154f
	github.com/lib/pq v1.10.4
	github.com/ourrootsorg/go-oidc v2.2.1+incompatible
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/http-swagger v1.1.2
	github.com/swaggo/swag v1.7.8
	gocloud.dev v0.24.0
	gocloud.dev/pubsub/rabbitpubsub v0.24.0
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	golang.org/x/text v0.3.7
)

require (
	cloud.google.com/go/iam v0.1.0 // indirect
	contrib.go.opencensus.io/integrations/ocsql v0.1.7 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/aws/aws-sdk-go-v2 v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.10.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.9.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.14.0 // indirect
	github.com/aws/smithy-go v1.10.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/wire v0.5.0 // indirect
	github.com/googleapis/gax-go/v2 v2.1.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/swaggo/files v0.0.0-20210815190702-a29dd2bc99b2 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce // indirect
	golang.org/x/image v0.0.0-20211028202545-6944b10bf410 // indirect
	golang.org/x/net v0.0.0-20220114011407-0dd24b26b47d // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
	golang.org/x/tools v0.1.8 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/api v0.65.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220114231437-d2e6a121cae0 // indirect
	google.golang.org/grpc v1.43.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
