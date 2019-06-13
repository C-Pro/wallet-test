package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"

	"github.com/c-pro/wallet-test/models"
	"github.com/shopspring/decimal"
)

var counter int64

func randomName() string {
	counter++
	return fmt.Sprintf("account%d%d", counter, rand.Int63n(100000000))
}

func URL(route string) string {
	return fmt.Sprintf("%s%s", srv.URL, route)
}

func addTestAccount(t *testing.T, name string, amount decimal.Decimal) {
	c := http.DefaultClient
	req := []byte(fmt.Sprintf(`{
		"Name": "%s",
		"Amount": "%s",
		"CurrencyId": 3
	}`, name, amount))
	res, _ := c.Post(URL("/accounts"),
		"Application/json",
		bytes.NewBuffer(req))
	b, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode == 200 {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	var errResp errorResponse
	if err := json.Unmarshal(b, &errResp); err != nil {
		t.Fatal(err)
	}
	if errResp.Error != "" {
		t.Fatalf("POST /accounts returned error: %s", errResp.Error)
	}
}

func getSomething(t *testing.T, route string, dest interface{}) {
	c := http.DefaultClient
	res, err := c.Get(URL(route))
	if err != nil {
		t.Fatalf("Unexpected error in Get request: %s", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatal("Status code is not 200")
	}
	if err := json.Unmarshal(b, dest); err != nil {
		t.Fatal(err)
	}
}

func TestCreateAccount(t *testing.T) {
	name := randomName()
	amount, _ := decimal.NewFromString("10.0")
	addTestAccount(t, name, amount)
}

func TestGetAccounts(t *testing.T) {
	name := randomName()
	amount, _ := decimal.NewFromString("666.777")
	addTestAccount(t, name, amount)

	accountsResp := getAccountsResponse{}
	getSomething(t, "/accounts", &accountsResp)
	if accountsResp.Error != "" {
		t.Fatalf("GET /accounts returned unexpected error: %s", accountsResp.Error)
	}
	if len(accountsResp.Accounts) == 0 {
		t.Fatal("GET /accounts returned empty result")
	}
	found := false
	for _, account := range accountsResp.Accounts {
		if account.Name == name {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Account %s was not found in GET /accounts result", name)
	}
}

func TestGetAccount(t *testing.T) {
	name := randomName()
	amount, _ := decimal.NewFromString("42")
	addTestAccount(t, name, amount)

	accountsResp := getAccountsResponse{}
	getSomething(t, "/accounts", &accountsResp)
	if accountsResp.Error != "" {
		t.Fatalf("GET /accounts returned unexpected error: %s", accountsResp.Error)
	}
	if len(accountsResp.Accounts) == 0 {
		t.Fatal("GET /accounts returned empty result")
	}
	ID := int64(0)
	for _, account := range accountsResp.Accounts {
		if account.Name == name {
			ID = account.ID
			break
		}
	}
	if ID == 0 {
		t.Errorf("Account %s was not found in GET /accounts result", name)
	}
	accountResp := models.Account{}
	getSomething(t, fmt.Sprintf("/account/%d", ID), &accountResp)
	if accountResp.Name != name {
		t.Errorf("Account name does not match. Expected %s, got %s",
			name,
			accountResp.Name)
	}
}
