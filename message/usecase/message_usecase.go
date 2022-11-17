package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/kunmingliu/messenger/domain"
)

type messageUsecase struct {
	messageRepo     domain.MessageRepository
	contextTimeout  time.Duration
	messageProvider domain.Provider
}

func NewMessageUsecase(m domain.MessageRepository, p domain.Provider, timeout time.Duration) domain.MessageUsecase {
	return &messageUsecase{
		messageRepo:     m,
		contextTimeout:  timeout,
		messageProvider: p,
	}
}

func (m *messageUsecase) Insert(c context.Context, msg *domain.Message) (err error) {
	ctx, cancel := context.WithTimeout(c, m.contextTimeout)
	defer cancel()
	err = m.messageRepo.Insert(ctx, msg)
	return
}

func (m *messageUsecase) ParseRequest(r *http.Request) (msg domain.Message, err error) {
	msg, err = m.messageProvider.ParseRequest(r)
	return
}

func (m *messageUsecase) Send(msg string) (err error) {
	err = m.messageProvider.SendMessage(msg)
	return
}

func (m *messageUsecase) GetByUserID(c context.Context, offset, limit int64, userID ...string) (messages *[]domain.Message, totalCount int64, err error) {
	ctx, cancel := context.WithTimeout(c, m.contextTimeout)
	defer cancel()
	messages, totalCount, err = m.messageRepo.GetByUserID(ctx, offset, limit, userID...)
	return
}
