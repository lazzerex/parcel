package notify

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	ebtypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type Event struct {
	JobID     string `json:"job_id"`
	FileID    string `json:"file_id"`
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	Operation string `json:"operation"`
	Status    string `json:"status"`
	Size      int64  `json:"size,omitempty"`
	SHA256    string `json:"sha256,omitempty"`
}

type SNSAPI interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

type EventBridgeAPI interface {
	PutEvents(ctx context.Context, params *eventbridge.PutEventsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error)
}

type Publisher struct {
	snsClient    SNSAPI
	ebClient     EventBridgeAPI
	topicARN     string
	eventBusName string
}

func New(snsClient SNSAPI, ebClient EventBridgeAPI, topicARN, eventBusName string) *Publisher {
	return &Publisher{
		snsClient:    snsClient,
		ebClient:     ebClient,
		topicARN:     topicARN,
		eventBusName: eventBusName,
	}
}

func (p *Publisher) Publish(ctx context.Context, event Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if p.topicARN != "" {
		if _, err := p.snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: aws.String(p.topicARN),
			Message:  aws.String(string(body)),
		}); err != nil {
			return fmt.Errorf("sns publish: %w", err)
		}
	}

	if p.eventBusName != "" {
		if _, err := p.ebClient.PutEvents(ctx, &eventbridge.PutEventsInput{
			Entries: []ebtypes.PutEventsRequestEntry{
				{
					Source:       aws.String("parcel.worker"),
					DetailType:   aws.String("ProcessingCompleted"),
					Detail:       aws.String(string(body)),
					EventBusName: aws.String(p.eventBusName),
				},
			},
		}); err != nil {
			return fmt.Errorf("eventbridge publish: %w", err)
		}
	}

	return nil
}