package models

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/shopspring/decimal"
)

var counter int64

func randomName() string {
	counter++
	return fmt.Sprintf("account%d%d", counter, rand.Int63n(100000000))
}

func TestSaveAccount(t *testing.T) {
	amount, _ := decimal.NewFromString("123.321")
	a := &Account{ID: 0, CurrencyID: 1, Amount: amount, Name: randomName()}
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Unexpected error in db.Begin(): %v", err)
	}
	defer tx.Rollback()

	if err := a.Save(tx); err != nil {
		t.Errorf("Unexpected error in Account.Save: %v", err)
	}
}

func TestGetAccount(t *testing.T) {
	amount, _ := decimal.NewFromString("123.321")
	a := &Account{CurrencyID: 1, Amount: amount}
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Unexpected error in db.Begin(): %v", err)
	}
	defer tx.Rollback()

	// insert new
	if err := a.Save(tx); err != nil {
		t.Errorf("Unexpected error in Account.Save: %v", err)
	}

	a2, err := GetAccount(tx, a.ID)
	if err != nil {
		t.Errorf("Unexpected error in GetAccount: %v", err)
	}

	if a2.ID != a.ID {
		t.Errorf("Returned account id %d does not match one we saved %d", a2.ID, a.ID)
	}

	if !a2.Amount.Equals(a2.Amount) {
		t.Errorf("Returned amount %s does not match one we saved %s", a2.Amount, a.Amount)
	}

	a2.Amount = a2.Amount.Sub(a.Amount) // shoud be zero

	// update existing
	if err := a2.Save(tx); err != nil {
		t.Errorf("Unexpected error in Account.Save: %v", err)
	}

	a3, err := GetAccount(tx, a.ID)
	if err != nil {
		t.Errorf("Unexpected error in GetAccount: %v", err)
	}

	if !a3.Amount.Equals(decimal.Zero) {
		t.Errorf("Expected amount to be zero, but got %s", a3.Amount)
	}
}

func TestGetAccounts(t *testing.T) {
	amount, _ := decimal.NewFromString("567.765")
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Unexpected error in db.Begin(): %v", err)
	}
	defer tx.Rollback()

	accNumber := 100

	for i := 0; i < accNumber; i++ {
		a := &Account{CurrencyID: 1, Amount: amount, Name: randomName()}
		// insert new
		if err := a.Save(tx); err != nil {
			t.Fatalf("Unexpected error in Account.Save: %v", err)
		}
	}

	accounts, err := GetAccounts(tx)
	if err != nil {
		t.Errorf("Unexpected error in GetAccounts: %v", err)
	}

	if len(accounts) != accNumber {
		t.Errorf("Expected %d accounts, but got %d", accNumber, len(accounts))
	}

	for _, acc := range accounts {
		if !acc.Amount.Equals(amount) {
			t.Errorf("Returned amount %s does not match one we saved %s", acc.Amount, amount)
		}
		if acc.CurrencyID != 1 {
			t.Errorf("Expected CurrencyID to be 1, but got %d", acc.CurrencyID)
		}
		if acc.CurrencyName != "USD" {
			t.Errorf("Expected curreCurrencyNamency_id to be USD, but got %s", acc.CurrencyName)
		}

	}

}
