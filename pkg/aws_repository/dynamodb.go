package aws_repository

import (
	"errors"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/logger"
	"go.uber.org/zap"
)

// TableExists determines whether a DynamoDB table exists.
func (r *AWSRepository) TableExists(tableName string) (bool, error) {
	_, err := r.dynamoClient.DescribeTable(
		r.ctx, &dynamodb.DescribeTableInput{
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

	if response.Item == nil {
		return errors.New("table row not found")
	}

	err = attributevalue.UnmarshalMap(response.Item, &schema)

	return err
}

// AddTableRow adds a new row to the DynamoDB table.
func (r *AWSRepository) AddTableRow(tableName string, schema views.DynamoDBSchema) error {
	item, err := attributevalue.MarshalMap(schema)
	if err != nil {
		return errors.New("error reading table row info")
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

	expr, err := expression.NewBuilder().WithUpdate(
		schema.GetUpdateFields(),
	).Build()

	if err != nil {
		return err
	}

	_, err = r.dynamoClient.UpdateItem(r.ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(tableName),
		Key:                       key,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	})

	return err
}

// DelTableRow removes a row from the DynamoDB table.
func (r *AWSRepository) DelTableRow(tableName string, schema views.DynamoDBSchema) error {
	key, err := schema.GetKey()
	if err != nil {
		return err
	}

	_, err = r.dynamoClient.DeleteItem(r.ctx, &dynamodb.DeleteItemInput{
		Key: key, TableName: aws.String(tableName),
	})

	return err
}

// DelTableRow removes a whole partition from the DynamoDB table.
func (r *AWSRepository) DelTablePartition(tableName string, partitionKey string, partitionValue string, schema views.DynamoDBSchema) error {
	keyEx := expression.Key(partitionKey).Equal(expression.Value(partitionValue))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return err
	}

	queryPaginator := dynamodb.NewQueryPaginator(r.dynamoClient, &dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})

	var wg sync.WaitGroup

	for queryPaginator.HasMorePages() {
		res, err := queryPaginator.NextPage(r.ctx)
		if err != nil {
			logger.Error("Unable to search page of partitions",
				zap.Error(err),
			)
			continue
		}
		for _, pageItem := range res.Items {
			var itemSchema = schema

			err = attributevalue.UnmarshalMap(pageItem, &itemSchema)
			if err != nil {
				logger.Error("Unable to unmarshal page for deletion",
					zap.Error(err),
				)
				continue
			}
			wg.Add(1)
			go func(itemSchema views.DynamoDBSchema) {
				key, err := itemSchema.GetKey()
				if err != nil {
					return
				}
				defer wg.Done()
				_, err = r.dynamoClient.DeleteItem(r.ctx, &dynamodb.DeleteItemInput{
					Key: key, TableName: aws.String(tableName),
				})
				if err != nil {
					logger.Error("Unable to delete item",
						zap.String("tableName", tableName),
						zap.Any("tableKey", key),
						zap.Error(err),
					)
				}
			}(itemSchema)
		}
	}
	wg.Wait()

	return nil
}
