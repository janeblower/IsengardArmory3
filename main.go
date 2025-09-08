package main

import (
	"ezserver/cookie"
	"ezserver/db"
	"ezserver/http"
	"ezserver/parser"
	"fmt"
	"time"
)

func main() {

	// Запуск базы и парсинга
	st := 0
	// stMax, _ := parser.ParseMaxSt(cookie.GetCookie())

	db.InitMongo("ezwow", "armory")

	go http.RunServer()

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for t := range ticker.C {

		singleCookie := cookie.GetCookie()
		chars, stMax, _ := parser.ParseCharacters(st, singleCookie)
		db.UpsertCharacters(chars)

		if st == stMax {
			st = 0
		} else {
			st = st + 20
		}

		fmt.Println("Страница ", st, t)
	}

}
