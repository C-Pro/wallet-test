package models

import (
	"context"
	"database/sql"

	"github.com/shopspring/decimal"
)

// Account is a representation of a particular account balance in currency
type Account struct {
	ID           int64
	Name         string
	CurrencyID   int64
	CurrencyName string
	Amount       decimal.Decimal
}

// GetAccounts returns all accounts from the database
func GetAccounts(tx *sql.Tx) ([]Account, error) {
	accounts := []Account{}
	query := `select a.id,
					 a.currency_id,
					 c.name,
					 a.amount,
					 a.name
				from accounts a
				join currencies c on (a.currency_id = c.id)
				order by a.id`
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return accounts, err
	}
	defer rows.Close()
	for rows.Next() {
		account := Account{}
		err := rows.Scan(&account.ID,
			&account.CurrencyID,
			&account.CurrencyName,
			&account.Amount,
			&account.Name,
		)
		if err != nil {
			// If it was a context timeout, return context error
			if ctx.Err() != nil {
				err = ctx.Err()
			}
			return accounts, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

// Save inserts or updates Account record in the database
// if Account.ID is zero, new record is created
// otherwise existing record is updated
func (a *Account) Save(tx *sql.Tx) error {
	query := `update accounts
			  set amount = $1,
			      name = $2
			  where id = $3
			  returning id`
	params := []interface{}{a.Amount, a.Name, a.ID}
	if a.ID == 0 {
		query = `insert into accounts(currency_id, amount, name)
			  values($1, $2, $3)
			  returning id`
		params = []interface{}{a.CurrencyID, a.Amount, a.Name}
	}
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	err := tx.QueryRowContext(ctx, query, params...).Scan(&a.ID)
	if err != nil {
		// If it was a context timeout, return context error
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return err
	}

	return nil
}

// GetAccount returns account with given ID from the database
func GetAccount(tx *sql.Tx, id int64) (Account, error) {
	account := Account{}
	query := `select a.id,
					 a.currency_id,
					 c.name,
					 a.amount,
					 a.name
				from accounts a
				join currencies c on (a.currency_id = c.id)
				where a.id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	err := tx.QueryRowContext(ctx, query, id).Scan(&account.ID,
		&account.CurrencyID,
		&account.CurrencyName,
		&account.Amount,
		&account.Name)
	if err != nil {
		// If it was a context timeout, return context error
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return account, err
	}
	return account, nil
}

// lockAccountsForTransaction acquires lock for a pair of accounts and returns true.
// If one or both of accounts are locked it will return false and we need to retry attempt later.
// Transaction should be rolled back if we get false, otherwise we can hold lock for one of
// the accounts for no reason.
func lockAccountsForTransaction(tx *sql.Tx, id1, id2 int64) (bool, error) {
	count := 0

	query := `select count(*) from accounts
			   where id in ($1, $2)`
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	err := tx.QueryRowContext(ctx, query, id1, id2).Scan(&count)
	if err != nil {
		// If it was a context timeout, return context error
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		cancel()
		return false, err
	}
	cancel()
	if count != 2 {
		return false, sql.ErrNoRows
	}

	query = `select count(*) from
			  (select * from accounts
			   where id in ($1, $2)
			  for update skip locked) v`
	ctx, cancel = context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	err = tx.QueryRowContext(ctx, query, id1, id2).Scan(&count)
	if err != nil {
		// If it was a context timeout, return context error
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return false, err
	}
	return count == 2, nil
}
