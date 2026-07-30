package wallet

import "context"

type Repository interface {
	SearchAll(ctx context.Context, userID UserID) ([]*Wallet, error)
	SearchByID(ctx context.Context, walletID WalletID, userId UserID) (*Wallet, error)
	Save(ctx context.Context, wallet *Wallet) error
	Update(ctx context.Context, wallet *Wallet, userId UserID) error
	Delete(ctx context.Context, wallet *Wallet, userId UserID) error

	SearchHighestBalance(ctx context.Context, userID UserID) (*Wallet, error)
	SearchMostActive(ctx context.Context, userID UserID) (*Wallet, error)
	SearchTotalBalance(ctx context.Context, userID UserID) (uint64, error)
}
