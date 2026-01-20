package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"log"
	"time"
)

func InitDatabase() *sql.DB {
	dsn := fmt.Sprintf("host=%s port =%s user=%s password=%s dbname=%s sslmode=disable",
		"localhost",
		"5432",
		"postgres",
		"admin",
		"walletwise",
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to open db:", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("failed to ping db:", err)
	}

	// Connection pool config (INI PENTING)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Hour)

	log.Println("Successfully connected to PostgreSQL")
	return db
}
