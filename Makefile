# Includes .env file
# Currently, user REGISTRY variable to get the ECR URL
include .env

VERSION=$(shell cat VERSION)

ifeq ($(shell test -f "/etc/alpine-release" && echo -n true),true)
	TAGS=-tags musl
else
	TAGS=
endif

default: clean test build-local

.PHONY: run
run:
	go run cmd/filepoint/main.go

.PHONY: run-webhooks-sender
run-webhooks-sender:
	go run cmd/filepoint-webhooks-sender/main.go

.PHONY: deps
deps:
	go mod download

.PHONY: clean
clean:
	rm -rf target

.PHONY: prepare
prepare:
	chmod +x scripts/* && scripts/prepare.sh

.PHONY: build-binary
build-binary: prepare
	scripts/build-binary.sh ${VERSION} ${TAGS}

.PHONY: test
test: prepare
	go test -outputdir=target/tests -coverprofile=coverage.out -v ./... \
	&& go tool cover -func target/tests/coverage.out

.PHONY: integration-test
integration-test: prepare
	go test -tags integration -outputdir=target/tests -coverprofile=coverage.out -v  ./... \
	&& go tool cover -func target/tests/coverage.out

.PHONY: generate
generate:
	go generate ./...

.PHONY: swagger
swagger:
	swag init -g cmd/filepoint/main.go --output api

.PHONY: godoc
godoc:
	godoc -http=:6060

BASE_TAG = ${REGISTRY}/prod-filepoint-base-repo:latest
.PHONY: build-base
build-base:
	docker build --tag ${BASE_TAG} -f "build/base/docker/Dockerfile" .

.PHONY: base-publish
publish-base:
	docker push ${BASE_TAG}

CONFIG_FILE = "config/config-prod.yaml"
SUFFIX = ${VERSION}-latest

.PHONY: build-images
build-images:
	scripts/build-images.sh ${REGISTRY} ${CONFIG_FILE} ${SUFFIX} ${OS_ARCH}

FILEPOINT_TAG = ${REGISTRY}/prod-filepoint-repo:${SUFFIX}
.PHONY: publish-filepoint
publish-filepoint:
	docker tag filepoint:${SUFFIX} ${FILEPOINT_TAG} \
	&& docker push ${FILEPOINT_TAG}

WEBHOOKS_TAG = ${REGISTRY}/prod-filepoint-webhooks-repo:${SUFFIX}
.PHONY: publish-webhooks-sender
publish-webhooks-sender:
	docker tag filepoint-webhooks-sender:${SUFFIX} ${WEBHOOKS_TAG} \
	&& docker push ${WEBHOOKS_TAG}
