DOCKER_IMAGE         ?= adobe/cloudinventory
DOCKER_IMAGE_TAG     ?= master
DOCKER_IMAGE_TAG_ARM ?= armhf


all: mod-tidy test vet lint install
mod: mod-tidy verify get

mod-tidy:
	@echo ">> Running go mod tidy"
	@GO111MODULE=on go mod tidy

vendor:
	@echo ">> Running go mod vendor"
	@GO111MODULE=on go mod vendor

verify:
	@echo ">> Running go mod verify"
	@GO111MODULE=on go mod verify

get:
	@echo ">> Getting Dependencies"
	@GO111MODULE go get

build:
	@echo ">> Running Build"
	@GO111MODULE=on go build

build-main:
	@echo ">> Building Binary for current ARCH"
	@go build

install:
	@echo ">> Building and Installing"
	@GO111MODULE=on go install
	@echo ">> Done Install"

test-short:
	@echo ">> Running Quick Tests"
	@GO111MODULE=on go test -short ./...

test:
	@echo ">> Running Tests"
	@GO111MODULE=on go test -cover -v ./...

vet:
	@echo ">> Running Vet"
	@GO111MODULE=on go vet ./...

lint:
	@echo ">> Running Lint"
	@GO111MODULE=on go list ./... | grep -v vender/ | golint

docker:
	@echo ">> Building Docker Image"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG) .

docker-push:
	@docker login -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD)
	@docker push $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG)

docker-arm:
	@echo ">> Building Docker Image-ARM"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG_ARM) -f Dockerfile-ARM .

docker-push-arm:
	@docker login -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD)
	@docker push $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG_ARM)
