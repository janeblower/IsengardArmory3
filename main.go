package main

import (
	"context"
	"ezserver/addon"
	"ezserver/cookie"
	"ezserver/db"
	"ezserver/http"
	"ezserver/parser"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func processTicker(st int) int {
	singleCookie := cookie.GetCookie()
	chars, stMax, ok := parser.ParseCharacters(st, singleCookie)
	if !ok {
		log.Printf("Ошибка парсинга персонажей")
		return st
	}
	db.UpsertCharacters(chars)

	if st >= stMax {
		st = 0
		addon.GenerateAddon()
	} else {
		st = st + 20
	}

	log.Println("Страница", st, "из", stMax)
	return st
}

func main() {
	db.InitMongo("ezwow", "armory")

	st := 0

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для сигналов ОС
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Запуск HTTP сервера в отдельной горутине
	go http.RunServer()

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ctx.Done():
			log.Println("Завершение по контексту")
			break loop
		case sig := <-sigCh:
			log.Printf("Получен сигнал завершения: %v", sig)
			cancel()
			break loop
		case <-ticker.C:
			st = processTicker(st)
		}
	}

	log.Println("Приложение завершено")
}
