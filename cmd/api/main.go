package main

import (
	"log"
	"net/http"
	// Sesuaikan nama "walletwise" dengan nama module di go.mod kamu
	appService "walletwise/internal/application/transaction"
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
	trxService := appService.NewService(trxRepo)

	// C. Layer Transport (Buat Handler, kasih dia Service)
	trxHandler := transport.NewTransactionHandler(trxService)

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
