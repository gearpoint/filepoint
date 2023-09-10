FROM golang:1.21 as modules

ARG OS_NAME
ARG ARCH
ARG BINARY_NAME

ENV GOOS=${OS_NAME:-linux} \
    GOARCH=${ARCH:-amd64}

WORKDIR /code

COPY Makefile go.mod go.sum ./

RUN make deps

# Start from the latest golang base image
FROM golang:1.21 as builder

ARG OS_NAME
ARG ARCH
ARG BINARY_NAME

ENV GOOS=${OS_NAME:-linux} \
    GOARCH=${ARCH:-amd64}

WORKDIR /code

COPY . .

RUN echo ${GOOS}/${GOARCH} > scripts/plataforms \
    && make clean test build-local

######## Start a new stage from scratch #######
FROM alpine:latest

RUN apk --no-cache add ca-certificates
RUN apk add libc6-compat

LABEL maintainer="Gearpoint" \
 contributor="Luan Baggio <luanbaggio0@gmail.com>"

ARG OS_NAME
ARG ARCH
ARG BINARY_NAME

ENV GOOS=${OS_NAME:-linux} \
    GOARCH=${ARCH:-amd64} \
    APP_NAME=${BINARY_NAME}

WORKDIR /app

COPY .env .env
COPY ./config/config-docker.yaml ./config/config-docker.yaml
COPY --from=builder /code/target/bin/${GOOS}_${GOARCH}/${APP_NAME}_1.0.0-SNAPSHOT_${GOOS}_${GOARCH} ./${APP_NAME}

RUN chmod +x ./${APP_NAME}

CMD ./${APP_NAME} -config ./config/config-docker.yaml
