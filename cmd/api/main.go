package main

import (
	"log"
	"net/http"

	// Sesuaikan nama "walletwise" dengan nama module di go.mod kamu
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
	// ==========================================
	// 1. INISIASI DATABASE
	// ==========================================
	db, err := postgres.InitDatabase()
	if err != nil {
		// Kalau gagal connect DB, matikan aplikasi seketika!
		log.Fatalf("Gagal menyalakan database: %v", err)
	}
	defer db.Close() // Pastikan pintu database ditutup saat server mati

	if err := postgres.RunMigrations(db); err != nil {
		log.Fatalf("Database Migration Error: %v", err)
	}

	// ==========================================
	// 2. DEPENDENCY INJECTION (PERAKITAN)
	// ==========================================

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

	// ==========================================
	// 3. SETUP ROUTER (RESEPSIONIS API)
	// ==========================================
	mux := http.NewServeMux()

	// --- TRANSACTIONS ---
	mux.HandleFunc("POST /transactions", trxHandler.CreateTransaction)
	mux.HandleFunc("GET /transactions", trxHandler.GetTransactions)         // Asumsi nama fungsi GetTransactions (jamak)
	mux.HandleFunc("GET /transactions/{id}", trxHandler.GetTransactionById) // FIX: {transactionID} diganti jadi {id}
	mux.HandleFunc("PUT /transactions/{id}", trxHandler.UpdateTransaction)  // FIX: Tambah {id}
	mux.HandleFunc("DELETE /transactions/{id}", trxHandler.DeleteTransaction)

	// --- USERS --- (Diubah jadi jamak /users agar sesuai standar REST)
	mux.HandleFunc("POST /users", userHandler.CreateUser)
	mux.HandleFunc("GET /users/{id}", userHandler.GetUserByID)
	mux.HandleFunc("GET /users/email/{email}", userHandler.GetUserByEmail)
	mux.HandleFunc("PUT /users/{id}", userHandler.UpdateUser)    // FIX: Tambah {id}
	mux.HandleFunc("DELETE /users/{id}", userHandler.DeleteUser) // FIX: Tambah {id}

	// --- CATEGORIES ---
	mux.HandleFunc("GET /categories", categoriesHandler.GetAllCategories)

	// --- WALLETS ---
	mux.HandleFunc("POST /wallets", walletHandler.CreateWallets) // FIX: Rute POST sebelumnya ketinggalan
	mux.HandleFunc("GET /wallets", walletHandler.SearchAllWallets)
	mux.HandleFunc("GET /wallets/{id}", walletHandler.SearchWalletsByID)
	mux.HandleFunc("PUT /wallets/{id}", walletHandler.UpdateWallet)    // FIX: Tambah {id}
	mux.HandleFunc("DELETE /wallets/{id}", walletHandler.DeleteWallet) // FIX: Tambah {id}

	// Fitur Aggregasi Wallet (Diubah jadi GET karena sifatnya mengambil data, bukan update data)
	mux.HandleFunc("GET /wallets/users/{userId}/highest-balance", walletHandler.SearchHighestBalance)
	mux.HandleFunc("GET /wallets/users/{userId}/most-active", walletHandler.SearchMostActive)
	mux.HandleFunc("GET /wallets/users/{userId}/total-balance", walletHandler.SearchTotalBalance)

	// --- SAVING GOALS ---
	mux.HandleFunc("POST /saving-goals", savingGoalsHandler.CreateGoal)
	mux.HandleFunc("GET /saving-goals", savingGoalsHandler.GetAllGoals)
	mux.HandleFunc("GET /saving-goals/{id}", savingGoalsHandler.GetGoalByID)
	mux.HandleFunc("PUT /saving-goals/{id}", savingGoalsHandler.UpdateGoal)
	mux.HandleFunc("DELETE /saving-goals/{id}", savingGoalsHandler.DeleteGoal)

	// ==========================================
	// --- BUDGETS ---
	// ==========================================
	mux.HandleFunc("POST /budgets", budgetHandler.CreateBudget)
	// Mengambil list budget bulanan (Gunakan query params saat manggil API-nya)
	// Contoh request: GET /budgets?user_id=1&month=7&year=2026
	mux.HandleFunc("GET /budgets", budgetHandler.GetBudgetsByMonth)
	// Operasi spesifik per ID Budget
	mux.HandleFunc("GET /budgets/{id}", budgetHandler.GetBudgetByID)
	mux.HandleFunc("PUT /budgets/{id}", budgetHandler.UpdateBudget)
	mux.HandleFunc("DELETE /budgets/{id}", budgetHandler.DeleteBudget)

	// ==========================================
	// 4. NYALAKAN SERVER
	// ==========================================
	port := ":8080"
	log.Println("🚀 Server WalletWise menyala dan mendengarkan di port", port)

	// Fungsi ini akan memblokir kode agar terus berjalan selamanya
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server mati secara tidak wajar: %v", err)
	}
}
