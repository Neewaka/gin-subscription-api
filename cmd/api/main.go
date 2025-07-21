package main

import (
	"context"
	"fmt"
	_ "gin-subscription/docs"
	"gin-subscription/internal/database"
	"gin-subscription/internal/env"
	"log"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
)

type application struct {
	port   int
	models database.Models
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		user, password, host, port, dbname,
	)
	db, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Failed to connect db: %v", err)
	}

	defer db.Close(context.Background())

	models := database.NewModels(db)
	app := &application{
		port:   env.GetEnvInt("PORT", 8080),
		models: models,
	}

	if err := app.serve(); err != nil {
		log.Fatal(err)
	}
}
