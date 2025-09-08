package http

import (
	"context"
	"ezserver/db"
	"ezserver/parser"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// --- Handlers ---

func getStats(c *gin.Context) {
	coll := db.GetCollection("ezwow", "armory")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"characters": count,
		"position":   42,
		"ezwow":      map[string]int{"maxSt": 100},
		"cookies":    10,
	})
}

func getCharacterByName(c *gin.Context) {
	name := c.Param("name")

	coll := db.GetCollection("ezwow", "armory")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1) Найти персонажа по имени
	var character parser.Character
	err := coll.FindOne(ctx, bson.M{"name": name}).Decode(&character)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Character not found"})
		return
	}

	// 2) Получить login найденного персонажа
	login := character.Login

	// 3) Найти всех персонажей с этим login
	cursor, err := coll.Find(ctx, bson.M{"login": login})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(ctx)

	var characters []parser.Character
	if err := cursor.All(ctx, &characters); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding results"})
		return
	}

	// 4) Вернуть список найденных персонажей
	c.JSON(http.StatusOK, characters)
}

func getRaces(c *gin.Context) {
	coll := db.GetCollection("ezwow", "armory")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$race"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "count", Value: -1},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	type Stat struct {
		ID    string `bson:"_id"`
		Count int    `bson:"count"`
	}

	var results []Stat
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func getClasses(c *gin.Context) {
	coll := db.GetCollection("ezwow", "armory")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$class"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "count", Value: -1},
		}}},
	}
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	type Stat struct {
		ID    string `bson:"_id"`
		Count int    `bson:"count"`
	}

	var results []Stat
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// --- Router ---

func RunServer() {

	router := gin.Default()

	// Раздача статики
	router.Static("/static", "./static")
	router.StaticFile("/", "./static/index.html")

	// CORS если нужен
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Next()
	})

	// API endpoints
	router.GET("/api/stats", getStats)
	router.GET("/api/characters/:name", getCharacterByName)
	router.GET("/api/races", getRaces)
	router.GET("/api/classes", getClasses)

	log.Println("Server running at :8080")
	router.Run(":8080")
}
