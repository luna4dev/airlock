package service

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBService struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBService(tableName string) (*DynamoDBService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(cfg)

	return &DynamoDBService{
		client:    client,
		tableName: tableName,
	}, nil
}
