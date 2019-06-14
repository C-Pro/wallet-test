# Wallet service

[![pipeline status](https://gitlab.com/C-Pro/wallet-test/badges/master/pipeline.svg)](https://gitlab.com/C-Pro/wallet-test/commits/master)

Wallet service stores accounts' balances and allows to make payments between them.

Accounts can be identified uniquely and have currency and amount assigned to them.

Two accounts (current limitation) can participate in a payment operation where one account is a buyer (loses amount) and another is a seller (gains amount). Amount gained is equal to amount lost. Payment operation can be performed only when buyer has enough money (amount greater or equal to payment amount) in a payment currency.

Multiple pairs of accounts can request payment operations at the same time. One account can participate in multiple payment operations at the same time. There should not be any race conditions during multiple parallel operations.

Service should provide HTTP API for the following list of operations:

* listing of accounts
* creating new account
* listing of payments
* making payment

Service should be able to work correctly with multiple instances running behind some kind of a load balancer.

## API

Service provides HTTP JSON API for payments and accounts on port 8080 by default

In case of error service will return json object with `Error` string

```json
{"Error":"Account not found"}
```

### API Methods

* `GET http://localhost:8080/accounts`

    Lists all accounts in the database

    Input: None

    Output:

    ```json
    {"Accounts":[{"ID":1,"Name":"buyer","CurrencyID":1,"CurrencyName":"USD","Amount":"1000"},{"ID":2,"Name":"seller","CurrencyID":1,"CurrencyName":"USD","Amount":"0"}]}
    ```

* `GET http://localhost:8080/account/{id}`

    Get specific account info

    Input: No body. Account ID in URL

    Output:

    ```json
    {"ID":1,"Name":"buyer","CurrencyID":1,"CurrencyName":"USD","Amount":"499.9"}
    ```

* `POST http://localhost:8080/accounts`

    Create a new account

    Input:

    ```json
    {"Name":"buyer", "Amount": "1000", "CurrencyID": 1}
    ```

    Output: empty or error


* `GET http://localhost:8080/payments`

    Lists all payments in the database

    Input: None

    Output:

    ```json
    {"Payments":[{"ID":1,"CurrencyID":1,"CurrencyName":"USD","Amount":"500.1","BuyerAccountID":1,"SellerAccountID":2,"OperationTimestamp":"2019-06-13T03:21:29.933672Z"}]}
    ```


* `POST http://localhost:8080/payments`

    Makes a new payment

    Input:

    ```json
    {"BuyerAccountID":1, "SellerAccountID":2, "Amount": 500.1}
    ```

    Output:

    ```json
    {"Payments":[{"ID":1,"CurrencyID":1,"CurrencyName":"USD","Amount":"500.1","BuyerAccountID":1,"SellerAccountID":2,"OperationTimestamp":"2019-06-13T03:21:29.933672Z"}]}
    ```

## Building and running the service

To build and run the service you need to have docker and docker-compose installed.

### To build and run

```
$ docker-compose up --build
```

After image is built and started, you can proceed with trying out the service with curl

### Curl fun

List accounts

`$ curl http://localhost:8080/accounts`

```json
{}
```

Add two accounts: buyer with 1000 USD balance and seller with zero balance.

`$ curl -H "content-type: Application/json" -d '{"Name":"buyer", "Amount": "1000", "CurrencyID": 1}' http://localhost:8080/accounts`

```json
{}
```

`$ curl -H "content-type: Application/json" -d '{"Name":"seller", "Amount": 0, "CurrencyID": 1}' http://localhost:8080/accounts`

```json
{}
```

List accounts again

`$ curl http://localhost:8080/accounts`

```json
{"Accounts":[{"ID":1,"Name":"buyer","CurrencyID":1,"CurrencyName":"USD","Amount":"1000"},{"ID":2,"Name":"seller","CurrencyID":1,"CurrencyName":"USD","Amount":"0"}]}
```

Make a payment with 500.1 USD amount. Buyer balance should decrease and seller balance should increase as a result

`$ curl -H "content-type: Application/json" -d '{"BuyerAccountID":1, "SellerAccountID":2, "Amount": 500.1}' http://localhost:8080/payments`

```json
{"ID":1,"CurrencyID":1,"Amount":"500.1","BuyerAccountID":1,"SellerAccountID":2,"OperationTimestamp":"2019-06-13T03:21:29.933672Z"}
```

List payments. We see our payment now

`$ curl http://localhost:8080/payments`

```json
{"Payments":[{"ID":1,"CurrencyID":1,"CurrencyName":"USD","Amount":"500.1","BuyerAccountID":1,"SellerAccountID":2,"OperationTimestamp":"2019-06-13T03:21:29.933672Z"}]}
```

Now let's see our seller and buyer accounts balances one by one

`$ curl http://localhost:8080/account/1`

```json
{"ID":1,"Name":"buyer","CurrencyID":1,"CurrencyName":"USD","Amount":"499.9"}
```

`$ curl http://localhost:8080/account/2`

```json
{"ID":2,"Name":"seller","CurrencyID":1,"CurrencyName":"USD","Amount":"500.1"}
```

### Running tests

There are tests for models, HTTP API tests and one benchmark. Tests on models are more elaborate and do test payment operation for most error cases and highload situations.

`TestMakePaymentParallel` test creates 1 account with 500USD and 1 account with 600RUB. Then another 98 accounts with zero balances, half of them in RUB and half in USD. Then 100 goroutines launch and each does 100 payments with random amounts between two accounts selected at random at each iteration. Obviously many operations fail because of account currencies mismatch or insufficient balance, but many still do transfer money between corresponding accounts.
When all goroutines finish, accounts balances are summed up and checked if their collective balance equals 500USD and 600RUB correspondingly.

API tests are just smoke tests to make shure basic operations work as expected.

You need to have `mak`e and `go` installed to run tests.

```
$ make test
docker run -d --rm --name pg -p 5432:5432 -v /home/cpro/go/src/github.com/c-pro/wallet-test/sql:/docker-entrypoint-initdb.d postgres:11-alpine
2d1ce3525f8c60bf8d3850a3bec87a9278c05e0903d2f8a3ec3e9243a118cba4
sleep 10 # wait for pg to start up
go test -v -count 1 -race -cover ./...
=== RUN   TestCreateAccount
--- PASS: TestCreateAccount (0.01s)
=== RUN   TestGetAccounts
--- PASS: TestGetAccounts (0.02s)
=== RUN   TestGetAccount
--- PASS: TestGetAccount (0.02s)
=== RUN   TestMakePayment
--- PASS: TestMakePayment (0.04s)
=== RUN   TestGetPayments
--- PASS: TestGetPayments (0.04s)
PASS
coverage: 73.4% of statements
ok  	github.com/c-pro/wallet-test	1.154s	coverage: 73.4% of statements
=== RUN   TestSaveAccount
--- PASS: TestSaveAccount (0.00s)
=== RUN   TestGetAccount
--- PASS: TestGetAccount (0.01s)
=== RUN   TestGetAccounts
--- PASS: TestGetAccounts (0.08s)
=== RUN   TestSavePayment
--- PASS: TestSavePayment (0.00s)
=== RUN   TestGetPayments
--- PASS: TestGetPayments (0.00s)
=== RUN   TestMakePayment
--- PASS: TestMakePayment (0.02s)
=== RUN   TestMakePaymentParallel
--- PASS: TestMakePaymentParallel (8.74s)
PASS
coverage: 74.1% of statements
ok  	github.com/c-pro/wallet-test/models	9.880s	coverage: 74.1% of statements
go test -v -run Bench -bench=. ./...
PASS
ok  	github.com/c-pro/wallet-test	0.007s
goos: linux
goarch: amd64
pkg: github.com/c-pro/wallet-test/models
BenchmarkMakePaymentParallel-8   	    3000	    413359 ns/op
PASS
ok  	github.com/c-pro/wallet-test/models	1.406s
docker rm -f pg
pg
```

Benchmark shows about 0.4 ms for payment operation on my notebook, when running 3000 operations on 8 goroutines simultaneously.


### To build image

```
$ make
docker build -t gitlab.com/c-pro/wallet-test .
Sending build context to Docker daemon  17.23MB
Step 1/8 : FROM golang:1.12.5-alpine3.9 as builder
 ---> c7330979841b
Step 2/8 : ADD . /build
 ---> 1f9ca6c6591e
Step 3/8 : WORKDIR /build
 ---> Running in 0fa39d8b5294
Removing intermediate container 0fa39d8b5294
 ---> d0251789d96c
Step 4/8 : RUN GO111MODULE=on CGO_ENABLED=0 go build -mod=vendor -o wallet .
 ---> Running in 8b89117ac258
Removing intermediate container 8b89117ac258
 ---> defd6cd22e93
Step 5/8 : FROM scratch
 --->
Step 6/8 : EXPOSE 8080
 ---> Using cache
 ---> a8b29f7bf1e9
Step 7/8 : COPY --from=builder /build/wallet /
 ---> 423705b0c714
Step 8/8 : CMD ["/wallet"]
 ---> Running in fc5dcf5a3967
Removing intermediate container fc5dcf5a3967
 ---> 97eb1cfcb3bc
Successfully built 97eb1cfcb3bc
Successfully tagged gitlab.com/c-pro/wallet-test:latest
```

## Limitations

Being a test task this service is developed with a set of limitations in mind:

* only two accounts can partitcipate in one payment operation (no exchange type orderbook trades)
* service uses shared database for all instances (SPOF, possible lock contention and performance bottleneck point). Alternative would be distributed consensus based payment operation. But it has a tricky implementation and should be tested VERY extensively because of multitude of failure modes
* no proper logging and instrumentation
* errors are not wrapped with origin function names etc.
* no users, authentication and authorization concepts introduced
* no database schema migration scaffolding
* database initialization method (through default postgres image initdb hack) is not production ready
* features missing: paging, search (filters), no balance history, no soft delete operations supported, no API for currencies
