package service

import (
	"context"
	"strconv"
	"time"

	"github.com/luna4dev/airlock/internal/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type UserService struct {
	dynamoDBService *DynamoDBService
}

func NewUserService() (*UserService, error) {
	dbService, err := NewDynamoDBService("Luna4Users")
	if err != nil {
		return nil, err
	}

	return &UserService{
		dynamoDBService: dbService,
	}, nil
}

func (u *UserService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	emailIndex := "email-index"

	input := &dynamodb.QueryInput{
		TableName:              aws.String(u.dynamoDBService.tableName),
		IndexName:              aws.String(emailIndex),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
		},
	}

	result, err := u.dynamoDBService.client.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var user model.User
	err = attributevalue.UnmarshalMap(result.Items[0], &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *UserService) UpdateUserEmailAuth(ctx context.Context, userID string, tokenHash string) error {
	now := time.Now().UnixMilli()

	emailAuth := &model.EmailAuth{
		Token:     tokenHash,
		SentAt:    now,
		Completed: false,
	}

	emailAuthValue, err := attributevalue.Marshal(emailAuth)
	if err != nil {
		return err
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(u.dynamoDBService.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: userID},
		},
		UpdateExpression: aws.String("SET emailAuth = :emailAuth, updatedAt = :updatedAt"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":emailAuth": emailAuthValue,
			":updatedAt": &types.AttributeValueMemberN{Value: strconv.FormatInt(now, 10)},
		},
	}

	_, err = u.dynamoDBService.client.UpdateItem(ctx, input)
	return err
}

func (u *UserService) CompleteEmailAuth(ctx context.Context, userID string) error {
	now := time.Now().UnixMilli()

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(u.dynamoDBService.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: userID},
		},
		UpdateExpression: aws.String("SET emailAuth.completed = :completed, lastLoginAt = :lastLoginAt, updatedAt = :updatedAt"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":completed":   &types.AttributeValueMemberBOOL{Value: true},
			":lastLoginAt": &types.AttributeValueMemberN{Value: strconv.FormatInt(now, 10)},
			":updatedAt":   &types.AttributeValueMemberN{Value: strconv.FormatInt(now, 10)},
		},
	}

	_, err := u.dynamoDBService.client.UpdateItem(ctx, input)
	return err
}
