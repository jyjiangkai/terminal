DOCKER_REPO="hub.easystack.io/production/terminal"
DOCKER_TAG="v1"
NAME = terminal
PROJECT = github.com/easystack/terminal
PROJECT_DIR = $(GOPATH)src/github.com/easystack

all: fmt prepare build

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

prepare:
	mkdir -p bin
#	cp config/crd/* bin/

# Build binarys
build:
	go build -o bin/terminal cmd/main.go

image: build-image push-image

build-image:
	go build -o bin/terminal cmd/main.go
	docker build . -t $(DOCKER_REPO):$(DOCKER_TAG)

push-image:
	docker push $(DOCKER_REPO):$(DOCKER_TAG)


.PHONY: copy-k8s-device-plugin
copy-k8s-device-plugin:
	@mkdir -p $(PROJECT_DIR)
	@cp -rf $(shell command pwd;) $(PROJECT_DIR)
	@mkdir -p $(PROJECT_DIR)/$(NAME)/bin
	@cd $(PROJECT_DIR)/$(NAME)
	@echo $(PROJECT_DIR)/$(NAME)

.PHONY: test-style
test-style:
	@echo "TODO"

.PHONY: test-unit
test-unit:
	@echo "TODO"

.PHONY: coverage
coverage:
	@echo "TODO"
