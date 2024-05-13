ARG REGISTRY

FROM ${REGISTRY}/prod-filepoint-base-repo:latest as builder

ARG OS_NAME
ARG ARCH
ARG BINARY_NAME
ARG CONFIG_FILE

ENV GOOS=${OS_NAME:-linux} \
    GOARCH=${ARCH:-amd64}

WORKDIR /code

COPY . .

RUN make deps

RUN echo ${GOOS}/${GOARCH} > scripts/plataforms \
    && make clean test build-binary

# Start a new stage from scratch
FROM ${REGISTRY}/prod-filepoint-base-repo:latest

LABEL maintainer="Gearpoint" \
 contributor="Luan Baggio <luanbaggio0@gmail.com>"

ARG OS_NAME
ARG ARCH
ARG BINARY_NAME
ARG CONFIG_FILE

ENV GOOS=${OS_NAME:-linux} \
    GOARCH=${ARCH:-amd64} \
    APP_NAME=${BINARY_NAME}

WORKDIR /app

COPY .env .env
COPY .aws .aws
COPY ${CONFIG_FILE} ./config/config.yaml
COPY --from=builder /code/target/bin/${GOOS}_${GOARCH}/${APP_NAME}_1.0.0-SNAPSHOT_${GOOS}_${GOARCH} ./${APP_NAME}

RUN chmod +x ./${APP_NAME}

CMD ./${APP_NAME} -config ./config/config.yaml
