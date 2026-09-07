package main

import (
	"net/http"
	"walletwise/internal/middleware"
	"walletwise/pkg/logger"

	budgetService "walletwise/internal/application/budget"
	categoryService "walletwise/internal/application/category"
	savingGoalService "walletwise/internal/application/saving_goal"
	trxService "walletwise/internal/application/transaction"
	userService "walletwise/internal/application/user"
	walletService "walletwise/internal/application/wallet"
	"walletwise/internal/infrastructure/postgres"
	"walletwise/internal/infrastructure/transport"

	"go.uber.org/zap"
)

func main() {

	logger.InitLogger()
	defer zap.L().Sync()

	db, err := postgres.InitDatabase()
	if err != nil {
		zap.L().Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	if err := postgres.RunMigrations(db); err != nil {
		zap.L().Fatal("failed to run migration", zap.Error(err))
	}

	trxRepo := postgres.NewTransactionRepo(db)
	trxsService := trxService.NewService(trxRepo)
	trxHandler := transport.NewTransactionHandler(trxsService)

	userRepo := postgres.NewUserRepo(db)
	usersService := userService.NewService(userRepo)
	userHandler := transport.NewUserHandler(usersService)

	categoriesRepo := postgres.NewCategoriesRepo(db)
	catService := categoryService.NewService(categoriesRepo)
	categoriesHandler := transport.NewCategoryHandler(catService)

	walletRepo := postgres.NewWalletRepo(db)
	wService := walletService.NewService(walletRepo)
	walletHandler := transport.NewWalletHandler(wService)

	savingGoalsRepo := postgres.NewSavingGoalsRepo(db)
	savService := savingGoalService.NewService(savingGoalsRepo)
	savingGoalsHandler := transport.NewSavingGoalHandler(savService)

	budgetRepo := postgres.NewBudgetRepo(db)
	budgetService := budgetService.NewService(budgetRepo)
	budgetHandler := transport.NewBudgetHandler(budgetService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", userHandler.CreateUser)
	mux.HandleFunc("POST /login", userHandler.Login)

	mux.Handle("POST /transactions", middleware.AuthMiddleware(http.HandlerFunc(trxHandler.CreateTransaction)))
	mux.Handle("GET /transactions", middleware.AuthMiddleware(http.HandlerFunc(trxHandler.GetTransactions)))
	mux.Handle("GET /transactions/{id}", middleware.AuthMiddleware(http.HandlerFunc(trxHandler.GetTransactionByID)))
	mux.Handle("PUT /transactions/{id}", middleware.AuthMiddleware(http.HandlerFunc(trxHandler.UpdateTransaction)))
	mux.Handle("DELETE /transactions/{id}", middleware.AuthMiddleware(http.HandlerFunc(trxHandler.DeleteTransaction)))

	// --- Users ---
	mux.Handle("GET /users/{id}", middleware.AuthMiddleware(http.HandlerFunc(userHandler.GetUserByID)))
	mux.Handle("GET /users/email/{email}", middleware.AuthMiddleware(http.HandlerFunc(userHandler.GetUserByEmail)))
	mux.Handle("PUT /users/{id}", middleware.AuthMiddleware(http.HandlerFunc(userHandler.UpdateUser)))
	mux.Handle("DELETE /users/{id}", middleware.AuthMiddleware(http.HandlerFunc(userHandler.DeleteUser)))

	// --- Categories ---
	mux.Handle("GET /categories", middleware.AuthMiddleware(http.HandlerFunc(categoriesHandler.GetAllCategories)))

	// --- Wallets ---
	mux.Handle("POST /wallets", middleware.AuthMiddleware(http.HandlerFunc(walletHandler.CreateWallet)))
	mux.Handle("GET /wallets", middleware.AuthMiddleware(http.HandlerFunc(walletHandler.GetWallets)))
	mux.Handle("GET /wallets/{id}", middleware.AuthMiddleware(http.HandlerFunc(walletHandler.GetWalletByID)))
	mux.Handle("PUT /wallets/{id}", middleware.AuthMiddleware(http.HandlerFunc(walletHandler.UpdateWallet)))
	mux.Handle("DELETE /wallets/{id}", middleware.AuthMiddleware(http.HandlerFunc(walletHandler.DeleteWallet)))

	// --- Wallet Analytics ---
	mux.Handle("GET /wallets/users/{userId}/highest-balance", middleware.AuthMiddleware(http.HandlerFunc(walletHandler.SearchHighestBalance)))
	mux.Handle("GET /wallets/users/{userId}/most-active", middleware.AuthMiddleware(http.HandlerFunc(walletHandler.SearchMostActive)))
	mux.Handle("GET /wallets/users/{userId}/total-balance", middleware.AuthMiddleware(http.HandlerFunc(walletHandler.SearchTotalBalance)))

	// --- Saving Goals ---
	mux.Handle("POST /saving-goals", middleware.AuthMiddleware(http.HandlerFunc(savingGoalsHandler.CreateGoal)))
	mux.Handle("GET /saving-goals", middleware.AuthMiddleware(http.HandlerFunc(savingGoalsHandler.GetAllGoals)))
	mux.Handle("GET /saving-goals/{id}", middleware.AuthMiddleware(http.HandlerFunc(savingGoalsHandler.GetGoalByID)))
	mux.Handle("PUT /saving-goals/{id}", middleware.AuthMiddleware(http.HandlerFunc(savingGoalsHandler.UpdateGoal)))
	mux.Handle("DELETE /saving-goals/{id}", middleware.AuthMiddleware(http.HandlerFunc(savingGoalsHandler.DeleteGoal)))

	// --- Budgets ---
	mux.Handle("POST /budgets", middleware.AuthMiddleware(http.HandlerFunc(budgetHandler.CreateBudget)))
	mux.Handle("GET /budgets", middleware.AuthMiddleware(http.HandlerFunc(budgetHandler.GetBudgetsByMonth)))
	mux.Handle("GET /budgets/{id}", middleware.AuthMiddleware(http.HandlerFunc(budgetHandler.GetBudgetByID)))
	mux.Handle("PUT /budgets/{id}", middleware.AuthMiddleware(http.HandlerFunc(budgetHandler.UpdateBudget)))
	mux.Handle("DELETE /budgets/{id}", middleware.AuthMiddleware(http.HandlerFunc(budgetHandler.DeleteBudget)))

	port := ":8080"
	zap.L().Info("🚀 Server WalletWise menyala dan mendengarkan di port", zap.String("Port", port))

	if err := http.ListenAndServe(port, mux); err != nil {
		zap.L().Fatal("Server mati secara tidak wajar: %v", zap.Error(err))
	}
}
