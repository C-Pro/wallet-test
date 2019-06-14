package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/c-pro/wallet-test/models"
	"github.com/shopspring/decimal"
)

func addTestPayment(t *testing.T, amount decimal.Decimal) models.Payment {
	c := http.DefaultClient
	sellerName := randomName()
	buyerName := randomName()
	addTestAccount(t, buyerName, amount)
	addTestAccount(t, sellerName, decimal.Zero)

	accountsResp := getAccountsResponse{}
	getSomething(t, "/accounts", &accountsResp)
	if len(accountsResp.Accounts) == 0 {
		t.Fatal("GET /accounts returned empty result")
	}
	buyerAccountID := int64(0)
	sellerAccountID := int64(0)
	for _, account := range accountsResp.Accounts {
		if account.Name == buyerName {
			buyerAccountID = account.ID
		}
		if account.Name == sellerName {
			sellerAccountID = account.ID
		}
	}
	if buyerAccountID == 0 || sellerAccountID == 0 {
		t.Fatal("Accounts were not found in GET /accounts result")
	}

	req := []byte(fmt.Sprintf(`{
		"BuyerAccountID": %d,
		"SellerAccountID": %d,
		"Amount": %s}`,
		buyerAccountID,
		sellerAccountID,
		amount))
	res, _ := c.Post(URL("/payments"),
		"Application/json",
		bytes.NewBuffer(req))
	b, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 200 {
		t.Logf("URL: %s", URL("/payments"))
		t.Logf("body: %s", string(b))
		t.Fatalf("Error code %d", res.StatusCode)
	}
	if err != nil {
		t.Fatal(err)
	}
	payment := models.Payment{}
	if err := json.Unmarshal(b, &payment); err != nil {
		t.Fatal(err)
	}
	return payment
}

func TestMakePayment(t *testing.T) {
	amount, _ := decimal.NewFromString("10.0")
	payment := addTestPayment(t, amount)
	if !payment.Amount.Equals(amount) {
		t.Errorf("Expected amount %s, got %s", amount, payment.Amount)
	}
}

func TestGetPayments(t *testing.T) {
	satoshi, _ := decimal.NewFromString("0.00000001")
	payment := addTestPayment(t, satoshi)
	if !payment.Amount.Equals(satoshi) {
		t.Errorf("Expected amount %s, got %s", satoshi, payment.Amount)
	}

	paymentsResp := getPaymentsResponse{}
	getSomething(t, "/payments", &paymentsResp)
	if len(paymentsResp.Payments) == 0 {
		t.Fatal("GET /payments returned empty result")
	}
	found := false
	for _, pay := range paymentsResp.Payments {
		if pay.ID == payment.ID && pay.Amount.Equals(payment.Amount) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Payment %d was not found in GET /payments result", payment.ID)
	}
}
