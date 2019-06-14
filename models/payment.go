package models

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
)

// Constant errors
var (
	ErrInsufficientAmount  = errors.New("Buyer account does not have sufficient balance")
	ErrCurrencyMismatch    = errors.New("Payment allowed only when account currencies match")
	ErrPaymentNotUpdatable = errors.New("Payments can not be updated")
	ErrNoPaymentToSelf     = errors.New("Could not make payment to self")
	ErrNonPositiveAmount   = errors.New("Amount should be positive")
	ErrLockFailed          = errors.New("Failed to acquire lock on accounts")
)

// Payment is a representation of a payment operation, transferring amount from buyer account to seller account
type Payment struct {
	ID                 int64
	CurrencyID         int64
	CurrencyName       string `json:"CurrencyName,omitempty"`
	Amount             decimal.Decimal
	BuyerAccountID     int64
	SellerAccountID    int64
	OperationTimestamp time.Time
}

// GetPayments returns all payments from a database
func GetPayments(tx *sql.Tx) ([]Payment, error) {
	payments := []Payment{}
	query := `select p.id,
					 p.currency_id,
					 c.name,
					 p.amount,
					 p.buyer_account_id,
					 p.seller_account_id,
					 p.operation_timestamp
				from payments p
				join currencies c on (p.currency_id = c.id)
				order by p.operation_timestamp`
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return payments, err
	}
	defer rows.Close()
	for rows.Next() {
		payment := Payment{}
		err := rows.Scan(&payment.ID,
			&payment.CurrencyID,
			&payment.CurrencyName,
			&payment.Amount,
			&payment.BuyerAccountID,
			&payment.SellerAccountID,
			&payment.OperationTimestamp,
		)
		if err != nil {
			// If it was a context timeout, return context error
			if ctx.Err() != nil {
				err = ctx.Err()
			}
			return payments, err
		}
		payments = append(payments, payment)
	}
	return payments, nil
}

// Save inserts Payment record in the database
func (p *Payment) Save(tx *sql.Tx) error {
	if p.ID != 0 {
		return ErrPaymentNotUpdatable
	}
	query := `insert into payments(currency_id,
								  amount,
								  buyer_account_id,
								  seller_account_id)
			values($1, $2, $3, $4)
			returning id, operation_timestamp`

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	err := tx.QueryRowContext(ctx, query,
		p.CurrencyID,
		p.Amount,
		p.BuyerAccountID,
		p.SellerAccountID).Scan(&p.ID, &p.OperationTimestamp)
	if err != nil {
		// If it was a context timeout, return context error
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return err
	}
	return nil
}

// MakePayment makes atomic payment operation for given amount between seller and buyer accounts
// Prior to commencing operation lock on both accounts is acquired and some sanity checks are performed
func MakePayment(db *sql.DB,
	buyerAccountID,
	sellerAccountID int64,
	amount decimal.Decimal) (Payment, error) {

	payment := Payment{}

	// could not pay to self
	if buyerAccountID == sellerAccountID {
		return payment, ErrNoPaymentToSelf
	}

	// amount should be greater then zero
	if amount.Cmp(decimal.Zero) <= 0 {
		return payment, ErrNonPositiveAmount
	}

	var tx *sql.Tx

	// try to lock both accounts
	numRetries := uint(10)
	for tryNum := uint(1); tryNum <= numRetries; tryNum++ {
		var err error
		tx, err = db.Begin()
		if err != nil {
			return payment, err
		}

		// lock both accounts
		success, err := lockAccountsForTransaction(tx, buyerAccountID, sellerAccountID)
		if err != nil || !success {
			RollbackWithLog(tx)
		}

		if err != nil {
			return payment, err
		}

		// if lock for both accounts is acquired, proceed
		if success {
			break
		}

		// failed to acquire after numRetries
		if tryNum == numRetries {
			return payment, ErrLockFailed
		}

		interval := (1<<tryNum)*5 + rand.Intn((1<<tryNum)*5)
		time.Sleep(time.Millisecond * time.Duration(interval))
	}

	defer RollbackWithLog(tx)

	buyer, err := GetAccount(tx, buyerAccountID)
	if err != nil {
		return payment, err
	}

	seller, err := GetAccount(tx, sellerAccountID)
	if err != nil {
		return payment, err
	}

	// buyer account should have enough money
	if buyer.Amount.Cmp(amount) < 0 {
		return payment, ErrInsufficientAmount
	}

	// buyer and seller currencies should match
	if buyer.CurrencyID != seller.CurrencyID {
		return payment, ErrCurrencyMismatch
	}

	buyer.Amount = buyer.Amount.Sub(amount)
	seller.Amount = seller.Amount.Add(amount)

	if err := buyer.Save(tx); err != nil {
		return payment, err
	}

	if err := seller.Save(tx); err != nil {
		return payment, err
	}

	payment.CurrencyID = buyer.CurrencyID
	payment.BuyerAccountID = buyerAccountID
	payment.SellerAccountID = sellerAccountID
	payment.Amount = amount

	if err := payment.Save(tx); err != nil {
		return payment, err
	}

	return payment, tx.Commit()
}

// RollbackWithLog rolls back transaction and logs error if any. For use in defer statement.
func RollbackWithLog(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		log.Printf("Error on tx.Rollback: %v", err)
	}
}
