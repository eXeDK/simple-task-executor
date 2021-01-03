package main

import (
	"context"
	"fmt"
	"os"

	"github.com/simple-task-executor/golang/taskHandler/taskExecutor"

	"github.com/simple-task-executor/golang/target"
	"github.com/simple-task-executor/golang/task"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

var dynamoDbSession *dynamodb.DynamoDB
var sqsSession *sqs.SQS
var queueUrl string

func init() {
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	dynamoDbSession = dynamodb.New(awsSession)
	sqsSession = sqs.New(awsSession)
	queueUrl = os.Getenv("QUEUE_URL")
}

func Handler(ctx context.Context, sqsEvent events.SQSEvent) (Response, error) {
	targetID := getTargetID(sqsEvent)
	isTestInvoke := getTestInvoke(sqsEvent)

	// Get Target config
	targetConfig, err := target.GetConfig(dynamoDbSession, targetID)
	if err != nil || targetConfig == nil {
		// TODO: Handle this
	}

	// If we are not active, stop the task
	if !targetConfig.Active && !isTestInvoke {
		fmt.Print("Task stopped - target not active")
		return Response{}, nil
	}

	// Set up defer
	if !isTestInvoke {
		defer func() {
			err := task.Schedule(sqsSession, queueUrl, targetID, *targetConfig)
			if err != nil {
				fmt.Printf("Unable to schedule new task")
			}
		}()
	}

	// Execute logic for target
	taskResult := taskExecutor.ExecuteTask(*targetConfig)

	// Save task response
	err = task.SaveResult(dynamoDbSession, targetID, taskResult)
	if err != nil {
		fmt.Print("Unable to save task result")
		return Response{}, nil
	}

	return Response{}, nil
}

func main() {
	lambda.Start(Handler)
}

func getTargetID(event events.SQSEvent) string {
	return *event.Records[0].MessageAttributes["TargetId"].StringValue
}

func getTestInvoke(event events.SQSEvent) bool {
	if val, ok := event.Records[0].MessageAttributes["TestInvoke"]; ok {
		return *val.StringValue == "true"
	}

	return false
}
