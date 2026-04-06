package main

import (
	"walletwise/database"
	"walletwise/internal/infrastructure/postgres"
)

func main() {
	DB := postgres.InitDatabase()
	database.NewHandler(DB)

	//userRepo := repositories.NewUserRepository(DB)
}
