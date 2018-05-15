DOCKER_IMAGE         ?= tchaudhry/cloudinventory
DOCKER_IMAGE_TAG     ?= master
DOCKER_IMAGE_TAG_ARM ?= armhf


all: test vet lint install
build:
	@echo ">> Running Build"
	@go build ./...

build-main:
	@echo ">> Building Binary for current ARCH"
	@go build

install:
	@echo ">> Building and Installing"
	@go install

test:
	@echo ">> Running Tests"
	@go test -v ./...

vet:
	@echo ">> Running Vet"
	@go vet ./...

lint:
	@echo ">> Running Lint"
	@golint ./...

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
