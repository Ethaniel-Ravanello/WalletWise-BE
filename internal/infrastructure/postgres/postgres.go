package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os" // Tambahkan package os untuk membaca environment variables
	"time"

	_ "github.com/lib/pq"
)

// InitDatabase sekarang mengembalikan error agar bisa ditangani oleh pemanggilnya (main.go)
func InitDatabase() (*sql.DB, error) {
	// Membaca kredensial dari Environment Variables (Lebih Aman!)
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Fallback untuk local development jika env kosong (Opsional tapi berguna)
	if host == "" {
		host = "localhost"
		port = "5432"
		user = "postgres"
		password = "admin"
		dbname = "walletwise"
	}

	// Sintaks spasi pada port diperbaiki
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	// Connection pool config (INI PENTING)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	// 1. Baca isi file migration.sql
	// (Path ini mengasumsikan kamu menjalankan aplikasi dari root folder project)
	sqlBytes, err := os.ReadFile("migration/migration.sql")
	if err != nil {
		return fmt.Errorf("gagal membaca file migrasi: %w", err)
	}

	// 2. Eksekusi isi teks SQL tersebut ke database sekaligus
	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("gagal menjalankan script migrasi: %w", err)
	}

	log.Println("✅ Migrasi database (tabel-tabel) berhasil dibuat/diperbarui!")
	return nil
}
