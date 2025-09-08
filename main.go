package main

import (
	"ezserver/cookie"
	"ezserver/db"
	"ezserver/parser"
	"fmt"
	"time"
)

func main() {
	// singleCookie := cookie.GetCookie()
	st := 0
	stMax, _ := parser.ParseMaxSt(cookie.GetCookie())
	// chars, charsFound := parser.ParseCharacters(st, singleCookie)

	// fmt.Println(chars[0])
	db.InitMongo("ezwow", "armory")
	// db.UpsertCharacters(chars)

	// if !charsFound {
	// 	fmt.Println("Таблица не найдена. Сохраняю весь HTML в debug.html")
	// }

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop() // корректно остановим при завершении

	for t := range ticker.C {

		if st == stMax {
			st = 0
		}

		singleCookie := cookie.GetCookie()
		chars, _ := parser.ParseCharacters(st, singleCookie)
		db.UpsertCharacters(chars)

		fmt.Println("Страница ", st, t)
		st = st + 20
	}
}
