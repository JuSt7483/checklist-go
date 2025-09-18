package main

import (
	"log"

	"checklist-go/services/db-service/internal/app"
)

func main(){
	a, err := app.New()
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}
	a.Run()
}