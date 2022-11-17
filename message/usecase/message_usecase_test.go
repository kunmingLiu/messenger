package usecase

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/kunmingliu/messenger/domain"
	mockDomain "github.com/kunmingliu/messenger/internal/mocks/domain"
)

func Test_messageUsecase_Insert(t *testing.T) {

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	timeout := time.Second * 5
	backgroundCtx := context.Background()

	mockRepository := mockDomain.NewMockMessageRepository(ctl)
	mockProvider := mockDomain.NewMockProvider(ctl)

	m := &domain.Message{}
	fakeError := errors.New("fake error")
	gomock.InOrder(
		//context arguments would be treated as different even if they are the same type.
		mockRepository.EXPECT().Insert(gomock.Any(), m).Return(nil),
		mockRepository.EXPECT().Insert(gomock.Any(), m).Return(fakeError),
	)

	usecase := NewMessageUsecase(mockRepository, mockProvider, timeout)

	err := usecase.Insert(backgroundCtx, m)
	if err != nil {
		t.Errorf("unexpected error:%v", err)
	}
	err = usecase.Insert(backgroundCtx, m)
	if err.Error() != fakeError.Error() {
		t.Errorf("error inconsistent, caught error:%v, expected error:%v", err, fakeError)
	}
}

func Test_messageUsecase_Parse(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	timeout := time.Second * 5

	mockRepository := mockDomain.NewMockMessageRepository(ctl)
	mockProvider := mockDomain.NewMockProvider(ctl)

	req, _ := http.NewRequest("post", "google.com", nil)
	fakeMsg := domain.Message{
		UserID:  "123",
		Message: "test message",
	}
	fakeError := errors.New("fake error")
	gomock.InOrder(
		mockProvider.EXPECT().ParseRequest(req).Return(fakeMsg, nil),
		mockProvider.EXPECT().ParseRequest(req).Return(domain.Message{}, fakeError),
	)

	usecase := NewMessageUsecase(mockRepository, mockProvider, timeout)

	msg, err := usecase.ParseRequest(req)
	if err != nil {
		t.Errorf("unexpected error:%v", err)
	}
	if msg.UserID != fakeMsg.UserID || msg.Message != fakeMsg.Message {
		t.Errorf("data inconsistent, msg:%v, expected message:%v", msg, fakeMsg)
	}

	msg, err = usecase.ParseRequest(req)
	if err.Error() != fakeError.Error() {
		t.Errorf("error inconsistent, caught error:%v, expected error:%v", err, fakeError)
	}
}

func Test_messageUsecase_GetByUserID(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	timeout := time.Second * 5
	backgroundCtx := context.Background()

	mockRepository := mockDomain.NewMockMessageRepository(ctl)
	mockProvider := mockDomain.NewMockProvider(ctl)

	userIDs := []string{
		"123",
		"456",
	}
	fakeMessages := []domain.Message{
		{
			Message: "test message1",
			UserID:  "user 1",
		},
		{
			Message: "test message2",
			UserID:  "user 2",
		},
	}
	fakeError := errors.New("fake error")
	gomock.InOrder(
		//context arguments would be treated as different even if they are the same type.
		mockRepository.EXPECT().GetByUserID(gomock.Any(), int64(0), int64(20), userIDs).Return(&fakeMessages, int64(len(fakeMessages)), nil),
		mockRepository.EXPECT().GetByUserID(gomock.Any(), int64(0), int64(20), userIDs).Return(nil, int64(0), fakeError),
	)

	usecase := NewMessageUsecase(mockRepository, mockProvider, timeout)

	messages, totalCount, err := usecase.GetByUserID(backgroundCtx, 0, 20, userIDs...)
	if err != nil {
		t.Errorf("unexpected error:%v", err)
	}
	if messages == nil || len(*messages) != len(fakeMessages) {
		t.Errorf("data inconsistent, messages:%v, expected message:%v", messages, fakeMessages)
	}
	if int(totalCount) != len(fakeMessages) {
		t.Errorf("count inconsistent, total count:%v, expected count:%v", totalCount, len(fakeMessages))
	}

	for i, message := range *messages {
		target := fakeMessages[i]

		if message.UserID != target.UserID || message.Message != target.Message {
			t.Errorf("data inconsistent, messages:%v, expected message:%v", *messages, fakeMessages)
		}
	}

	_, _, err = usecase.GetByUserID(backgroundCtx, 0, 20, userIDs...)
	if err.Error() != fakeError.Error() {
		t.Errorf("error inconsistent, caught error:%v, expected error:%v", err, fakeError)
	}
}
