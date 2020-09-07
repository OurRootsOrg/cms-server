# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=main.lambda
PUBLISHER_BINARY=publisher.lambda
RECORDSWRITER_BINARY=recordswriter.lambda
IMAGESWRITER_BINARY=imageswriter.lambda
PG_PORT=15432
RABBIT_PORT=35672
ES_PORT=19200

all: clean test build package
build:
	cd server && go generate && $(GOBUILD) && GOOS=linux $(GOBUILD) -o $(BINARY_NAME)
	cd publisher && $(GOBUILD) && GOOS=linux $(GOBUILD) -o $(PUBLISHER_BINARY)
	cd recordswriter && $(GOBUILD) && GOOS=linux $(GOBUILD) -o $(RECORDSWRITER_BINARY)
	cd imageswriter && $(GOBUILD) && GOOS=linux $(GOBUILD) -o $(IMAGESWRITER_BINARY)
package:
	zip -r deploy/awslambda/$(BINARY_NAME).zip server/$(BINARY_NAME) db/migrations/*
	zip -r deploy/awslambda/${PUBLISHER_BINARY}.zip publisher/$(PUBLISHER_BINARY)
	zip -r deploy/awslambda/${RECORDSWRITER_BINARY}.zip recordswriter/$(RECORDSWRITER_BINARY)
	zip -r deploy/awslambda/${IMAGESWRITER_BINARY}.zip imageswriter/$(IMAGESWRITER_BINARY)
test: test-setup test-exec test-teardown
test-setup:
	docker-compose -f docker-compose-dependencies.yaml up --detach --build
	cd db && ./wait-for-db.sh $(PG_PORT) && ./db_setup.sh $(PG_PORT) && ./db_load_test.sh $(PG_PORT)
	cd elasticsearch && ./wait-for-es.sh $(ES_PORT) && ./es_setup.sh $(ES_PORT)
	cd rabbitmq && ./wait-for-rabbitmq.sh ${RABBIT_PORT}
test-exec:
	DATABASE_URL="postgres://ourroots:password@localhost:$(PG_PORT)/cms?sslmode=disable" \
		DYNAMODB_TEST_TABLE_NAME="test-cms" \
		MIGRATION_DATABASE_URL="postgres://ourroots_schema:password@localhost:$(PG_PORT)/cms?sslmode=disable" \
    RABBIT_SERVER_URL="amqp://guest:guest@localhost:$(RABBIT_PORT)/" \
	$(GOTEST) -v -race -p=1 ./...
	# Re-run tests against DynamoDB
	# DYNAMODB_TABLE_NAME=test-cms \
  #   RABBIT_SERVER_URL="amqp://guest:guest@localhost:$(RABBIT_PORT)/" \
	# $(GOTEST) -v -race -p=1 ./...
test-teardown:
	docker-compose -f docker-compose-dependencies.yaml down --volumes
clean:
	rm -f server/$(BINARY_NAME)
	rm -f publisher/$(PUBLISHER_BINARY)
	rm -f recordswriter/$(RECORDSWRITER_BINARY)
	rm -f imageswriter/$(IMAGESWRITER_BINARY)
	rm -f server/server
	rm -f publisher/publisher
	rm -f recordswriter/recordswriter
	rm -f imageswriter/imageswriter
	rm -f deploy/awslambda/$(BINARY_NAME).zip
	rm -f deploy/awslambda/${PUBLISHER_BINARY}.zip
	rm -f deploy/awslambda/${RECORDSWRITER_BINARY}.zip
	rm -f deploy/awslambda/${IMAGESWRITER_BINARY}.zip
