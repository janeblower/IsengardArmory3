package db

import (
	"context"
	"ezserver/parser"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var mongoURI = "mongodb://root:example@localhost:27017/"

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
		fmt.Println("Коллекция создана:", collName)
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
		fmt.Println("TTL индекс создан")
	}

	if !existsUniqueName {
		uniqueIndex := mongo.IndexModel{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("unique_name"),
		}
		if _, err := coll.Indexes().CreateOne(ctx, uniqueIndex); err != nil {
			log.Fatalf("Ошибка создания уникального индекса по name: %v", err)
		}
		fmt.Println("Уникальный индекс по name создан")
	}
}

// GetCollection просто возвращает подключение к коллекции
func GetCollection(dbName, collName string) (*mongo.Collection, context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}

	return client.Database(dbName).Collection(collName), ctx, cancel
}

// UpsertCharacters теперь использует GetCollection
func UpsertCharacters(chars []parser.Character) {
	coll, ctx, cancel := GetCollection("ezwow", "armory")
	defer cancel()

	for _, char := range chars {
		filter := bson.M{"name": char.Name}
		update := bson.M{"$set": char}
		opts := options.UpdateOne().SetUpsert(true)

		res, err := coll.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Matched: %d, Modified: %d\n", res.MatchedCount, res.ModifiedCount)
	}
}
