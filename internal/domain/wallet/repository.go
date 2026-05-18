package wallet

import "context"

type Repository interface {
	SearchAll(ctx context.Context, userID UserID) ([]*Wallet, error)
	SearchByID(ctx context.Context, userID UserID, walletID WalletID) (*Wallet, error)
	Save(ctx context.Context, wallet *Wallet) error
	Update(ctx context.Context, wallet *Wallet) error
	Delete(ctx context.Context, wallet *Wallet) error

	SearchHighestBalance(ctx context.Context, userID UserID) (*Wallet, error)
	SearchMostActive(ctx context.Context, userID UserID) (*Wallet, error)
	SearchTotalBalance(ctx context.Context, userID UserID) (*Wallet, error)
}
