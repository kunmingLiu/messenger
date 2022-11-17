package mongo

import (
	"context"
	"time"

	"github.com/kunmingliu/messenger/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (m *mongoRepository) GetByUserID(ctx context.Context, offset, limit int64, userID ...string) (*[]domain.Message, int64, error) {
	filter := make([]bson.E, 0)

	for _, id := range userID {
		filter = append(filter, bson.E{Key: "user_id", Value: id})
	}

	findOptions := options.Find()
	findOptions.SetSkip(offset)
	findOptions.SetLimit(limit)

	totalCount, err := m.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	cursor, err := m.Collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	var messages []domain.Message
	err = cursor.All(ctx, &messages)
	if err != nil {
		return nil, 0, err
	}
	return &messages, totalCount, nil
}
