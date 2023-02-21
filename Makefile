
# generated-from:9b190835b2d77961188fa78aaa24d9c1730f764e3db70f91a664febb2e938681 DO NOT REMOVE, DO UPDATE

PLATFORM=$(shell uname -s | tr '[:upper:]' '[:lower:]')
PWD := $(shell pwd)

ifndef VERSION
	VERSION := $(shell git describe --tags --abbrev=0)
endif

COMMIT_HASH :=$(shell git rev-parse --short HEAD)
DEV_VERSION := dev-${COMMIT_HASH}

USERID := $(shell id -u $$USER)
GROUPID:= $(shell id -g $$USER)

export GOPRIVATE=github.com/moov-io

all: install update build

.PHONY: install
install:
	go install github.com/markbates/pkger/cmd/pkger
	go mod vendor

update:
	pkger -include /configs/config.default.yml
	go mod vendor

build:
	go build -mod=vendor -ldflags "-X github.com/moov-io/ach-test-harness.Version=${VERSION}" -o bin/ach-test-harness github.com/moov-io/ach-test-harness/cmd/ach-test-harness

.PHONY: setup
setup:
	docker-compose up -d --force-recreate --remove-orphans

.PHONY: check
check:
ifeq ($(OS),Windows_NT)
	@echo "Skipping checks on Windows, currently unsupported."
else
	@wget -O lint-project.sh https://raw.githubusercontent.com/moov-io/infra/master/go/lint-project.sh
	@chmod +x ./lint-project.sh
	COVER_THRESHOLD=60.0 ./lint-project.sh
endif

.PHONY: teardown
teardown:
	-docker-compose down --remove-orphans

docker: update
	docker build --pull --build-arg VERSION=${VERSION} -t moov/ach-test-harness:${VERSION} -f Dockerfile .
	docker tag moov/ach-test-harness:${VERSION} moov/ach-test-harness:latest

docker-push:
	docker push moov/ach-test-harness:${VERSION}
	docker push moov/ach-test-harness:latest

.PHONY: dev-docker
dev-docker: update
	docker build --pull --build-arg VERSION=${DEV_VERSION} -t moov/ach-test-harness:${DEV_VERSION} -f Dockerfile .
	docker tag moov/ach-test-harness:${DEV_VERSION} moov/ach-test-harness:${DEV_VERSION}

.PHONY: dev-push
dev-push:
	docker push moov/ach-test-harness:${DEV_VERSION}
	docker push moov/ach-test-harness:${DEV_VERSION}

# Extra utilities not needed for building

run: update build
	./bin/ach-test-harness

docker-run:
	docker run -v ${PWD}/data:/data -v ${PWD}/configs:/configs --env APP_CONFIG="/configs/config.yml" -it --rm moov/ach-test-harness:${VERSION}

test:
	go test -cover github.com/moov-io/ach-test-harness/...

.PHONY: clean
clean:
ifeq ($(OS),Windows_NT)
	@echo "Skipping cleanup on Windows, currently unsupported."
else
	@rm -rf cover.out coverage.txt misspell* staticcheck*
	@rm -rf ./bin/
endif

# For open source projects

# From https://github.com/genuinetools/img
.PHONY: AUTHORS
AUTHORS:
	@$(file >$@,# This file lists all individuals having contributed content to the repository.)
	@$(file >>$@,# For how it is generated, see `make AUTHORS`.)
	@echo "$(shell git log --format='\n%aN <%aE>' | LC_ALL=C.UTF-8 sort -uf)" >> $@

dist: clean build
ifeq ($(OS),Windows_NT)
	CGO_ENABLED=1 GOOS=windows go build -o bin/ach-test-harness.exe cmd/ach-test-harness/*
else
	CGO_ENABLED=1 GOOS=$(PLATFORM) go build -o bin/ach-test-harness-$(PLATFORM)-amd64 cmd/ach-test-harness/*
endif
