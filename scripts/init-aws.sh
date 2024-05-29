#!/bin/bash

awslocal s3 mb s3://filepoint-us

awslocal sqs create-queue \
    --queue-name filepoint_upload_queueing

awslocal dynamodb create-table \
     --table-name filepoint_upload \
     --key-schema \
          AttributeName=userId,KeyType=HASH \
          AttributeName=prefix,KeyType=RANGE \
     --attribute-definitions \
          AttributeName=userId,AttributeType=S \
          AttributeName=prefix,AttributeType=S \
     --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
          # AttributeName=requestId,AttributeType=S \
          # AttributeName=correlationId,AttributeType=S \
          # AttributeName=definitionsMap,AttributeType=B \
          # AttributeName=fileLabels,AttributeType=B \