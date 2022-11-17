package mongo

import (
	"context"
	"time"

	"github.com/kunmingliu/messenger/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoRepository struct {
	DB         *mongo.Database
	Collection *mongo.Collection
}

const (
	collectionName = "message"
)

func NewMongoRepository(DB *mongo.Database) domain.MessageRepository {
	return &mongoRepository{DB, DB.Collection(collectionName)}
}

func (m *mongoRepository) Insert(ctx context.Context, msg *domain.Message) error {
	msg.ID = primitive.NewObjectID().String()
	msg.CreatedAt = time.Now().UTC()
	_, err := m.Collection.InsertOne(ctx, msg)
	return err
}

func (m *mongoRepository) GetByUserID(ctx context.Context, userID ...string) (*[]domain.Message, error) {
	conditions := make([]bson.E, 0)

	for _, id := range userID {
		conditions = append(conditions, bson.E{Key: "user_id", Value: id})
	}

	cursor, err := m.Collection.Find(ctx, conditions)
	if err != nil {
		return nil, err
	}
	var messages []domain.Message
	err = cursor.All(ctx, &messages)
	if err != nil {
		return nil, err
	}
	return &messages, nil
}
