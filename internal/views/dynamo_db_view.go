package views

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gearpoint/filepoint/pkg/utils"
)

type FileLabelling []map[string][]string

type DynamoDBSchema interface {
	GetKey() (map[string]types.AttributeValue, error)
}

// DynamoDBUploadSchema is the DynamoDB upload schema view.
type DynamoDBUploadSchema struct {
	UserId         string                       `dynamodbav:"userId"`
	Prefix         string                       `dynamodbav:"prefix"`
	Author         string                       `dynamodbav:"author"`
	Title          string                       `dynamodbav:"title"`
	RequestId      string                       `dynamodbav:"requestId"`
	CorrelationId  string                       `dynamodbav:"correlationId"`
	DefinitionsMap utils.FileDefinitionsMapping `dynamodbav:"definitionsMap"`
	FileLabels     FileLabelling                `dynamodbav:"fileLabels"`
	OccurredOn     time.Time                    `dynamodbav:"occurredOn"`
}

func (d DynamoDBUploadSchema) GetKey() (map[string]types.AttributeValue, error) {
	userId, err := attributevalue.Marshal(d.UserId)
	if err != nil {
		return nil, err
	}

	prefix, err := attributevalue.Marshal(d.Prefix)
	if err != nil {
		return nil, err
	}

	return map[string]types.AttributeValue{
		"userId": userId,
		"year":   prefix,
	}, nil
}
