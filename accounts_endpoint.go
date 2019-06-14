package main

import (
	"context"
	"database/sql"

	"github.com/c-pro/wallet-test/models"
	"github.com/go-kit/kit/endpoint"
	"github.com/shopspring/decimal"
)

type getAccountsResponse struct {
	Accounts []models.Account `json:"Accounts,omitempty"`
}

type getAccountRequest struct {
	AccountID int64
}

type errorResponse struct {
	Error string `json:"Error,omitempty"`
	Code  int    `json:"-"`
}

type createAccountRequest struct {
	Name       string
	CurrencyID int64
	Amount     decimal.Decimal
}

func makeGetAccountsEndpoint(svc AccountService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		accounts, err := svc.GetAccounts()
		if err != nil {
			return errorResponse{err.Error(), 500}, nil
		}
		return getAccountsResponse{accounts}, nil
	}
}

func makeGetAccountEndpoint(svc AccountService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(getAccountRequest)
		account, err := svc.GetAccount(req.AccountID)
		if err != nil {
			if err == sql.ErrNoRows {
				return errorResponse{"Account not found", 404}, nil
			}
			return errorResponse{err.Error(), 500}, nil
		}
		return account, nil
	}
}

func makeCreateAccountEndpoint(svc AccountService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(createAccountRequest)
		err := svc.CreateAccount(models.Account{Name: req.Name,
			CurrencyID: req.CurrencyID,
			Amount:     req.Amount})
		if err != nil {
			return errorResponse{err.Error(), 500}, err
		}
		return struct{}{}, nil
	}
}
