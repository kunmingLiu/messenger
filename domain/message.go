package domain

import (
	"context"
	"net/http"
	"time"
)

type Message struct {
	ID        string     `bson:"_id" json:"id"`
	CreatedAt time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	UserID    string     `bson:"user_id" json:"user_id" validate:"required"`
	Message   string     `bson:"message" json:"message" validate:"required"`
}

type Provider interface {
	ParseRequest(r *http.Request) (Message, error)
	SendMessage(msg string) error
}

type MessageRepository interface {
	Insert(ctx context.Context, m *Message) error
	GetByUserID(ctx context.Context, userIds ...string) (messages *[]Message, err error)
}

type MessageUsecase interface {
	Insert(ctx context.Context, m *Message) error
	ParseRequest(r *http.Request) (Message, error)
	Send(msg string) error
	GetByUserID(ctx context.Context, userIds ...string) (messages *[]Message, err error)
}
