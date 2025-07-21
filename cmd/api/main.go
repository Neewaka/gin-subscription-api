package main

import (
	"context"
	"fmt"
	_ "gin-subscription/docs"
	"gin-subscription/internal/database"
	"gin-subscription/internal/env"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
)

type application struct {
	port   int
	models database.Models
}

func main() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		user, password, host, port, dbname,
		// "postgres://postgres_user:postgres_password@postgres:5432/postgres_db"
	)
	db, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
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
