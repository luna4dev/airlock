package model

import "time"

type EmailAuth struct {
	Token     string `json:"token" dynamodbav:"token"`
	SentAt    int64  `json:"sentAt" dynamodbav:"sentAt"`
	Completed bool   `json:"completed" dynamodbav:"completed"`
}

type UserPreferences struct {
	// TODO: add preference schema
}

type User struct {
	ID          string           `json:"id" dynamodbav:"id"`
	Email       string           `json:"email" dynamodbav:"email"`
	Status      UserStatus       `json:"status" dynamodbav:"status"`
	Preferences *UserPreferences `json:"preferences,omitempty" dynamodbav:"preferences,omitempty"`
	CreatedAt   int64            `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt   int64            `json:"updatedAt" dynamodbav:"updatedAt"`
	LastLoginAt *int64           `json:"lastLoginAt,omitempty" dynamodbav:"lastLoginAt,omitempty"`
	EmailAuth   *EmailAuth       `json:"emailAuth,omitempty" dynamodbav:"emailAuth,omitempty"`
}

func (u *User) SetCreatedAt() {
	u.CreatedAt = time.Now().UnixMilli()
	u.UpdatedAt = u.CreatedAt
}

func (u *User) SetUpdatedAt() {
	u.UpdatedAt = time.Now().UnixMilli()
}

func (u *User) SetLastLoginAt() {
	now := time.Now().UnixMilli()
	u.LastLoginAt = &now
}
