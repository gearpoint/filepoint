# Filepoint

Filepoint is the Gearpoint's file management service. It's built for performance and high availability.

## Tech Stack

- [Golang](https://go.dev/)
- [Gin Framework](https://gin-gonic.com/)
- **Logs and tracing**
  - [Zap Logger](https://github.com/uber-go/zap)
  - [Zap Logger Gin Middleware](https://github.com/gin-contrib/zap)
- **Pub/Sub and Kafka**
  - [Apache Kafka](https://kafka.apache.org/get-started)
  - [Sarama Apache Kafka Go Library](https://github.com/IBM/sarama)
  - [Watermill](https://github.com/ThreeDotsLabs/watermill)
  - [Watermill Kafka Pub/Sub Implementation](https://github.com/ThreeDotsLabs/watermill-kafka)
  - [Watermill Http Pub/Sub Implementation](https://github.com/ThreeDotsLabs/watermill-http)
- **Caching**
  - [Redis](https://redis.io/)
  - [Redis client for Go](https://github.com/redis/go-redis)
- **Docs**
  - [Swagger](https://swagger.io/)
  - [Gin Swagger](https://github.com/swaggo/gin-swagger)
- **AWS**
  - [AWS S3](https://docs.aws.amazon.com/s3/)
  - [AWS DynamoDB](https://docs.aws.amazon.com/dynamodb/)
  - [AWS Cloudfront](https://aws.amazon.com/cloudfront/)
  - [AWS Rekognition](https://aws.amazon.com/rekognition/)
- **File Handling**
  - [libvips](https://github.com/libvips/libvips)
  - [bimg](https://github.com/h2non/bimg)

## The project

![Project](./docs/Filepoint.drawio.png)

<br>

## Building and running

First of all, set the environment file:

```sh
cp .env.example .env
```

Then, you need to setup [LocalStack](https://www.localstack.cloud/):

```sh
python3 -m pip install localstack
pip install awscli-local
```

Now, you'll need to setup your environment configuration. Here we'll be using the available Docker Compose configuration:

1. Start by running the required services.

    ```sh
    make run-services
    ```

2. Configure the Webhook listener:
    First, change the {{ your_unique_id }} value in your config files to your actual webhook endpoint.

    You can use the local webhooks.site deploy at [localhost:8084](http://localhost:8084) or any other webhooks listener.

    To get your unique ID, execute as following:

    ```sh
    curl --request POST --url 'http://localhost:8084/token'
    ```

    Files to change:

      - config/config.yml - For containerized execution
      - config/config-local.yml - For terminal execution

3. [optional] If any problems happened in the Localstack initialization, you can manually run the script:

    ```sh
    bash ./scripts/init-aws.sh
    ```

4. Build the base image.

    To make faster builds, Filepoint uses a base image that contains all necessary libs installed.
    You will need it when trying to build the images (using Makefile or Docker Compose).

    ```sh
    make build-base
    ```

    When you're using a production app, you can also publish this image to the registry with ```make publish-base```.

### Running in terminal

To run the project without Docker, use the following command:

```sh
# Filepoint API
make run

# Filepoint webhooks provider
make run-webhooks-provider
```

Pay attention that you must have the following services running:

- Redis
- LocalStack

If you don't want to use containerization, you will need to setup them manually.

### The binary

To build the binary, use the available command in Makefile:

```sh
make build-binary
```

### Building Docker image

For the Docker images, use this Makefile command:

```sh
make build-images
```

### Using Docker Compose

Just run the Docker Compose:

```sh
docker compose build && docker compose up
```

<br>

## Signing Cloudfront URLs

To sign the Cloudfront URLs, you must provide a valid private key in PEM format. The public key has to be registered in the Cloudfront configuration as well.

You can check it in the [docs](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-trusted-signers.html).

The key must be in the ```.aws``` folder.

> Local setups don't need signed URLs. It'll only work with proper AWS configuration.

<br>

## Docs

The repository docs include this readme, OpenAPI and Godoc.
The documentation files can be saved in /docs.

### Creating API documentation

To create API docs, you must have Swaggo installed

```sh
go install github.com/swaggo/swag/cmd/swag@latest
```

Then, simply run ```make swagger```

### Accessing Godoc

To access the Go Documentation Server, you need Godoc installed.

```sh
  go install -v golang.org/x/tools/cmd/godoc@latest
```

Then, you can run with ```make godoc```

<br>

## Setting up the Container Registry

You can use any container registry for this project.
For this to work across the repo, you will need to change the {{ REPOSITORY }} variables to your current registry, as the default value is "localhost". See Makefile and .env.example.
We will show an example of configuration with AWS ECR being used.

> This is only necessary when building and publishing images from local machine.

### Configuring

1. Configure [aws cli](https://aws.amazon.com/cli/):

    ```sh
    aws configure # use your credentials here
    ```

2. Authenticate in Docker:

    For this step, use Docker GUI or docker login command.

    Then, execute the following:

    ```sh
    aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin {{ REGISTRY }}
    ```

    Now, you should see "Login Succeeded"

### Pulling

To pull from ECR private repositories, you must first be authenticated in AWS:

docker login
docker pull gearpoint/filepoint
docker pull gearpoint/filepoint-webhooks-sender

```sh
docker pull {{ REGISTRY }}/prod-filepoint-repo
docker pull {{ REGISTRY }}/prod-filepoint-webhooks-repo
```

### Publishing

To publish to Docker Hub, you must build, tag and push your images:

```sh
make build-base # only first time
make build-images
make publish-filepoint
make publish-webhooks-sender
```
