package main

import (
	"log"
	"net/http"

	// Sesuaikan nama "walletwise" dengan nama module di go.mod kamu
	categoriesService "walletwise/internal/application/categories"
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

	// A. Layer Infrastruktur (Buat Repo, kasih dia DB)
	trxRepo := postgres.NewTransactionRepo(db)
	// B. Layer Aplikasi (Buat Service, kasih dia Repo)
	trxsService := trxService.NewService(trxRepo)
	// C. Layer Transport (Buat Handler, kasih dia Service)
	trxHandler := transport.NewTransactionHandler(trxsService)

	userRepo := postgres.NewUserRepo(db)
	usersService := userService.NewService(userRepo)
	userHandler := transport.NewUserHandler(usersService)

	categoriesRepo := postgres.NewCategoriesRepo(db)
	categoriesService := categoriesService.NewService(categoriesRepo)
	categoriesHandler := transport.NewCategoriesHandler(categoriesService)

	walletRepo := postgres.NewWalletRepo(db)
	walletService := walletService.NewWalletService(walletRepo)
	walletHandler := transport.NewWalletHandler(walletService)

	// ==========================================
	// 3. SETUP ROUTER (RESEPSIONIS API)
	// ==========================================
	mux := http.NewServeMux()

	// Daftarkan semua Endpoint.
	// PENTING: Perhatikan kata kunci {id} dan {transactionID} di sini!
	// Pastikan kata di dalam kurung kurawal INI SAMA PERSIS dengan yang
	// kamu panggil di r.PathValue("...") pada file handler-mu!

	mux.HandleFunc("POST /transactions", trxHandler.CreateTransaction)
	mux.HandleFunc("GET /transactions", trxHandler.GetTransaction)
	mux.HandleFunc("GET /transactions/{transactionID}", trxHandler.GetTransactionById)
	mux.HandleFunc("PUT /transactions", trxHandler.UpdateTransaction)
	mux.HandleFunc("DELETE /transactions/{id}", trxHandler.DeleteTransaction)

	mux.HandleFunc("POST /user", userHandler.CreateUser)
	mux.HandleFunc("GET /user/{id}", userHandler.GetUserByID)
	mux.HandleFunc("GET /user/email/{email}", userHandler.GetUserByEmail)
	mux.HandleFunc("PUT /user", userHandler.UpdateUser)
	mux.HandleFunc("DELETE /user", userHandler.DeleteUser)

	mux.HandleFunc("GET /categories", categoriesHandler.GetAllCategories)

	mux.HandleFunc("GET /wallets", walletHandler.SearchAllWallets)
	mux.HandleFunc("GET /wallets/{id}", walletHandler.SearchWalletsByID)
	mux.HandleFunc("PUT /wallets/{userId}/highest-balance", walletHandler.SearchHighestBalance)
	mux.HandleFunc("PUT /wallets/{userId}/most-active", walletHandler.SearchMostActive)
	mux.HandleFunc("PUT /wallets/{userId}/total-balance", walletHandler.SearchTotalBalance)
	mux.HandleFunc("PUT /wallets", walletHandler.UpdateWallet)
	mux.HandleFunc("DELETE /wallets", walletHandler.DeleteWallet)

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
