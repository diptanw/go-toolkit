package awssqs

import (
	"context"
	"fmt"

	"github.com/diptanw/go-toolkit/message"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSReceiver is a struct that handles SQS queue comunication.
type SQSReceiver struct {
	client *sqs.Client
	input  *sqs.ReceiveMessageInput
}

// NewSQSReceiver creates a new instance of SQSReceiver.
func NewSQSReceiver(client *sqs.Client, queueURL string) *SQSReceiver {
	const (
		defaultMaxMsgNum         = 10
		defaultVisibilityTimeout = 60
		defaultWaitTime          = 20
	)

	input := &sqs.ReceiveMessageInput{
		AttributeNames: []sqs.QueueAttributeName{
			sqs.QueueAttributeNameCreatedTimestamp,
		},
		MessageAttributeNames: []string{
			"All",
		},
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: aws.Int64(defaultMaxMsgNum),
		VisibilityTimeout:   aws.Int64(defaultVisibilityTimeout),
		WaitTimeSeconds:     aws.Int64(defaultWaitTime),
	}

	return &SQSReceiver{
		client: client,
		input:  input,
	}
}

// Receive requests the SQS messages available in the queue.
func (r *SQSReceiver) Receive(ctx context.Context) ([]message.Message, error) {
	req := r.client.ReceiveMessageRequest(r.input)

	resp, err := req.Send(ctx)
	if err != nil {
		return nil, fmt.Errorf("receiving messages: %w", err)
	}

	ms := make([]message.Message, len(resp.Messages))

	for i, m := range resp.Messages {
		ms[i] = message.Message{
			ID:   aws.StringValue(m.MessageId),
			Data: []byte(aws.StringValue(m.Body)),
			Ack:  r.ack(m),
		}
	}

	return ms, nil
}

func (r *SQSReceiver) ack(m sqs.Message) func(ctx context.Context, ok bool) error {
	return func(ctx context.Context, ok bool) (err error) {
		if ok {
			req := r.client.DeleteMessageRequest(&sqs.DeleteMessageInput{
				QueueUrl:      r.input.QueueUrl,
				ReceiptHandle: m.ReceiptHandle,
			})

			_, err = req.Send(ctx)
		}

		return err
	}
}

// GetQueueURL returns the queue URL for a given queue name.
func GetQueueURL(ctx context.Context, client *sqs.Client, queueName string) (string, error) {
	params := &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	}

	req := client.GetQueueUrlRequest(params)

	resp, err := req.Send(ctx)
	if err != nil {
		return "", fmt.Errorf("getting queue url: %w", err)
	}

	return aws.StringValue(resp.QueueUrl), nil
}
