package postgres

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func InitDatabase() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" {
		host = "localhost"
		port = "5432"
		user = "postgres"
		password = "admin"
		dbname = "walletwise"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		zap.L().Error("Failed to connect to database", zap.Error(err))
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		zap.L().Error("Failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	sqlBytes, err := os.ReadFile("migration/migration.sql")
	if err != nil {
		zap.L().Error("Failed to read migration file", zap.Error(err))
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		zap.L().Error("Failed to execute migration script", zap.Error(err))
		return fmt.Errorf("failed to execute migration script: %w", err)
	}

	zap.L().Info("Database migrations executed successfully")
	return nil
}
