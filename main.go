package main

import (
	"ezserver/cookie"
	"ezserver/parser"
	"fmt"
)

func main() {
	singleCookie := cookie.GetCookie()
	st := 20
	stMax, _ := parser.ParseMaxSt(singleCookie)
	chars, charsFound := parser.ParseCharacters(st, singleCookie)

	fmt.Println(chars[0], stMax)

	if !charsFound {
		fmt.Println("Таблица не найдена. Сохраняю весь HTML в debug.html")
	}

}
