package main

import (
	"log"

	"cinematique/cmd"
)

func main() {
	// Запускаем приложение
	if err := cmd.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
