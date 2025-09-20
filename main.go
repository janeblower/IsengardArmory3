package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ezserver/addon"
	"ezserver/cookie"
	"ezserver/db"
	"ezserver/http"
	"ezserver/parser"
	"ezserver/utils"
)

func main() {
	cfg := LoadConfig()

	// Подключение к Монге
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort)

	store, err := db.NewMongoStore(uri, cfg.DBName, cfg.DBCollection)
	if err != nil {
		log.Fatalf("Ошибка подключения к MongoDB: %v", err)
	}

	// Сайтик в горутине
	http.SetStore(store)
	go http.RunServer()

	interval := time.Duration(cfg.ParserInterval) * time.Second
	if interval <= 0 {
		interval = 20 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для сигналов ОС
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Парсер в горутине
	go runParserLoop(ctx, store, interval)

	// Ожидание сигнала завершения
	sig := <-sigCh
	log.Printf("Получен сигнал завершения: %v", sig)
	cancel()

	// Даем горутине завершиться
	time.Sleep(1 * time.Second)
	log.Println("Приложение завершено")
}

type Config struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPass         string
	DBName         string
	DBCollection   string
	ParserInterval int
}

func LoadConfig() Config {
	return Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "27017"),
		DBUser:         getEnv("DB_USER", "root"),
		DBPass:         getEnv("DB_PASS", "example"),
		DBName:         getEnv("DB_NAME", "ezwow"),
		DBCollection:   getEnv("DB_COLLECTION", "armory"),
		ParserInterval: utils.ParseInt(getEnv("PARSER_INTERVAL", "20")), // интервал по умолчанию 20 секунд
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func processTicker(st int, store *db.MongoStore) int {
	singleCookie := cookie.GetCookie()
	chars, stMax, ok := parser.ParseCharacters(st, singleCookie)
	if !ok {
		log.Printf("Ошибка парсинга персонажей")
		return st
	}

	if err := store.UpsertCharacters(chars); err != nil {
		log.Printf("Ошибка сохранения персонажей: %v", err)
	}

	if st >= stMax {
		st = 0
		addon.GenerateAddon(store)
	} else {
		st += 20
	}

	log.Println("Страница", st, "из", stMax)
	return st
}

func runParserLoop(ctx context.Context, store *db.MongoStore, interval time.Duration) {
	st := 0
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ctx.Done():
			log.Println("Завершение парсера по контексту")
			break loop
		case <-ticker.C:
			st = processTicker(st, store)
		}
	}
}
