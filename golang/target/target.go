package target

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Config struct {
	Active     bool
	Config     string
	Delay      int
	TargetType Type
}

type Type string

const (
	TypeCertCheck    Type = "certCheck"
	TypeHTTPPing     Type = "httpPing"
	TypeSLLLabsCheck Type = "sslLabsCheck"
)

func GetConfig(dynamoDbSession *dynamodb.DynamoDB, targetID string) (*Config, error) {
	dynamodbParams := dynamodb.GetItemInput{
		TableName: aws.String("Targets"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(targetID),
			},
		},
		ProjectionExpression: aws.String("active, config, delay, targetType"),
	}

	dynamodbOutput, err := dynamoDbSession.GetItem(&dynamodbParams)
	if err != nil {
		fmt.Printf("getTargetConfig error: %+v", err)
		return nil, errors.New(fmt.Sprintf("An error occured during the call to dynamodb for target: '%s': %+v", targetID, err))
	}

	if dynamodbOutput.Item == nil {
		return nil, errors.New(fmt.Sprintf("Unable to find config for target: '%s'", targetID))
	}

	resultTargetConfig := Config{}
	err = dynamodbattribute.UnmarshalMap(dynamodbOutput.Item, &resultTargetConfig)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to unmarshal config for target: '%s', item: %+v", targetID, dynamodbOutput.Item))
	}

	return &resultTargetConfig, nil
}
