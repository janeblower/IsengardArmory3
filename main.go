package main

import (
	"ezserver/addon"
	"ezserver/db"
)

func main() {

	db.InitMongo("ezwow", "armory")

	// st := 33080

	// go http.RunServer()

	// ticker := time.NewTicker(20 * time.Second)
	// defer ticker.Stop()

	// for t := range ticker.C {

	// 	singleCookie := cookie.GetCookie()
	// 	chars, stMax, _ := parser.ParseCharacters(st, singleCookie)
	// 	db.UpsertCharacters(chars)

	// 	if st >= stMax {
	// 		st = 0
	// 		// addon.GenerateAddon()
	// 	} else {
	// 		st = st + 20
	// 	}

	// 	fmt.Println("Страница", st, "из", stMax, t)
	// }

	addon.GenerateAddon()

}
