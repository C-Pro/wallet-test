package models

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func makeAccount(tx *sql.Tx, currencyID int64, amount string) Account {
	amountD, _ := decimal.NewFromString(amount)
	a := Account{CurrencyID: currencyID, Amount: amountD, Name: randomName()}
	if err := a.Save(tx); err != nil {
		// checking error in test helper function does not worth all the fuss
		panic(fmt.Sprintf("Unexpected error in Account.Save: %v", err))
	}
	return a
}

func TestSavePayment(t *testing.T) {
	amount, _ := decimal.NewFromString("123.321")
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Unexpected error in db.Begin(): %v", err)
	}
	defer tx.Rollback()

	b := makeAccount(tx, 1, "500.0")
	s := makeAccount(tx, 1, "0")

	p := Payment{CurrencyID: b.CurrencyID,
		Amount:          amount,
		BuyerAccountID:  b.ID,
		SellerAccountID: s.ID}

	if err := p.Save(tx); err != nil {
		t.Fatalf("Unexpected error in Account.Save: %v", err)
	}

	if p.ID == 0 {
		t.Error("Payment ID should not be zero")
	}

	if time.Now().Sub(p.OperationTimestamp) > time.Minute {
		t.Errorf("Operation timestamp is far from Now: %s", p.OperationTimestamp)
	}

	if err := p.Save(tx); err != ErrPaymentNotUpdatable {
		t.Errorf("Attempt to save existing payment should fail with ErrPaymentNotUpdatable. Got %v", err)
	}
}

func TestGetPayments(t *testing.T) {
	amount, _ := decimal.NewFromString("123.321")
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Unexpected error in db.Begin(): %v", err)
	}
	defer tx.Rollback()

	b := makeAccount(tx, 1, "500.0")
	s := makeAccount(tx, 1, "0")

	p := Payment{CurrencyID: b.CurrencyID,
		Amount:          amount,
		BuyerAccountID:  b.ID,
		SellerAccountID: s.ID}

	if err := p.Save(tx); err != nil {
		t.Fatalf("Unexpected error in Account.Save: %v", err)
	}

	payments, err := GetPayments(tx)
	if err != nil {
		t.Fatalf("Unexpected error in GetPayments: %v", err)
	}

	if len(payments) != 1 {
		t.Fatalf("Expected to get 1 payment, got %d", len(payments))
	}
}

func cleanDb(t *testing.T) {
	if _, err := db.Exec("delete from payments"); err != nil {
		t.Fatalf("Failed to clean up payments table")
	}

	if _, err := db.Exec("delete from accounts"); err != nil {
		t.Fatalf("Failed to clean up accounts table")
	}
}

func TestMakePayment(t *testing.T) {
	amount, _ := decimal.NewFromString("250.00000001")
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Unexpected error in db.Begin(): %v", err)
	}

	b := makeAccount(tx, 1, "500.0")
	s := makeAccount(tx, 1, "0")
	s2 := makeAccount(tx, 2, "0")

	tx.Commit()

	defer cleanDb(t)

	_, err = MakePayment(db, b.ID, s2.ID, amount)
	if err != ErrCurrencyMismatch {
		t.Errorf("Expected MakePayment to return ErrCurrencyMismatch, got %v", err)
	}

	_, err = MakePayment(db, b.ID, b.ID, amount)
	if err != ErrNoPaymentToSelf {
		t.Errorf("Expected MakePayment to return ErrNoPaymentToSelf, got %v", err)
	}

	_, err = MakePayment(db, b.ID, s.ID, decimal.Zero)
	if err != ErrNonPositiveAmount {
		t.Errorf("Expected MakePayment to return ErrNonPositiveAmount, got %v", err)
	}

	p, err := MakePayment(db, b.ID, s.ID, amount)
	if err != nil {
		t.Errorf("Unexpected error in MakePayment: %v", err)
	}

	if !p.Amount.Equals(amount) {
		t.Errorf("Expected payment amount to be %s, but got %s", amount, p.Amount)
	}

	tx, err = db.Begin()
	if err != nil {
		t.Fatalf("Unexpected error in db.Begin(): %v", err)
	}

	b1, err := GetAccount(tx, b.ID)
	if err != nil {
		t.Fatalf("Unexpected error in GetAccount: %v", err)
	}

	expected, _ := decimal.NewFromString("249.99999999")
	if !b1.Amount.Equals(expected) {
		t.Errorf("Buyer amount expected to be %s, got %s", expected, b1.Amount)
	}

	s1, err := GetAccount(tx, s.ID)
	if err != nil {
		t.Fatalf("Unexpected error in GetAccount: %v", err)
	}

	if !s1.Amount.Equals(amount) {
		t.Errorf("Seller amount expected to be %s, got %s", amount, s1.Amount)
	}

	tx.Rollback()

	_, err = MakePayment(db, b.ID, s.ID, amount)
	if err != ErrInsufficientAmount {
		t.Errorf("Expected MakePayment to return ErrInsufficientAmount, got %v", err)
	}

}

func TestMakePaymentParallel(t *testing.T) {
	cleanDb(t)
	defer cleanDb(t)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Unexpected error in db.Begin(): %v", err)
	}

	accounts := []Account{
		makeAccount(tx, 1, "500.0"),
		makeAccount(tx, 2, "600.0"),
	}
	for i := 1; i < 98; i++ {
		accounts = append(accounts, makeAccount(tx, int64(i%2+1), "0"))
	}
	tx.Commit()

	goodErrors := make(map[error]struct{})
	goodErrors[ErrCurrencyMismatch] = struct{}{}
	goodErrors[ErrInsufficientAmount] = struct{}{}
	goodErrors[ErrNoPaymentToSelf] = struct{}{}
	goodErrors[ErrNonPositiveAmount] = struct{}{}

	var wg sync.WaitGroup
	wg.Add(100)

	for j := 0; j < 100; j++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				amount := decimal.New(rand.Int63n(1000), -2)
				bID := accounts[rand.Int63n(int64(len(accounts)))].ID
				sID := accounts[rand.Int63n(int64(len(accounts)))].ID
				_, err := MakePayment(db, bID, sID, amount)
				if _, ok := goodErrors[err]; !ok && err != nil {
					t.Fatalf("Unexpected error in MakePayment: %v", err)
				}
			}
		}()
	}
	wg.Wait()

	tx, err = db.Begin()
	if err != nil {
		t.Fatalf("Unexpected error in db.Begin(): %v", err)
	}
	defer tx.Rollback()

	sumUSD := decimal.Zero
	sumRUB := decimal.Zero

	for _, a := range accounts {
		account, err := GetAccount(tx, a.ID)
		if err != nil {
			t.Fatalf("Unexpected error in GetAccount: %v", err)
		}
		if account.CurrencyID == 1 {
			sumUSD = sumUSD.Add(account.Amount)
		} else {
			sumRUB = sumRUB.Add(account.Amount)
		}
	}

	expUSD, _ := decimal.NewFromString("500")
	if !sumUSD.Equals(expUSD) {
		t.Errorf("Sum USD is expected to be 500, but got %s", sumUSD)
	}

	expRUB, _ := decimal.NewFromString("600")
	if !sumRUB.Equals(expRUB) {
		t.Errorf("Sum RUB is expected to be 600, but got %s", sumRUB)
	}
}

func BenchmarkMakePaymentParallel(b *testing.B) {
	tx, err := db.Begin()
	if err != nil {
		b.Fatalf("Unexpected error in db.Begin(): %v", err)
	}

	accounts := []Account{
		makeAccount(tx, 1, "500.0"),
		makeAccount(tx, 2, "600.0"),
	}
	for i := 1; i < 98; i++ {
		accounts = append(accounts, makeAccount(tx, int64(i%2+1), "0"))
	}
	tx.Commit()
	b.ResetTimer()

	goodErrors := make(map[error]struct{})
	goodErrors[ErrCurrencyMismatch] = struct{}{}
	goodErrors[ErrInsufficientAmount] = struct{}{}
	goodErrors[ErrNoPaymentToSelf] = struct{}{}
	goodErrors[ErrNonPositiveAmount] = struct{}{}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			amount := decimal.NewFromFloat(rand.Float64() * 10.0)
			bID := accounts[rand.Int63n(int64(len(accounts)))].ID
			sID := accounts[rand.Int63n(int64(len(accounts)))].ID
			_, err := MakePayment(db, bID, sID, amount)
			if _, ok := goodErrors[err]; !ok && err != nil {
				b.Fatalf("Unexpected error in MakePayment: %v", err)
			}
		}
	})

}
