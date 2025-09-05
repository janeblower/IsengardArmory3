package main

import (
	"ezserver/cookie"
	"ezserver/parser"
	"fmt"
)

func main() {
	singleCookie := cookie.GetCookie()
	st := 20
	chars, found := parser.ParseCharacters(st, singleCookie)

	fmt.Println(chars)

	if !found {
		fmt.Println("Таблица не найдена. Сохраняю весь HTML в debug.html")
	}

}
