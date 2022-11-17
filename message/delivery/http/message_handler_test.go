package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kunmingliu/messenger/domain"
	mockDomain "github.com/kunmingliu/messenger/internal/mocks/domain"
)

func TestMessageHandler_GetMessages(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	now := time.Now().UTC()

	userIDs := []string{}
	fakeMessages := []domain.Message{
		{
			ID:        "1",
			UserID:    "user 1",
			Message:   "message 1",
			CreatedAt: now,
		},
		{
			ID:        "2",
			UserID:    "user 2",
			Message:   "message 2",
			CreatedAt: now,
			UpdatedAt: &now,
		},
	}
	mockUsecase := mockDomain.NewMockMessageUsecase(ctl)
	mockUsecase.EXPECT().GetByUserID(gomock.Any(), int64(0), int64(20), userIDs).Return(&fakeMessages, int64(len(fakeMessages)), nil)

	e := gin.New()
	NewMessageHandler(e, mockUsecase)

	req, _ := http.NewRequest("GET", "/messages", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if w.Code != http.StatusOK {
		t.Errorf("code inconsistent, code:%v, expected code:%v", w.Code, http.StatusOK)
	}

	offset := response["offset"].(float64)
	if offset != 0 {
		t.Errorf("offset inconsistent, offset:%v, expected offset:%v", offset, 0)
	}

	limit := response["limit"].(float64)
	if limit != 20 {
		t.Errorf("limit inconsistent, limit:%v, expected limit:%v", limit, 20)
	}

	totalCount := response["total_count"].(float64)
	if totalCount != float64(len(fakeMessages)) {
		t.Errorf("total count inconsistent, count:%v, expected count:%v", totalCount, len(fakeMessages))
	}

	hasNext := response["has_next"].(bool)
	if hasNext {
		t.Errorf("has next inconsistent, value:%v, expected value:%v", hasNext, false)
	}

	var messages []domain.Message
	b, _ := json.Marshal(response["data"])
	json.Unmarshal(b, &messages)
	if len(messages) != len(fakeMessages) {
		t.Errorf("data length inconsistent, length:%v, expected length:%v", len(messages), len(fakeMessages))
	}
	for i, m := range messages {
		fakeMessage := fakeMessages[i]
		if !reflect.DeepEqual(m, fakeMessage) {
			t.Errorf("data inconsistent, type:%v, expected type:%v", m, fakeMessage)
		}
	}
}

func TestMessageHandler_PostMessages(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	fakeMessage := "fake message"
	fakeError := errors.New("fake error")
	validBody, err := json.Marshal(map[string]string{
		"message": fakeMessage,
	})
	if err != nil {
		t.FailNow()
	}
	invalidBody, err := json.Marshal(map[string]string{})

	if err != nil {
		t.FailNow()
	}

	mockUsecase := mockDomain.NewMockMessageUsecase(ctl)
	gomock.InOrder(
		mockUsecase.EXPECT().Send(fakeMessage).Return(nil),
		mockUsecase.EXPECT().Send(fakeMessage).Return(fakeError),
	)

	e := gin.New()
	NewMessageHandler(e, mockUsecase)

	cases := []struct {
		name     string
		arg      []byte
		success  bool
		httpCode int
		err      string
	}{
		{
			name:     "post success",
			arg:      validBody,
			success:  true,
			httpCode: http.StatusCreated,
		},
		{
			name:     "post failed because send message failed",
			arg:      validBody,
			success:  false,
			httpCode: http.StatusInternalServerError,
			err:      fakeError.Error(),
		},
		{
			name:     "post failed because body is invalid",
			arg:      invalidBody,
			success:  false,
			httpCode: http.StatusBadRequest,
			err:      "Key: 'Message' Error:Field validation for 'Message' failed on the 'required' tag",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer(c.arg))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)

			if w.Code != c.httpCode {
				t.Errorf("code inconsistent, code:%v, expected code:%v", w.Code, c.httpCode)
			}
			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			if c.success {
				if response["status"] != "OK" {
					t.Errorf("response inconsistent, response:%v, expected response:%v", response["status"], "OK")
				}
			} else {
				if response["error"] != c.err {
					t.Errorf("response inconsistent, response:%v, expected response:%v", response["error"], c.err)
				}
			}
		})
	}
}

func TestMessageHandler_HandleWebhook(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	fakeParseError := errors.New("parse failed")
	fakeInsertError := errors.New("insert failed")
	fakeMessage := domain.Message{
		UserID:  "user1",
		Message: "message1",
	}
	mockUsecase := mockDomain.NewMockMessageUsecase(ctl)
	gomock.InOrder(
		mockUsecase.EXPECT().ParseRequest(gomock.All()).Return(fakeMessage, nil),
		mockUsecase.EXPECT().Insert(gomock.All(), &fakeMessage).Return(nil),
		mockUsecase.EXPECT().ParseRequest(gomock.All()).Return(fakeMessage, fakeParseError),
		mockUsecase.EXPECT().ParseRequest(gomock.All()).Return(fakeMessage, nil),
		mockUsecase.EXPECT().Insert(gomock.All(), &fakeMessage).Return(fakeInsertError),
	)

	e := gin.New()
	NewMessageHandler(e, mockUsecase)

	cases := []struct {
		name     string
		arg      []byte
		success  bool
		httpCode int
		err      string
	}{
		{
			name:     "post success",
			success:  true,
			httpCode: http.StatusCreated,
		},
		{
			name:     "post failed because parse request failed",
			success:  false,
			httpCode: http.StatusInternalServerError,
			err:      fakeParseError.Error(),
		},
		{
			name:     "post failed because insert failed",
			success:  false,
			httpCode: http.StatusInternalServerError,
			err:      fakeInsertError.Error(),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/webhook", nil)
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)

			if w.Code != c.httpCode {
				t.Errorf("code inconsistent, code:%v, expected code:%v", w.Code, c.httpCode)
			}
			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			if c.success {
				if response["status"] != "OK" {
					t.Errorf("response inconsistent, response:%v, expected response:%v", response["status"], "OK")
				}
			} else {
				if response["error"] != c.err {
					t.Errorf("response inconsistent, response:%v, expected response:%v", response["error"], c.err)
				}
			}

		})
	}
}
