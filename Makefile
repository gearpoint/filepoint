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
	scripts/build-binary.sh ${VERSION}

.PHONY: build-image
build-image:
	scripts/build-image.sh ${VERSION} ${REGISTRY} ${SUFFIX_TAG} ${OS_ARCH}

.PHONY: test
test: prepare
	go test -outputdir=target/tests -coverprofile=coverage.out -v ./... \
	&& go tool cover -func target/tests/coverage.out

.PHONY: integration-test
integration-test: prepare
	export GIN_MODE=debug \
	&& go test -tags integration -outputdir=target/tests -coverprofile=coverage.out -v  ./... \
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