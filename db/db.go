package db

import (
	"context"
	"ezserver/parser"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoStore struct {
	client     *mongo.Client
	dbName     string
	collection string
}

func NewMongoStore(uri, dbName, collName string) (*MongoStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	store := &MongoStore{
		client:     client,
		dbName:     dbName,
		collection: collName,
	}

	if err := store.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	log.Println("MongoDB подключен")
	return store, nil
}

func (m *MongoStore) ensureIndexes(ctx context.Context) error {
	coll := m.GetCollection()

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "expireAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0).SetName("expireAt_1"),
		},
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("char_id"),
		},
	}

	_, err := coll.Indexes().CreateMany(ctx, indexes)
	return err
}

func (m *MongoStore) GetCollection() *mongo.Collection {
	return m.client.Database(m.dbName).Collection(m.collection)
}

func (m *MongoStore) UpsertCharacters(chars []parser.Character) error {
	if len(chars) == 0 {
		log.Println("Нет персонажей для апдейта")
		return nil
	}

	coll := m.GetCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var models []mongo.WriteModel
	for _, char := range chars {
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"id": char.ID}).
			SetUpdate(bson.M{"$set": char}).
			SetUpsert(true))
	}

	opts := options.BulkWrite().SetOrdered(false)
	res, err := coll.BulkWrite(ctx, models, opts)
	if err != nil {
		return err
	}

	log.Printf("BulkWrite: Matched %d, Modified %d, Upserts %d\n",
		res.MatchedCount, res.ModifiedCount, res.UpsertedCount)
	return nil
}

func (m *MongoStore) GetCharactersSorted() ([]parser.Character, error) {
	coll := m.GetCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cur, err := coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "login", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var chars []parser.Character
	if err := cur.All(ctx, &chars); err != nil {
		return nil, err
	}
	return chars, nil
}

func (m *MongoStore) CountUniqueLogins() (int64, error) {
	coll := m.GetCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$login"}}}},
		{{Key: "$count", Value: "uniqueLogins"}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err = cursor.All(ctx, &result); err != nil {
		return 0, err
	}

	if len(result) > 0 {
		switch v := result[0]["uniqueLogins"].(type) {
		case int32:
			return int64(v), nil
		case int64:
			return v, nil
		}
	}
	return 0, nil
}

func (m *MongoStore) CountCharacters() (int64, error) {
	coll := m.GetCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return coll.CountDocuments(ctx, bson.M{})
}
