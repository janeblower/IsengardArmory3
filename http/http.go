package http

import (
	"context"
	"ezserver/db"
	"ezserver/parser"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Храним store для всех обработчиков
var mongoStore *db.MongoStore

func SetStore(store *db.MongoStore) {
	mongoStore = store
}

// --- Handlers ---

func getStats(c *gin.Context) {
	count, _ := mongoStore.CountCharacters()
	uniqueLogins, _ := mongoStore.CountUniqueLogins()

	c.JSON(http.StatusOK, gin.H{
		"characters": count,
		"accounts":   uniqueLogins,
		"position":   42,
		"ezwow":      map[string]int{"maxSt": 100},
		"cookies":    10,
	})
}

func getCharacterByName(c *gin.Context) {
	name := c.Param("name")

	coll := mongoStore.GetCollection()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var character parser.Character
	if err := coll.FindOne(ctx, map[string]interface{}{"name": name}).Decode(&character); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Character not found"})
		return
	}

	cursor, err := coll.Find(ctx, map[string]interface{}{"login": character.Login})
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

	c.JSON(http.StatusOK, characters)
}

func getRaces(c *gin.Context) {
	coll := mongoStore.GetCollection()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	pipeline := []map[string]interface{}{
		{"$group": map[string]interface{}{"_id": "$race", "count": map[string]interface{}{"$sum": 1}}},
		{"$sort": map[string]interface{}{"count": -1}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func getClasses(c *gin.Context) {
	coll := mongoStore.GetCollection()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	pipeline := []map[string]interface{}{
		{"$group": map[string]interface{}{"_id": "$class", "count": map[string]interface{}{"$sum": 1}}},
		{"$sort": map[string]interface{}{"count": -1}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// --- Router ---

func RunServer() {
	router := gin.Default()

	router.Static("/static", "./static")
	router.StaticFile("/", "./static/index.html")

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Next()
	})

	router.GET("/api/stats", getStats)
	router.GET("/api/characters/:name", getCharacterByName)
	router.GET("/api/races", getRaces)
	router.GET("/api/classes", getClasses)

	log.Println("Server running at :8080")
	router.Run(":8080")
}
