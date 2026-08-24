package store

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type fakeDynamoDB struct {
	getItemOutput   *dynamodb.GetItemOutput
	getItemErr      error
	updateItemInput *dynamodb.UpdateItemInput
	updateItemErr   error
}

func (f *fakeDynamoDB) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return f.getItemOutput, f.getItemErr
}

func (f *fakeDynamoDB) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	f.updateItemInput = params
	return &dynamodb.UpdateItemOutput{}, f.updateItemErr
}

func TestGetStatusReturnsStatusWhenPresent(t *testing.T) {
	client := &fakeDynamoDB{
		getItemOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"status": &types.AttributeValueMemberS{Value: "COMPLETED"},
			},
		},
	}

	status, err := GetStatus(context.Background(), client, "parcel-metadata", "abc123")
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status != "COMPLETED" {
		t.Fatalf("GetStatus() = %q, want %q", status, "COMPLETED")
	}
}

func TestGetStatusReturnsEmptyWhenMissing(t *testing.T) {
	client := &fakeDynamoDB{getItemOutput: &dynamodb.GetItemOutput{}}

	status, err := GetStatus(context.Background(), client, "parcel-metadata", "abc123")
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status != "" {
		t.Fatalf("GetStatus() = %q, want empty", status)
	}
}

func TestSetStatusUsesPartitionKeyAndCondition(t *testing.T) {
	client := &fakeDynamoDB{}

	if err := SetStatus(context.Background(), client, "parcel-metadata", "abc123", "PROCESSING"); err != nil {
		t.Fatalf("SetStatus() error = %v", err)
	}

	key := client.updateItemInput.Key["PK"].(*types.AttributeValueMemberS).Value
	if key != "FILE#abc123" {
		t.Errorf("Key[PK] = %q, want %q", key, "FILE#abc123")
	}
	status := client.updateItemInput.ExpressionAttributeValues[":status"].(*types.AttributeValueMemberS).Value
	if status != "PROCESSING" {
		t.Errorf("status = %q, want %q", status, "PROCESSING")
	}
}

func TestSetCompletedWritesAllFields(t *testing.T) {
	client := &fakeDynamoDB{}

	if err := SetCompleted(context.Background(), client, "parcel-metadata", "abc123", 1024, "deadbeef", "image/png"); err != nil {
		t.Fatalf("SetCompleted() error = %v", err)
	}

	values := client.updateItemInput.ExpressionAttributeValues
	if got := values[":status"].(*types.AttributeValueMemberS).Value; got != "COMPLETED" {
		t.Errorf("status = %q, want COMPLETED", got)
	}
	if got := values[":size"].(*types.AttributeValueMemberN).Value; got != "1024" {
		t.Errorf("size = %q, want 1024", got)
	}
	if got := values[":sha256"].(*types.AttributeValueMemberS).Value; got != "deadbeef" {
		t.Errorf("sha256 = %q, want deadbeef", got)
	}
	if got := values[":content_type"].(*types.AttributeValueMemberS).Value; got != "image/png" {
		t.Errorf("content_type = %q, want image/png", got)
	}

	// "size" is a DynamoDB reserved keyword and must be aliased.
	if got := client.updateItemInput.ExpressionAttributeNames["#size"]; got != "size" {
		t.Errorf("ExpressionAttributeNames[#size] = %q, want %q", got, "size")
	}
}
