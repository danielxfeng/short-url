package main

import (
	"log"

	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/db/sqlc"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/dep"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := dep.LoadConfigFromEnv()
	if err != nil {
		log.Fatal("failed to init cfg", "err: ", err)
	}

	if err := db.MigrateDB(cfg.DbURL); err != nil {
		log.Fatal("failed to migrate db", "err: ", err)
	}

	log.Println("migration successful")
}
