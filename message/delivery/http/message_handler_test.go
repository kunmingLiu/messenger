package http

import (
	"encoding/json"
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
