#!/bin/bash

awslocal s3 mb s3://filepoint-us

awslocal sqs create-queue \
    --queue-name filepoint_upload_queueing

awslocal dynamodb create-table \
   --table-name filepoint_upload \
   --attribute-definitions \
        AttributeName=userId,AttributeType=S \
        AttributeName=prefix,AttributeType=S \
        AttributeName=requestId,AttributeType=S \
        AttributeName=correlationId,AttributeType=S \
        AttributeName=definitionsMap,AttributeType=M \
        AttributeName=fileLabels,AttributeType=M \
   --key-schema \
        AttributeName=id,KeyType=HASH \
        AttributeName=prefix,KeyType=RANGE \

   --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5