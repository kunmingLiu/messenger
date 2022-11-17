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

//go:generate mockgen -destination=../internal/mocks/domain/provider_mock.go -package=domain github.com/kunmingliu/messenger/domain Provider
type Provider interface {
	ParseRequest(r *http.Request) (Message, error)
	SendMessage(msg string) error
}

//go:generate mockgen -destination=../internal/mocks/domain/repository_mock.go -package=domain github.com/kunmingliu/messenger/domain MessageRepository
type MessageRepository interface {
	Insert(ctx context.Context, m *Message) error
	GetByUserID(ctx context.Context, offset, limit int64, userIds ...string) (messages *[]Message, totalCount int64, err error)
}

//go:generate mockgen -destination=../internal/mocks/domain/usecase_mock.go -package=domain github.com/kunmingliu/messenger/domain MessageUsecase
type MessageUsecase interface {
	Insert(ctx context.Context, m *Message) error
	ParseRequest(r *http.Request) (Message, error)
	Send(msg string) error
	GetByUserID(ctx context.Context, offset, limit int64, userIds ...string) (messages *[]Message, totalCount int64, err error)
}
