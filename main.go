package main

import (
	"walletwise/database"
)

func main() {
	DB := database.InitDatabase()
	database.NewHandler(DB)

	//userRepo := repositories.NewUserRepository(DB)
}
