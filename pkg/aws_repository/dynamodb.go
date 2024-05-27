package aws_repository

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gearpoint/filepoint/internal/views"
)

// TableExists determines whether a DynamoDB table exists.
func (r *AWSRepository) TableExists(tableName string) (bool, error) {
	_, err := r.dynamoClient.DescribeTable(
		context.TODO(), &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		},
	)

	if err == nil {
		return true, nil
	}

	var notFoundEx *types.ResourceNotFoundException
	if errors.As(err, &notFoundEx) {
		return false, nil
	}

	return false, err
}

// GetTableRow gets row data from the DynamoDB table by using the primary composite key.
func (r *AWSRepository) GetTableRow(tableName string, schema views.DynamoDBSchema) error {
	key, err := schema.GetKey()
	if err != nil {
		return err
	}

	response, err := r.dynamoClient.GetItem(r.ctx, &dynamodb.GetItemInput{
		Key: key, TableName: aws.String(tableName),
	})

	if err != nil {
		return err
	}

	err = attributevalue.UnmarshalMap(response.Item, &schema)

	return err
}

// AddTableRow adds a new row to the DynamoDB table.
func (r *AWSRepository) AddTableRow(tableName string, schema views.DynamoDBSchema) error {
	item, err := attributevalue.MarshalMap(schema)
	if err != nil {
		panic(err)
	}
	_, err = r.dynamoClient.PutItem(r.ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		return errors.New("error adding table row")
	}

	return err
}

// UpdateTableRow adds a new row to the DynamoDB table.
func (r *AWSRepository) UpdateTableRow(tableName string, schema views.DynamoDBSchema) error {
	key, err := schema.GetKey()
	if err != nil {
		return err
	}

	// todo: fix logic

	// var response *dynamodb.UpdateItemOutput
	// var attributeMap map[string]map[string]interface{}

	// update := expression.Set(expression.Name("info.rating"), expression.Value(schema.Info["rating"]))
	// update.Set(expression.Name("info.plot"), expression.Value(movie.Info["plot"]))
	// expr, err := expression.NewBuilder().WithUpdate(update).Build()

	// if err != nil {
	// 	return err
	// }

	_, err = r.dynamoClient.UpdateItem(r.ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key:       key,
		// ExpressionAttributeNames:  expr.Names(),
		// ExpressionAttributeValues: expr.Values(),
		// UpdateExpression:          expr.Update(),
		ReturnValues: types.ReturnValueUpdatedNew,
	})

	return err
}
