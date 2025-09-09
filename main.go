package main

import (
	"ezserver/addon"
	"ezserver/cookie"
	"ezserver/db"
	"ezserver/http"
	"ezserver/parser"
	"log"
	"time"
)

func main() {

	db.InitMongo("ezwow", "armory")

	st := 0

	go http.RunServer()

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for t := range ticker.C {

		singleCookie := cookie.GetCookie()
		chars, stMax, _ := parser.ParseCharacters(st, singleCookie)
		db.UpsertCharacters(chars)

		if st >= stMax {
			st = 0
			addon.GenerateAddon()
		} else {
			st = st + 20
		}

		log.Println("Страница", st, "из", stMax, t)
	}

}
