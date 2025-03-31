package main

import (
	"crypto-project/pkg/application"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"os"
)

func main() {
	//TODO сваггер, грейсфул
	app, err := application.New()
	if err != nil {
		log.Fatalf("failed to init application: %v", err)
	}

	if err = app.Run(); err != nil {
		log.Printf("Application stopped with error: %v", err)
		os.Exit(1)
	}

}
