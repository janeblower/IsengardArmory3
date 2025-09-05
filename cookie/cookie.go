package cookie

import (
	"os"
	"strings"
)

// GetCookie читает cookie из cookie.txt в корне проекта
func GetCookie() string {
	data, err := os.ReadFile("cookie.txt")
	if err != nil {
		return "" // или можно вызвать panic(err), если cookie обязательны
	}
	return strings.TrimSpace(string(data))
}
