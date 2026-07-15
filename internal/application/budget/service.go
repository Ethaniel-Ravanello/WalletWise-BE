package budget

import (
	"context"
	"errors"
	"time"

	// Sesuaikan path import domain budget lu
	"walletwise/internal/domain/budget"
)

// --- DTO Inputs ---
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

type BudgetDetailResponse struct {
	ID            uint64 `json:"id"`
	UserID        uint64 `json:"user_id"`
	CategoryID    uint64 `json:"category_id"`
	Month         int    `json:"month"`
	Year          int    `json:"year"`
	MaxAmount     int64  `json:"max_amount"`
	CurrentAmount int64  `json:"current_amount"`
	Remaining     int64  `json:"remaining"`
}

type Service struct {
	repo budget.Repository
}

func NewService(repo budget.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateBudget membuat budget baru dan mengembalikan detail DTO-nya
func (s *Service) CreateBudget(ctx context.Context, input BudgetInput) (*BudgetDetailResponse, error) {
	existingBudget, err := s.repo.FindByUserAndCategory(
		ctx,
		budget.UserID(input.UserID),
		budget.CategoryID(input.CategoryID),
		input.Month,
		input.Year,
	)

	if err == nil && existingBudget != nil {
		return nil, errors.New("budget for this category in the specified month and year already exists")
	}

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

	walletId, err := s.repo.Save(ctx, newBudget)

	// Hitung pengeluaran (berjaga-jaga jika user membuat budget secara retroaktif untuk transaksi yang sudah ada)
	totalSpent, err := s.repo.CalculateTotalSpent(ctx, uint64(walletId), input.CategoryID, input.Month, input.Year)
	if err != nil {
		return nil, err
	}
	return &BudgetDetailResponse{
		ID:            uint64(newBudget.ID()),
		UserID:        uint64(newBudget.UserID()),
		CategoryID:    uint64(newBudget.CategoryID()),
		Month:         newBudget.Month(),
		Year:          newBudget.Year(),
		MaxAmount:     newBudget.Amount(),
		CurrentAmount: totalSpent,
		Remaining:     newBudget.Amount() - totalSpent,
	}, nil
}

// GetBudgetByID mengambil satu data budget dan menyertakan perhitungan real-time dari transaksi
func (s *Service) GetBudgetByID(ctx context.Context, id uint64) (*BudgetDetailResponse, error) {
	b, err := s.repo.FindByID(ctx, budget.BudgetID(id))
	if err != nil {
		return nil, err
	}

	// Ambil total pengeluaran dari transaksi
	totalSpent, _ := s.repo.CalculateTotalSpent(ctx, uint64(b.UserID()), uint64(b.CategoryID()), b.Month(), b.Year())

	return &BudgetDetailResponse{
		ID:            uint64(b.ID()),
		UserID:        uint64(b.UserID()),
		CategoryID:    uint64(b.CategoryID()),
		Month:         b.Month(),
		Year:          b.Year(),
		MaxAmount:     b.Amount(),
		CurrentAmount: totalSpent,
		Remaining:     b.Amount() - totalSpent,
	}, nil
}

// GetBudgetsByMonth mengambil semua budget user per bulan dan mengkonversinya ke list DTO
func (s *Service) GetBudgetsByMonth(ctx context.Context, userID uint64, month int, year int) ([]*BudgetDetailResponse, error) {
	budgets, err := s.repo.FindByUserAndMonth(ctx, budget.UserID(userID), month, year)
	if err != nil {
		return nil, err
	}

	if len(budgets) == 0 {
		return nil, errors.New("no budgets found for this month")
	}

	var responses []*BudgetDetailResponse
	for _, b := range budgets {
		// Looping untuk menghitung pengeluaran tiap-tiap kategori budget
		totalSpent, _ := s.repo.CalculateTotalSpent(ctx, uint64(b.UserID()), uint64(b.CategoryID()), b.Month(), b.Year())

		responses = append(responses, &BudgetDetailResponse{
			ID:            uint64(b.ID()),
			UserID:        uint64(b.UserID()),
			CategoryID:    uint64(b.CategoryID()),
			Month:         b.Month(),
			Year:          b.Year(),
			MaxAmount:     b.Amount(),
			CurrentAmount: totalSpent,
			Remaining:     b.Amount() - totalSpent,
		})
	}

	return responses, nil
}

// UpdateBudget memperbarui data budget (Kategori, Bulan, Tahun, Jumlah)
func (s *Service) UpdateBudget(ctx context.Context, input BudgetUpdateInput) error {
	existingBudget, err := s.repo.FindByID(ctx, budget.BudgetID(input.ID))
	if err != nil {
		return err
	}

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

		if err == nil && checkDuplicate != nil && checkDuplicate.ID() != existingBudget.ID() {
			return errors.New("another budget for this category and month already exists")
		}
	}

	err = existingBudget.UpdateBudget(
		budget.CategoryID(input.CategoryID),
		input.Month,
		input.Year,
		input.Amount,
	)
	if err != nil {
		return err
	}

	return s.repo.Update(ctx, existingBudget)
}

// DeleteBudget menghapus budget berdasarkan ID
func (s *Service) DeleteBudget(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, budget.BudgetID(id))
}
