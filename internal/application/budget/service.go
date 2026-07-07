package budget

import (
	"context"
	"errors"
	"time"

	// Sesuaikan path import domain budget lu
	"walletwise/internal/domain/budget"
)

// --- DTO Inputs ---
// Pastikan huruf depannya Kapital agar bisa dibaca oleh layer Transport (Handler)

type BudgetInput struct {
	UserID     uint64
	CategoryID uint64
	Month      int
	Year       int
	Amount     int64
}

type BudgetUpdateInput struct {
	ID         uint64
	CategoryID uint64
	Month      int
	Year       int
	Amount     int64
}

// --- Service ---

type Service struct {
	repo budget.Repository
}

func NewService(repo budget.Repository) *Service {
	return &Service{repo: repo}
}

// CreateBudget membuat budget baru dengan validasi duplikasi
func (s *Service) CreateBudget(ctx context.Context, input BudgetInput) (*budget.Budget, error) {
	// Validasi Bisnis: Cek apakah budget untuk kategori di bulan & tahun ini sudah ada
	existingBudget, err := s.repo.FindByUserAndCategory(
		ctx,
		budget.UserID(input.UserID),
		budget.CategoryID(input.CategoryID),
		input.Month,
		input.Year,
	)

	// Jika tidak ada error dan datanya ketemu, berarti duplikat
	if err == nil && existingBudget != nil {
		return nil, errors.New("budget for this category in the specified month and year already exists")
	}

	// Buat entity budget baru
	newBudget, err := budget.NewBudget(
		budget.UserID(input.UserID),
		budget.CategoryID(input.CategoryID),
		input.Month,
		input.Year,
		input.Amount,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return nil, err
	}

	// Simpan ke database
	if err := s.repo.Save(ctx, newBudget); err != nil {
		return nil, err
	}

	return newBudget, nil
}

// GetBudgetByID mengambil satu data budget berdasarkan ID-nya
func (s *Service) GetBudgetByID(ctx context.Context, id uint64) (*budget.Budget, error) {
	b, err := s.repo.FindByID(ctx, budget.BudgetID(id))
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GetBudgetsByMonth mengambil semua budget milik user pada bulan dan tahun tertentu
func (s *Service) GetBudgetsByMonth(ctx context.Context, userID uint64, month int, year int) ([]*budget.Budget, error) {
	budgets, err := s.repo.FindByUserAndMonth(ctx, budget.UserID(userID), month, year)
	if err != nil {
		return nil, err
	}

	if len(budgets) == 0 {
		return nil, errors.New("no budgets found for this month")
	}

	return budgets, nil
}

// UpdateBudget memperbarui data budget (Kategori, Bulan, Tahun, Jumlah)
func (s *Service) UpdateBudget(ctx context.Context, input BudgetUpdateInput) error {
	// 1. Cari data aslinya dulu
	existingBudget, err := s.repo.FindByID(ctx, budget.BudgetID(input.ID))
	if err != nil {
		return err
	}

	// 2. Cek apakah ada perubahan yang bisa menyebabkan duplikasi (ganti kategori atau bulan)
	if existingBudget.CategoryID() != budget.CategoryID(input.CategoryID) ||
		existingBudget.Month() != input.Month ||
		existingBudget.Year() != input.Year {

		checkDuplicate, err := s.repo.FindByUserAndCategory(
			ctx,
			existingBudget.UserID(),
			budget.CategoryID(input.CategoryID),
			input.Month,
			input.Year,
		)

		// Jika ketemu budget lain dengan kategori/bulan/tahun yang sama, tolak update
		if err == nil && checkDuplicate != nil && checkDuplicate.ID() != existingBudget.ID() {
			return errors.New("another budget for this category and month already exists")
		}
	}

	// 3. Update nilai di Entity
	err = existingBudget.UpdateBudget(
		budget.CategoryID(input.CategoryID),
		input.Month,
		input.Year,
		input.Amount,
	)
	if err != nil {
		return err
	}

	// 4. Simpan perubahan ke database
	return s.repo.Update(ctx, existingBudget)
}

// DeleteBudget menghapus budget berdasarkan ID
func (s *Service) DeleteBudget(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, budget.BudgetID(id))
}
