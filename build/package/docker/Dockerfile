FROM golang:1.18 as modules

ARG OS_NAME
ARG ARCH
ARG BINARY_NAME

ENV GOOS=${OS_NAME:-linux} \
    GOARCH=${ARCH:-amd64}

WORKDIR /code

COPY Makefile go.mod go.sum ./

RUN make deps

# Start from the latest golang base image
FROM golang:1.18 as builder


ARG OS_NAME
ARG ARCH
ARG BINARY_NAME

ENV GOOS=${OS_NAME:-linux} \
    GOARCH=${ARCH:-amd64}

WORKDIR /code
COPY .  .

RUN echo ${GOOS}/${GOARCH} > scripts/plataforms \
    && make clean test build-local



######## Start a new stage from scratch #######
FROM alpine:latest

RUN apk --no-cache add ca-certificates

LABEL maintainer="Bling" \
 contributor="Luan Baggio <luan.baggio@bling.com.br>"

ARG OS_NAME
ARG ARCH
ARG BINARY_NAME

ENV GOOS=${OS_NAME:-linux} \
 GOARCH=${ARCH:-amd64} \
 APP_NAME=${BINARY_NAME}

WORKDIR /app/

COPY --from=builder /code/target/bin/${GOOS}_${GOARCH}/${APP_NAME}_*_${GOOS}_${GOARCH}  ./${APP_NAME}
COPY ./configs/application.yaml ./configs/application.yaml

CMD ./${APP_NAME} -config ./configs/application.yml

ENTRYPOINT ["./bluebird"]
