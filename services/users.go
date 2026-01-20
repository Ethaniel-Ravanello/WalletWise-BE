package services

import (
	"errors"
	"walletwise/models"
	"walletwise/repositories"
)

type UserService struct {
	userRepo       repositories.UserRepository
	incomeRepo     repositories.IncomeRepository
	expenseRepo    repositories.ExpenseRepository
	savingRepo     repositories.SavingGoalRepository
	savingGoalRepo repositories.SavingGoalRepository
}

func NewUserService(repo repositories.UserRepository) *UserService {
	return &UserService{userRepo: repo}
}

func (s *UserService) GetUserDetail(userId int) (*models.User, error) {
	user, err := s.userRepo.GetUserById(userId)

	if err != nil {
		return nil, errors.New("User Not Found")
	}
	return user, nil
}

func (s *UserService) GetUserFinancial(userId int) (map[string]interface{}, error) {
	user, err := s.userRepo.GetUserById(userId)
	if err != nil {
		return nil, errors.New("User Not Found")
	}
	userIncome, err := s.incomeRepo.GetAllIncomes(userId)
	if err != nil {
		return nil, errors.New("Income Not Found")
	}
	userExpense, err := s.expenseRepo.GetAllExpense(userId)
	if err != nil {
		return nil, errors.New("Expense Not Found")
	}

	someIncomes := userIncome[0:3]
	someExpenses := userExpense[0:3]

	var totalIncomes float64
	var totalExpenses float64

	for _, income := range someIncomes {
		totalIncomes += income.Amount
	}

	for _, expense := range someExpenses {
		totalExpenses += expense.Amount
	}

	return map[string]interface{}{
			"user":          user,
			"income":        someIncomes,
			"expense":       someExpenses,
			"totalIncomes":  totalIncomes,
			"totalExpenses": totalExpenses,
		},
		nil
}

func (s *UserService) CreateUser(user *models.User) error {
	err := s.userRepo.CreateUser(user)
	if err != nil {
		return errors.New("Internal Server Error")
	}
	return nil
}

func (s *UserService) UpdateUser(user *models.User) error {
	err := s.userRepo.CreateUser(user)
	if err != nil {
		return errors.New("Internal Server Error")
	}
	return nil
}

func (s *UserService) DeleteUser(userId int) error {
	err := s.userRepo.DeleteUserById(userId)
	if err != nil {
		return errors.New("Internal Server Error")
	}
	return nil
}
