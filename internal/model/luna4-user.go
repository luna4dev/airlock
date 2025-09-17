package model

import "time"

type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
)

type Luna4User struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Status    UserStatus `json:"status"`
	CreatedAt int64      `json:"createdAt"`
	UpdatedAt int64      `json:"updatedAt"`
}

func (u *Luna4User) SetUpdatedAt() {
	u.UpdatedAt = time.Now().UnixMilli()
}
