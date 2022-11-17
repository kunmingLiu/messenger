package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v7/linebot"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/kunmingliu/messenger/domain"
	_messageHttpDelivery "github.com/kunmingliu/messenger/message/delivery/http"
	_messageRepo "github.com/kunmingliu/messenger/message/repository/mongo"
	_messageUsecase "github.com/kunmingliu/messenger/message/usecase"
)

type LineProvider struct {
	linebot.Client
}

func (l *LineProvider) ParseRequest(r *http.Request) (msg domain.Message, err error) {
	events, err := l.Client.ParseRequest(r)
	if err != nil {
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			userID := event.Source.UserID
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				msg.UserID = userID
				msg.Message = message.Text
			}
		}
	}
	return
}

func (l *LineProvider) SendMessage(msg string) (err error) {
	_, err = l.Client.BroadcastMessage(linebot.NewTextMessage(msg)).Do()
	return
}

func startServer() {
	if config.Secret == "" {
		panic("secret shouldn't be empty")
	}

	if config.Token == "" {
		panic("token shouldn't be empty")
	}

	if config.DBConfig.User == "" {
		panic("db_user shouldn't be empty")
	}

	if config.DBConfig.Password == "" {
		panic("db_password shouldn't be empty")
	}

	bot, err := linebot.New(config.Secret, config.Token)
	if err != nil {
		panic(err)
	}

	provider := &LineProvider{
		*bot,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("mongodb://%s:%s@%s:%s", config.User, config.Password, config.Host, config.DBConfig.Port)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	db := client.Database("db")

	e := gin.New()
	e.Use(gin.Logger())
	e.Use(gin.Recovery())

	messageRepo := _messageRepo.NewMongoRepository(db)
	timeoutContext := 5 * time.Second
	messageUsecase := _messageUsecase.NewMessageUsecase(messageRepo, provider, timeoutContext)
	_messageHttpDelivery.NewMessageHandler(e, messageUsecase)

	e.Run(":" + config.ServerConfig.Port)
}
