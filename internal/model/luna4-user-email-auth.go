package model

type Luna4EmailAuth struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Token     string `json:"token"`
	SentAt    int64  `json:"sentAt"`
	Completed bool   `json:"completed"`
}
