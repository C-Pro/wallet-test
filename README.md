# Wallet service


Wallet service stores accounts' balances and allows to make payments between them.

Accounts can be identified uniquely and have currency and amount assigned to them.

Two accounts (current limitation) can participate in a payment operation where one account is a buyer (loses amount) and other is a seller (gains amount). Amount gained is equal to amount lost. Payment operation can be performed only when buyer has enough money (amount greater or equal to payment amount) in a payment currency.

Multiple pairs of accounts can request payment operations at the same time. One account can participate in multiple payment operations at the same time. There should not be any race conditions during multiple parallel operations.

Service should provide two additional features:

* listing of accounts
* listing of payments

Service should be able to work correctly with multiple instances running behind some kind of a load balancer.

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

Add two accounts. Wealthy buyer with 1000 BTC balance. And seller w/o any money on BTC account.

```
curl -H "content-type: Application/json" -d '{"name":"buyer", "amount": "1000", "currency_id": 1}' http://localhost:8080/accounts
{}

curl -H "content-type: Application/json" -d '{"name":"seller", "amount": 0, "currency_id": 1}' http://localhost:8080/accounts
{}
```

List accounts again

`curl http://localhost:8080/accounts`

```json
{"accounts":[{"ID":1,"Name":"buyer","CurrencyID":1,"CurrencyName":"USD","Amount":"1000"},{"ID":2,"Name":"seller","CurrencyID":1,"CurrencyName":"USD","Amount":"0"}]}
```

Make payment with 500.1 BTC amount. Buyer balance should decrease and seller balance should increase as a result

`$ curl -H "content-type: Application/json" -d '{"buyer_account_id":1, "seller_account_id":2, "amount": 500.1}' http://localhost:8080/payments`

```json
{"ID":1,"CurrencyID":1,"Amount":"500.1","BuyerAccountID":1,"SellerAccountID":2,"OperationTimestamp":"2019-06-13T03:21:29.933672Z"}
```

List payments. We see our payment now

`curl http://localhost:8080/payments`

```json
{"payments":[{"ID":1,"CurrencyID":1,"currency_name":"USD","Amount":"500.1","BuyerAccountID":1,"SellerAccountID":2,"OperationTimestamp":"2019-06-13T03:21:29.933672Z"}]}
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

* only two accounts can partitcipate in one payment operations (no exchange type orderbook trades)
* service uses shared database for all instances (SPOF, possible lock contention and performance bottleneck point). Alternative would be distributed consensus based payment operation. But it has a tricky implementation and should be tested VERY extensively because of multitude of failure modes
* no users, authentication and authorization concept introduced
* no proper logging and instrumentation
* no database schema migration scaffolding
* database initialization method (through default postgres image initdb hack) is not production ready
* no balance history table
* no delete operations supported
