package main

import (
	"database/sql"

	"github.com/c-pro/wallet-test/models"
)

// AccountService provides methods to access accounts
type AccountService interface {
	GetAccounts() ([]models.Account, error)
	GetAccount(int64) (models.Account, error)
	CreateAccount(models.Account) error
}

// accountService implements interface above
type accountService struct {
	db *sql.DB
}

// GetAccounts returns all accounts in database
func (a *accountService) GetAccounts() ([]models.Account, error) {
	tx, err := a.db.Begin()
	if err != nil {
		return []models.Account{}, err
	}
	defer models.RollbackWithLog(tx)
	return models.GetAccounts(tx)
}

// GetAccount returns a particular account from the database
func (a *accountService) GetAccount(id int64) (models.Account, error) {
	tx, err := a.db.Begin()
	if err != nil {
		return models.Account{}, err
	}
	defer models.RollbackWithLog(tx)
	return models.GetAccount(tx, id)
}

// CreateAccount creates a new account in the database
func (a *accountService) CreateAccount(account models.Account) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer models.RollbackWithLog(tx)
	if err := account.Save(tx); err != nil {
		return err
	}
	return tx.Commit()
}
