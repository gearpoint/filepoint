package views

import (
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gearpoint/filepoint/pkg/utils"
)

type FileLabelling []map[string][]string

type DynamoDBSchema interface {
	GetKey() (map[string]types.AttributeValue, error)
	GetUpdateFields() expression.UpdateBuilder
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
		"prefix": prefix,
	}, nil
}

func (d DynamoDBUploadSchema) GetUpdateFields() expression.UpdateBuilder {
	uploadType := reflect.TypeOf(d)
	uploadValue := reflect.ValueOf(d)

	isKeyField := func(a string) bool {
		list := []string{"userId", "prefix"}
		for _, b := range list {
			if b == a {
				return true
			}
		}
		return false
	}
	var update expression.UpdateBuilder

	first := true
	for i := 0; i < uploadType.NumField(); i++ {
		field := uploadType.Field(i)
		tag := field.Tag.Get("dynamodbav")

		value := uploadValue.Field(i).Interface()
		if tag != "" && !isKeyField(tag) {
			if first {
				first = false
				update = expression.Set(expression.Name(tag), expression.Value(value))
				continue
			}
			update.Set(expression.Name(tag), expression.Value(value))
		}
	}

	return update
}
