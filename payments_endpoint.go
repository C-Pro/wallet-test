package main

import (
	"context"
	"database/sql"

	"github.com/c-pro/wallet-test/models"
	"github.com/go-kit/kit/endpoint"
	"github.com/shopspring/decimal"
)

type getPaymentsResponse struct {
	Payments []models.Payment `json:"Payments,omitempty"`
}

type makePaymentRequest struct {
	BuyerAccountID  int64
	SellerAccountID int64
	Amount          decimal.Decimal
}

func makeGetPaymentsEndpoint(svc PaymentService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		payments, err := svc.GetPayments()
		if err != nil {
			return errorResponse{err.Error(), 500}, nil
		}
		return getPaymentsResponse{payments}, nil
	}
}

func makeMakePaymentEndpoint(svc PaymentService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(makePaymentRequest)
		payment, err := svc.MakePayment(req.BuyerAccountID, req.SellerAccountID, req.Amount)
		if err != nil {
			if err == models.ErrCurrencyMismatch ||
				err == models.ErrInsufficientAmount ||
				err == models.ErrNoPaymentToSelf ||
				err == models.ErrNonPositiveAmount ||
				err == sql.ErrNoRows {
				return errorResponse{err.Error(), 400}, nil
			}
			return errorResponse{err.Error(), 500}, nil
		}
		return payment, nil
	}
}
