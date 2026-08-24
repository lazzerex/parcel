package store

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBAPI interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

func partitionKey(fileID string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: "FILE#" + fileID},
	}
}

func GetStatus(ctx context.Context, client DynamoDBAPI, table, fileID string) (string, error) {
	out, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key:       partitionKey(fileID),
	})
	if err != nil {
		return "", err
	}
	if out.Item == nil {
		return "", nil
	}

	status, ok := out.Item["status"].(*types.AttributeValueMemberS)
	if !ok {
		return "", nil
	}
	return status.Value, nil
}

func SetStatus(ctx context.Context, client DynamoDBAPI, table, fileID, status string) error {
	_, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:        aws.String(table),
		Key:              partitionKey(fileID),
		UpdateExpression: aws.String("SET #status = :status, updated_at = :updated_at"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":     &types.AttributeValueMemberS{Value: status},
			":updated_at": &types.AttributeValueMemberS{Value: now()},
		},
		ConditionExpression: aws.String("attribute_exists(PK)"),
	})
	return err
}

func SetCompleted(ctx context.Context, client DynamoDBAPI, table, fileID string, size int64, sha256, contentType string) error {
	_, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(table),
		Key:       partitionKey(fileID),
		UpdateExpression: aws.String(
			"SET #status = :status, content_type = :content_type, #size = :size, sha256 = :sha256, updated_at = :updated_at, processed_at = :processed_at",
		),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
			"#size":   "size",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":       &types.AttributeValueMemberS{Value: "COMPLETED"},
			":content_type": &types.AttributeValueMemberS{Value: contentType},
			":size":         &types.AttributeValueMemberN{Value: strconv.FormatInt(size, 10)},
			":sha256":       &types.AttributeValueMemberS{Value: sha256},
			":updated_at":   &types.AttributeValueMemberS{Value: now()},
			":processed_at": &types.AttributeValueMemberS{Value: now()},
		},
		ConditionExpression: aws.String("attribute_exists(PK)"),
	})
	return err
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}
