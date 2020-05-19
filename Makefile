# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=main.lambda
PG_PORT=15432
RABBIT_PORT=25672

all: clean test build package
build: 
	cd server && go generate && $(GOBUILD) && GOOS=linux $(GOBUILD) -o $(BINARY_NAME)
package:
	zip -r deploy/awslambda/$(BINARY_NAME).zip server/$(BINARY_NAME) 
test: test-setup test-exec test-teardown
test-setup:
	docker-compose -f docker-compose-dependencies.yaml up --detach --build
	cd db && ./wait-for-db.sh $(PG_PORT) && ./db_setup.sh $(PG_PORT)
	rabbitmq/wait-for-rabbitmq.sh ${RABBIT_PORT}
test-exec:
	DATABASE_URL="postgres://ourroots:password@localhost:$(PG_PORT)/cms?sslmode=disable" \
    RABBIT_SERVER_URL="amqp://guest:guest@localhost:$(RABBIT_PORT)/" \
	$(GOTEST) -v ./...
test-teardown:
	docker-compose -f docker-compose-dependencies.yaml down --volumes
clean:
	rm -f server/$(BINARY_NAME)
	rm -f server/server
	rm -f deploy/awslambda/$(BINARY_NAME).zip
