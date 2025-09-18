package utils

import (
	"log"
	"strconv"
)

func ParseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
