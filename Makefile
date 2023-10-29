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

.PHONY: build-image
build-image:
	scripts/build-image.sh ${VERSION} ${REGISTRY} ${SUFFIX_TAG} ${OS_ARCH}

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
FILEPOINT = "filepoint"
WEBHOOKS_SENDER = "filepoint-webhooks-sender"
VERSION=1.0.0-SNAPSHOT

.PHONY: filepoint-tag-latest
filepoint-tag-latest:
	docker tag $(FILEPOINT):$(VERSION) $(DOCKER_REPO)/$(FILEPOINT):latest

.PHONY: filepoint-publish-latest
filepoint-publish-latest: filepoint-tag-latest
	docker push $(DOCKER_REPO)/$(FILEPOINT):latest

.PHONY: webhooks-sender-tag-latest
webhooks-sender-tag-latest:
	docker tag $(WEBHOOKS_SENDER):$(VERSION) $(DOCKER_REPO)/$(WEBHOOKS_SENDER):latest

.PHONY: webhooks-sender-publish-latest
webhooks-sender-publish-latest: webhooks-sender-tag-latest
	docker push $(DOCKER_REPO)/$(WEBHOOKS_SENDER):latest
