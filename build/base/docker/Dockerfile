FROM golang:1.22-alpine3.19

# Install required repos
RUN apk add --update make \
    bash \
    build-base \
    musl-dev \
    --repository http://dl-3.alpinelinux.org/alpine/edge/community \
    --repository http://dl-3.alpinelinux.org/alpine/edge/main vips-dev
