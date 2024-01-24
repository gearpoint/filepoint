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

.PHONY: build-local
build-local: prepare
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

DOCKER_REPO = "gearpoint"
SUFFIX_TAG = "latest"

.PHONY: build-image
build-image:
	scripts/build-image.sh ".config/config-docker.yaml" ${VERSION} ${DOCKER_REPO} ${SUFFIX_TAG} ${OS_ARCH}

.PHONY: build-image-prod
build-image-prod:
	scripts/build-image.sh ".config/config-prod.yaml" ${VERSION} ${DOCKER_REPO} ${SUFFIX_TAG} ${OS_ARCH}

IMAGE_NAME = "filepoint"
.PHONY: filepoint-publish
filepoint-publish:
	docker push ${DOCKER_REPO}/${FILEPOINT}:${SUFFIX_TAG}

IMAGE_NAME = "filepoint-webhooks-sender"
.PHONY: webhooks-sender-publish
webhooks-sender-publish:
	docker push ${DOCKER_REPO}/${IMAGE_NAME}:${SUFFIX_TAG}
