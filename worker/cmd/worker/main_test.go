package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type fakeS3 struct {
	body     []byte
	err      error
	getCalls int
}

func (f *fakeS3) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	f.getCalls++
	if f.err != nil {
		return nil, f.err
	}
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(f.body))}, nil
}

type fakeDynamoDB struct {
	status        string
	getItemErr    error
	updateItemErr error
	getItemCalls  int
	statusWrites  []string
}

func (f *fakeDynamoDB) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	f.getItemCalls++
	if f.getItemErr != nil {
		return nil, f.getItemErr
	}
	if f.status == "" {
		return &dynamodb.GetItemOutput{}, nil
	}
	return &dynamodb.GetItemOutput{
		Item: map[string]types.AttributeValue{
			"status": &types.AttributeValueMemberS{Value: f.status},
		},
	}, nil
}

func (f *fakeDynamoDB) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	status := params.ExpressionAttributeValues[":status"].(*types.AttributeValueMemberS).Value
	f.statusWrites = append(f.statusWrites, status)
	if f.updateItemErr != nil {
		return nil, f.updateItemErr
	}
	return &dynamodb.UpdateItemOutput{}, nil
}

func validJobBody() string {
	return `{"job_id":"j1","file_id":"f1","bucket":"parcel-files","key":"uploads/f1/x.bin","operation":"inspect"}`
}

func TestHandleRecordSkipsWhenAlreadyCompleted(t *testing.T) {
	s3c := &fakeS3{}
	ddb := &fakeDynamoDB{status: "COMPLETED"}
	w := worker{s3Client: s3c, dynamodbClient: ddb, table: "parcel-metadata"}

	err := w.handleRecord(context.Background(), events.SQSMessage{Body: validJobBody()})
	if err != nil {
		t.Fatalf("handleRecord() error = %v", err)
	}
	if s3c.getCalls != 0 {
		t.Errorf("GetObject called %d times, want 0", s3c.getCalls)
	}
	if len(ddb.statusWrites) != 0 {
		t.Errorf("UpdateItem called %d times, want 0", len(ddb.statusWrites))
	}
}

func TestHandleRecordInvalidJobReturnsError(t *testing.T) {
	ddb := &fakeDynamoDB{}
	w := worker{s3Client: &fakeS3{}, dynamodbClient: ddb, table: "parcel-metadata"}

	err := w.handleRecord(context.Background(), events.SQSMessage{Body: "not json"})
	if err == nil {
		t.Fatal("handleRecord() error = nil, want error")
	}
	if ddb.getItemCalls != 0 {
		t.Errorf("GetItem called %d times, want 0", ddb.getItemCalls)
	}
}

func TestHandleRecordUnknownOperationReturnsError(t *testing.T) {
	ddb := &fakeDynamoDB{}
	w := worker{s3Client: &fakeS3{}, dynamodbClient: ddb, table: "parcel-metadata"}

	body := `{"job_id":"j1","file_id":"f1","bucket":"parcel-files","key":"uploads/f1/x.bin","operation":"compress"}`
	err := w.handleRecord(context.Background(), events.SQSMessage{Body: body})
	if err == nil {
		t.Fatal("handleRecord() error = nil, want error")
	}
	if ddb.getItemCalls != 0 {
		t.Errorf("GetItem called %d times, want 0 (dispatch should fail before status check)", ddb.getItemCalls)
	}
}

func TestHandleRecordSuccessWritesCompleted(t *testing.T) {
	s3c := &fakeS3{body: []byte("parcel test content")}
	ddb := &fakeDynamoDB{}
	w := worker{s3Client: s3c, dynamodbClient: ddb, table: "parcel-metadata"}

	err := w.handleRecord(context.Background(), events.SQSMessage{Body: validJobBody()})
	if err != nil {
		t.Fatalf("handleRecord() error = %v", err)
	}
	want := []string{"PROCESSING", "COMPLETED"}
	if len(ddb.statusWrites) != len(want) {
		t.Fatalf("statusWrites = %v, want %v", ddb.statusWrites, want)
	}
	for i, w := range want {
		if ddb.statusWrites[i] != w {
			t.Errorf("statusWrites[%d] = %q, want %q", i, ddb.statusWrites[i], w)
		}
	}
}

func TestHandleRecordProcessorFailureSetsFailedStatus(t *testing.T) {
	s3c := &fakeS3{err: errors.New("no such key")}
	ddb := &fakeDynamoDB{}
	w := worker{s3Client: s3c, dynamodbClient: ddb, table: "parcel-metadata"}

	err := w.handleRecord(context.Background(), events.SQSMessage{Body: validJobBody()})
	if err == nil {
		t.Fatal("handleRecord() error = nil, want error")
	}
	want := []string{"PROCESSING", "FAILED"}
	if len(ddb.statusWrites) != len(want) {
		t.Fatalf("statusWrites = %v, want %v", ddb.statusWrites, want)
	}
	for i, w := range want {
		if ddb.statusWrites[i] != w {
			t.Errorf("statusWrites[%d] = %q, want %q", i, ddb.statusWrites[i], w)
		}
	}
}

func TestHandleRequestStopsOnFirstError(t *testing.T) {
	s3c := &fakeS3{err: errors.New("no such key")}
	ddb := &fakeDynamoDB{}
	w := worker{s3Client: s3c, dynamodbClient: ddb, table: "parcel-metadata"}

	event := events.SQSEvent{Records: []events.SQSMessage{
		{Body: validJobBody()},
		{Body: validJobBody()},
	}}

	err := w.handleRequest(context.Background(), event)
	if err == nil {
		t.Fatal("handleRequest() error = nil, want error")
	}
	if s3c.getCalls != 1 {
		t.Errorf("GetObject called %d times, want 1 (should stop after first failure)", s3c.getCalls)
	}
}
