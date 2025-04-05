package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var DB *pgxpool.Pool

// @title Manajemen Produk & Transaksi API
// @version 1.0
// @description API ini terhubung ke PostgreSQL melalui koneksi pool menggunakan `pgx`.
// @host localhost:8080
// @BasePath /
// @schemes http
func ConnectDB() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
    
    dsn := os.Getenv("DATABASE_URL")

    DB, err = pgxpool.New(context.Background(), dsn)
    if err != nil {
        log.Fatalf("Unable to connect to database: %v\n", err)
    }

    log.Println("Connected to PostgreSQL!")
}



