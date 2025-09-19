package main

import (
	"checklist-go/services/api-service/internal/app"
	"log"
)



func main() {
	a, err := app.New()
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}
	a.Run()
}