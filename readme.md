# Filepoint

Filepoint is the Gearpoint's file manager service. It's built for performance.

## Useful Links

- [Golang](https://go.dev/)
- [Gin Framework](https://gin-gonic.com/)
- [Zap Logger](https://github.com/uber-go/zap)
- [Zap Logger Gin Middleware](https://github.com/gin-contrib/zap)
- [Swagger](https://swagger.io/)
- [Gin Swagger](https://github.com/swaggo/gin-swagger)
- [Apache Kafka](https://kafka.apache.org/get-started)
- [Watermill](https://github.com/ThreeDotsLabs/watermill)
- [OpenTelemetry](https://opentelemetry.io/)
- [OpenTelemetry for Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Redis](https://redis.io/)
- [Redis Go Client](https://redis.io/docs/clients/go/)
- [AWS S3](https://docs.aws.amazon.com/s3/)
- [AWS SDK for Go](https://docs.aws.amazon.com/sdk-for-go/)

## Tech Stack

- Golang
- Gin Framework REST API
    - Zap Logger
    - Gin Swagger
    - Watermill
- Apache Kafka for queueu proccess
- AWS S3 storage service
- Event driven architecture

## Building and running

First of all, set the environment file:
```sh
cp .env.example .env
```

### Running in Go

To run the project with go, use the following command:

```sh
go run cmd/filepoint/main.go -config ./config/config-local.yaml
```


### The binary

To build the binary, use the available command in Makefile:

```sh
make build-local
```

### Docker image

For the Docker image, use this Makefile command:

```sh
make build-image
```

### Docker Compose

Just run:

```sh
docker compose build && docker compose up
```