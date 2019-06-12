package main

import (
	"database/sql"

	"github.com/c-pro/wallet-test/models"
	"github.com/shopspring/decimal"
)

// PaymentService provides methods to access Payments
type PaymentService interface {
	GetPayments() ([]models.Payment, error)
	MakePayment(int64, int64, decimal.Decimal) (models.Payment, error)
}

// paymentService implements interface above
type paymentService struct {
	db *sql.DB
}

// GetPayments returns all payments in database
func (p *paymentService) GetPayments() ([]models.Payment, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return []models.Payment{}, err
	}
	defer tx.Rollback()
	return models.GetPayments(tx)
}

// MakePayment makes payment from one account to another
func (p *paymentService) MakePayment(buyerAccountID, sellerAccountID int64, amount decimal.Decimal) (models.Payment, error) {
	return models.MakePayment(p.db, buyerAccountID, sellerAccountID, amount)
}
