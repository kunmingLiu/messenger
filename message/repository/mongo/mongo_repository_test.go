package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/kunmingliu/messenger/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func compare(s, t domain.Message) bool {
	if s.ID != t.ID {
		return false
	}
	if s.UserID != t.UserID {
		return false
	}
	if s.Message != t.Message {
		return false
	}
	// millisecond would be truncated when inserting into database
	pattern := "2006-01-02 15:04:05"
	if s.CreatedAt.Format(pattern) != t.CreatedAt.Format(pattern) {
		return false
	}
	if s.UpdatedAt != nil && t.UpdatedAt != nil {
		return s.UpdatedAt.Format(pattern) == t.UpdatedAt.Format(pattern)
	}
	return s.UpdatedAt == t.UpdatedAt

}

func Test_mongoRepository_Insert(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("insert new message", func(mt *mtest.T) {
		messageCollection := mt.Coll
		ctx := context.Background()
		id := primitive.NewObjectID().String()
		now := time.Now().UTC()

		messageData := &domain.Message{
			Message: "test message",
			UserID:  "U12345",
		}
		rawData := bson.D{
			{Key: "_id", Value: id},
			{Key: "user_id", Value: messageData.UserID},
			{Key: "message", Value: messageData.Message},
			{Key: "created_at", Value: now},
			{Key: "updated_at", Value: nil},
		}
		mt.AddMockResponses(mtest.CreateSuccessResponse(), mtest.CreateCursorResponse(1, "db.message", mtest.FirstBatch, rawData))

		m := &mongoRepository{
			DB:         mt.DB,
			Collection: messageCollection,
		}
		err := m.Insert(ctx, messageData)
		if err != nil {
			t.Errorf("insert failed, err: %v", err)
		}
		res := mt.Coll.FindOne(ctx, bson.D{})
		var messageFromDB domain.Message
		res.Decode(&messageFromDB)

		messageData.ID = id
		messageData.CreatedAt = now
		if !compare(*messageData, messageFromDB) {
			t.Errorf("data inconsistent, origin:%+v, new:%+v", messageData, messageFromDB)
		}
	})
}

func Test_mongoRepository_GetByUserID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("GetByUserID success", func(mt *mtest.T) {
		messageCollection := mt.Coll
		ctx := context.Background()

		yesterday := time.Now().UTC().Add(-24 * time.Hour)
		now := time.Now().UTC()

		messageData1 := &domain.Message{
			ID:        primitive.NewObjectID().String(),
			Message:   "test message1",
			UserID:    "user1",
			CreatedAt: now,
		}
		messageData2 := &domain.Message{
			ID:        primitive.NewObjectID().String(),
			Message:   "test message2",
			UserID:    "user2",
			CreatedAt: yesterday,
			UpdatedAt: &now,
		}

		rawData1 := bson.D{
			{Key: "_id", Value: messageData1.ID},
			{Key: "user_id", Value: messageData1.UserID},
			{Key: "message", Value: messageData1.Message},
			{Key: "created_at", Value: messageData1.CreatedAt},
			{Key: "updated_at", Value: messageData1.UpdatedAt},
		}
		rawData2 := bson.D{
			{Key: "_id", Value: messageData2.ID},
			{Key: "user_id", Value: messageData2.UserID},
			{Key: "message", Value: messageData2.Message},
			{Key: "created_at", Value: messageData2.CreatedAt},
			{Key: "updated_at", Value: messageData2.UpdatedAt},
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, err := messageCollection.InsertOne(ctx, rawData1)
		if err != nil {
			t.Errorf("insert1 failed, err: %v", err)
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, err = messageCollection.InsertOne(ctx, rawData2)
		if err != nil {
			t.Errorf("insert2 failed, err: %v", err)
		}

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "db.message", mtest.FirstBatch, bson.D{
				{Key: "n", Value: 2},
			}),

			mtest.CreateCursorResponse(1, "db.message", mtest.FirstBatch, rawData1),
			mtest.CreateCursorResponse(1, "db.message", mtest.NextBatch, rawData2),
			mtest.CreateCursorResponse(0, "db.message", mtest.NextBatch),
		)

		m := &mongoRepository{
			DB:         mt.DB,
			Collection: messageCollection,
		}

		messages, totalCount, err := m.GetByUserID(ctx, 0, 20)
		if err != nil {
			t.Errorf("get messages failed, err: %v", err)
		}

		if messages == nil || len(*messages) != 2 {
			t.Errorf("data inconsistent, the length of response should be %d", 2)
		}

		if int(totalCount) != len(*messages) {
			t.Errorf("count inconsistent, total count:%v, expected count:%v", totalCount, len(*messages))
		}

		msgs := *messages
		if !compare(msgs[0], *messageData1) {
			t.Errorf("data inconsistent, origin:%+v, new:%+v", msgs[0], *messageData1)
		}
		if !compare(msgs[1], *messageData2) {
			t.Errorf("data inconsistent, origin:%+v, new:%+v", msgs[1], *messageData2)
		}
	})
}
