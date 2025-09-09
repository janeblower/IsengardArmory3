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

var mongoURI = "mongodb://root:example@localhost:27017/"

var mongoClient *mongo.Client

// InitMongo создаёт базу, коллекцию и индексы — вызывается один раз в main
func InitMongo(dbName, collName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Ошибка ping: %v", err)
	}

	db := client.Database(dbName)

	// Создаём коллекцию, если нет
	collections, err := db.ListCollectionNames(ctx, bson.M{"name": collName})
	if err != nil {
		log.Fatalf("Ошибка списка коллекций: %v", err)
	}
	if len(collections) == 0 {
		if err := db.CreateCollection(ctx, collName); err != nil {
			log.Fatalf("Ошибка создания коллекции: %v", err)
		}
		log.Println("Коллекция создана:", collName)
	}

	coll := db.Collection(collName)

	// Проверяем индексы
	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		log.Fatalf("Ошибка получения индексов: %v", err)
	}
	defer cursor.Close(ctx)

	existsTTL := false
	existsUniqueName := false

	for cursor.Next(ctx) {
		var idx bson.M
		if err := cursor.Decode(&idx); err != nil {
			log.Fatalf("Ошибка декодирования индекса: %v", err)
		}
		if name, ok := idx["name"].(string); ok {
			if name == "expireAt_1" {
				existsTTL = true
			}
			if name == "unique_name" {
				existsUniqueName = true
			}
		}
	}

	if !existsTTL {
		ttlIndex := mongo.IndexModel{
			Keys:    bson.D{{Key: "expireAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0).SetName("expireAt_1"),
		}
		if _, err := coll.Indexes().CreateOne(ctx, ttlIndex); err != nil {
			log.Fatalf("Ошибка создания TTL индекса: %v", err)
		}
		log.Println("TTL индекс создан")
	}

	if !existsUniqueName {
		uniqueIndex := mongo.IndexModel{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("char_id"),
		}
		if _, err := coll.Indexes().CreateOne(ctx, uniqueIndex); err != nil {
			log.Fatalf("Ошибка создания уникального индекса по id: %v", err)
		}
		log.Println("Уникальный индекс по id создан")
	}

	mongoClient = client
	log.Println("MongoDB подключен")
}

// GetCollection просто возвращает подключение к коллекции
func GetCollection(dbName, collName string) *mongo.Collection {
	return mongoClient.Database(dbName).Collection(collName)
}

// UpsertCharacters теперь использует GetCollection
func UpsertCharacters(chars []parser.Character) {
	coll := GetCollection("ezwow", "armory")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var models []mongo.WriteModel
	for _, char := range chars {
		filter := bson.M{"id": char.ID}
		update := bson.M{"$set": char}
		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
		models = append(models, model)
	}

	if len(models) == 0 {
		log.Println("Нет персонажей для апдейта")
		return
	}

	res, err := coll.BulkWrite(ctx, models)
	if err != nil {
		log.Println("BulkWrite error:", err)
		return
	}

	log.Printf("BulkWrite: Matched %d, Modified %d, Upserts %d\n",
		res.MatchedCount, res.ModifiedCount, res.UpsertedCount)
}

// GetCharactersSorted возвращает всех персонажей, отсортированных по login
func GetCharactersSorted() []parser.Character {
	coll := GetCollection("ezwow", "armory")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	findOptions := options.Find().SetSort(bson.D{{Key: "login", Value: 1}})
	cur, err := coll.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	var chars []parser.Character
	for cur.Next(ctx) {
		var c parser.Character
		if err := cur.Decode(&c); err != nil {
			log.Println("decode error:", err)
			continue
		}
		chars = append(chars, c)
	}
	return chars
}

// CountUniqueLogins возвращает количество уникальных логинов
func CountUniqueLogins() int64 {
	coll := GetCollection("ezwow", "armory")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$login"}}}},
		{{Key: "$count", Value: "uniqueLogins"}},
	}
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err = cursor.All(ctx, &result); err != nil {
		log.Fatal(err)
	}
	if len(result) > 0 {
		if val, ok := result[0]["uniqueLogins"].(int32); ok {
			return int64(val)
		}
		if val, ok := result[0]["uniqueLogins"].(int64); ok {
			return val
		}
	}
	return 0
}

// CountCharacters возвращает количество всех персонажей
func CountCharacters() int64 {
	coll := GetCollection("ezwow", "armory")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	return count
}
