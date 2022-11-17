package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kunmingliu/messenger/domain"
)

type ResponseError struct {
	Error string `json:"error"`
}

type MessageHandler struct {
	MessageUsecase domain.MessageUsecase
}

func NewMessageHandler(e *gin.Engine, ms domain.MessageUsecase) {
	handler := &MessageHandler{
		MessageUsecase: ms,
	}
	e.GET("/messages", handler.GetMessages)
}

func (m *MessageHandler) GetMessages(c *gin.Context) {
	ids := c.QueryArray("user_id")

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
		return
	}

	ctx := c.Request.Context()
	messages, totalCount, err := m.MessageUsecase.GetByUserID(ctx, int64(offset), int64(limit), ids...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseError{Error: err.Error()})
		return
	}

	hasNext := true
	nextOffset := limit + offset
	if (int64)(nextOffset) >= totalCount {
		hasNext = false
	}

	resp := gin.H{
		"total_count": totalCount,
		"offset":      offset,
		"limit":       limit,
		"has_next":    hasNext,
	}

	//return empty array instead
	if messages == nil || len(*messages) == 0 {
		resp["data"] = []interface{}{}
	} else {
		resp["data"] = *messages
	}
	c.JSON(http.StatusOK, resp)
}
