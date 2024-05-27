VERSION=$(shell cat VERSION)

ifeq ($(shell test -f "/etc/alpine-release" && echo -n true),true)
	TAGS=-tags musl
else
	TAGS=
endif

default: clean test build-local

# Running outside a container requires extra config. Checkout readme for more info.
.PHONY: run
run:
	go run cmd/filepoint/main.go -config config/config-local.yaml

.PHONY: run-webhooks-sender
run-webhooks-sender:
	go run cmd/filepoint-webhooks-sender/main.go -config config/config-local.yaml

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

REPOSITORY = "029272547936.dkr.ecr.us-east-1.amazonaws.com"

BASE_TAG = ${REPOSITORY}/prod-filepoint-base-repo:latest
.PHONY: build-base
build-base:
	docker build --tag ${BASE_TAG} -f "build/base/docker/Dockerfile" . \
	&& docker push ${BASE_TAG}

# The configuration file to be used.
# Important: if you pretend to use it in a Docker container for development,
# you can set this as a volume or build this with ""config/config.yaml" instead.
CONFIG_FILE = "config/config-prod.yaml"

# TAG controls the image tagging.
# You can use ${VERSION} or "latest"
TAG = ${VERSION}

.PHONY: build-images
build-images:
	scripts/build-images.sh ${REPOSITORY} ${CONFIG_FILE} ${VERSION} ${OS_ARCH}

FILEPOINT_TAG = ${REPOSITORY}/prod-filepoint-repo:${TAG}
.PHONY: publish-filepoint
publish-filepoint:
	docker tag filepoint:${VERSION} ${FILEPOINT_TAG} \
	&& docker push ${FILEPOINT_TAG}

WEBHOOKS_TAG = ${REPOSITORY}/prod-filepoint-webhooks-repo:${TAG}
.PHONY: publish-webhooks-sender
publish-webhooks-sender:
	docker tag filepoint-webhooks-sender:${VERSION} ${WEBHOOKS_TAG} \
	&& docker push ${WEBHOOKS_TAG}
