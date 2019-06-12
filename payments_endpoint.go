package main

import (
	"context"

	"github.com/c-pro/wallet-test/models"
	"github.com/go-kit/kit/endpoint"
	"github.com/shopspring/decimal"
)

type getPaymentsResponse struct {
	Payments []models.Payment `json:"payments,omitempty"`
	Error    string           `json:"error,omitempty"`
}

type makePaymentRequest struct {
	BuyerAccountID  int64 `json:"buyer_account_id"`
	SellerAccountID int64 `json:"seller_account_id"`
	Amount          decimal.Decimal
}

func makeGetPaymentsEndpoint(svc PaymentService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		payments, err := svc.GetPayments()
		if err != nil {
			return getPaymentsResponse{payments, err.Error()}, nil
		}
		return getPaymentsResponse{payments, ""}, nil
	}
}

func makeMakePaymentEndpoint(svc PaymentService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(makePaymentRequest)
		payment, err := svc.MakePayment(req.BuyerAccountID, req.SellerAccountID, req.Amount)
		if err != nil {
			return errorResponse{err.Error()}, err
		}
		return payment, nil
	}
}
