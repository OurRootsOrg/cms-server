# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=main.lambda
    
all: clean test build package
build: 
	cd server && go generate && $(GOBUILD) && GOOS=linux $(GOBUILD) -o $(BINARY_NAME)
package:
	zip -r deploy/awslambda/$(BINARY_NAME).zip server/$(BINARY_NAME) 
test: 
	$(GOTEST) -v ./...
clean: 
	rm -f server/$(BINARY_NAME)
	rm -f deploy/awslambda/$(BINARY_NAME).zip
