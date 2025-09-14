package model

import "time"

type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
)

type Luna4User struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Status      UserStatus `json:"status"`
	CreatedAt   int64      `json:"createdAt"`
	UpdatedAt   int64      `json:"updatedAt"`
	LastLoginAt *int64     `json:"lastLoginAt,omitempty"`
}

func (u *Luna4User) SetCreatedAt() {
	u.CreatedAt = time.Now().UnixMilli()
	u.UpdatedAt = u.CreatedAt
}

func (u *Luna4User) SetUpdatedAt() {
	u.UpdatedAt = time.Now().UnixMilli()
}

func (u *Luna4User) SetLastLoginAt() {
	now := time.Now().UnixMilli()
	u.LastLoginAt = &now
}
