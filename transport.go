package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

var errBadRoute = errors.New("bad route")
var errBadRequest = errors.New("bad request")

func makeHandlers(db *sql.DB) http.Handler {
	accSvc := &accountService{db}
	paySvc := &paymentService{db}

	getAccountsHandler := httptransport.NewServer(
		makeGetAccountsEndpoint(accSvc),
		decodeNilRequest,
		encodeResponse,
	)

	getAccountHandler := httptransport.NewServer(
		makeGetAccountEndpoint(accSvc),
		decodeGetAccountRequest,
		encodeResponse,
	)

	createAccountHandler := httptransport.NewServer(
		makeCreateAccountEndpoint(accSvc),
		decodeCreateAccountRequest,
		encodeResponse,
	)

	getPaymentsHandler := httptransport.NewServer(
		makeGetPaymentsEndpoint(paySvc),
		decodeNilRequest,
		encodeResponse,
	)

	makePaymentsHandler := httptransport.NewServer(
		makeMakePaymentEndpoint(paySvc),
		decodeMakePaymentRequest,
		encodeResponse,
	)

	r := mux.NewRouter()
	r.Handle("/accounts", getAccountsHandler).Methods("GET")
	r.Handle("/account/{id}", getAccountHandler).Methods("GET")
	r.Handle("/account", createAccountHandler).Methods("POST")
	r.Handle("/payments", getPaymentsHandler).Methods("GET")
	r.Handle("/payments", makePaymentsHandler).Methods("POST")
	return r
}

// For requests w/o bodies
func decodeNilRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func decodeGetAccountRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRoute
	}
	accountID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, errBadRequest
	}
	return getAccountRequest{accountID}, nil
}

func decodeCreateAccountRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := createAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeMakePaymentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := makePaymentRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}
