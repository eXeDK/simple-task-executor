package task

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/sqs"
	uuid "github.com/satori/go.uuid"
	"github.com/simple-task-executor/golang/target"
)

type Executor interface {
	Handle(config string) interface{}
}

type Result struct {
	ID        string      `json:"id"`
	Timestamp int64       `json:"timestamp"`
	TargetID  string      `json:"targetId"`
	Result    interface{} `json:"result"`
}

func Schedule(sqsSession *sqs.SQS, queueUrl string, targetID string, targetConfig target.Config) error {
	delay := int64(targetConfig.Delay)

	sendMessageRequest := sqs.SendMessageInput{
		QueueUrl:     aws.String(queueUrl),
		DelaySeconds: aws.Int64(delay),
		MessageBody:  aws.String("{}"),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"TargetId": {
				DataType:    aws.String("String"),
				StringValue: aws.String(targetID),
			},
		},
	}

	_, err := sqsSession.SendMessage(&sendMessageRequest)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to send message to queue '%s' for target: '%s', error: %+v", queueUrl, targetID, err))
	}

	return nil
}

func SaveResult(dynamodbSession *dynamodb.DynamoDB, targetID string, taskResult interface{}) error {
	if taskResult == nil {
		fmt.Print("TaskResult is nil, do not save it")
		return nil
	}

	// Prepare item
	item := Result{
		ID:        uuid.NewV4().String(),
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
		TargetID:  targetID,
		Result:    taskResult,
	}
	marshaledItem, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		// TODO: More handling?
		return err
	}

	putRequest := dynamodb.PutItemInput{
		TableName: aws.String("TargetTaskResults"),
		Item:      marshaledItem,
	}

	_, err = dynamodbSession.PutItem(&putRequest)
	return err
}
