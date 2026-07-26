package main

import (
	"log"
	"net/http"

	budgetService "walletwise/internal/application/budget"
	categoriesService "walletwise/internal/application/categories"
	savingGoalService "walletwise/internal/application/saving_goal"
	trxService "walletwise/internal/application/transaction"
	userService "walletwise/internal/application/user"
	walletService "walletwise/internal/application/wallet"
	"walletwise/internal/infrastructure/postgres"
	"walletwise/internal/infrastructure/transport"
)

func main() {

	db, err := postgres.InitDatabase()
	if err != nil {

		log.Fatalf("Gagal menyalakan database: %v", err)
	}
	defer db.Close()

	if err := postgres.RunMigrations(db); err != nil {
		log.Fatalf("Database Migration Error: %v", err)
	}

	trxRepo := postgres.NewTransactionRepo(db)
	trxsService := trxService.NewService(trxRepo)
	trxHandler := transport.NewTransactionHandler(trxsService)

	userRepo := postgres.NewUserRepo(db)
	usersService := userService.NewService(userRepo)
	userHandler := transport.NewUserHandler(usersService)

	categoriesRepo := postgres.NewCategoriesRepo(db)
	catService := categoriesService.NewService(categoriesRepo)
	categoriesHandler := transport.NewCategoriesHandler(catService)

	walletRepo := postgres.NewWalletRepo(db)
	wService := walletService.NewWalletService(walletRepo)
	walletHandler := transport.NewWalletHandler(wService)

	savingGoalsRepo := postgres.NewSavingGoalsRepo(db)
	savService := savingGoalService.NewService(savingGoalsRepo)
	savingGoalsHandler := transport.NewSavingGoalsHandler(savService)

	budgetRepo := postgres.NewBudgetRepo(db)
	budgetService := budgetService.NewService(budgetRepo)
	budgetHandler := transport.NewBudgetHandler(budgetService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /transactions", trxHandler.CreateTransaction)
	mux.HandleFunc("GET /transactions", trxHandler.GetTransactions)
	mux.HandleFunc("GET /transactions/{id}", trxHandler.GetTransactionById)
	mux.HandleFunc("PUT /transactions/{id}", trxHandler.UpdateTransaction)
	mux.HandleFunc("DELETE /transactions/{id}", trxHandler.DeleteTransaction)

	mux.HandleFunc("POST /users", userHandler.CreateUser)
	mux.HandleFunc("GET /users/{id}", userHandler.GetUserByID)
	mux.HandleFunc("GET /users/email/{email}", userHandler.GetUserByEmail)
	mux.HandleFunc("PUT /users/{id}", userHandler.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", userHandler.DeleteUser)

	mux.HandleFunc("GET /categories", categoriesHandler.GetAllCategories)

	mux.HandleFunc("POST /wallets", walletHandler.CreateWallets)
	mux.HandleFunc("GET /wallets", walletHandler.SearchAllWallets)
	mux.HandleFunc("GET /wallets/{id}", walletHandler.SearchWalletsByID)
	mux.HandleFunc("PUT /wallets/{id}", walletHandler.UpdateWallet)
	mux.HandleFunc("DELETE /wallets/{id}", walletHandler.DeleteWallet)

	mux.HandleFunc("GET /wallets/users/{userId}/highest-balance", walletHandler.SearchHighestBalance)
	mux.HandleFunc("GET /wallets/users/{userId}/most-active", walletHandler.SearchMostActive)
	mux.HandleFunc("GET /wallets/users/{userId}/total-balance", walletHandler.SearchTotalBalance)

	mux.HandleFunc("POST /saving-goals", savingGoalsHandler.CreateGoal)
	mux.HandleFunc("GET /saving-goals", savingGoalsHandler.GetAllGoals)
	mux.HandleFunc("GET /saving-goals/{id}", savingGoalsHandler.GetGoalByID)
	mux.HandleFunc("PUT /saving-goals/{id}", savingGoalsHandler.UpdateGoal)
	mux.HandleFunc("DELETE /saving-goals/{id}", savingGoalsHandler.DeleteGoal)

	mux.HandleFunc("POST /budgets", budgetHandler.CreateBudget)

	mux.HandleFunc("GET /budgets", budgetHandler.GetBudgetsByMonth)

	mux.HandleFunc("GET /budgets/{id}", budgetHandler.GetBudgetByID)
	mux.HandleFunc("PUT /budgets/{id}", budgetHandler.UpdateBudget)
	mux.HandleFunc("DELETE /budgets/{id}", budgetHandler.DeleteBudget)

	port := ":8080"
	log.Println("🚀 Server WalletWise menyala dan mendengarkan di port", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server mati secara tidak wajar: %v", err)
	}
}
