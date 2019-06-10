# Wallet service


Wallet service stores accounts' balances and allows to make payments between them.

Accounts can be identified uniquely and have currency and amount assigned to them.

Two accounts (current limitation) can participate in a payment operation where one account is a buyer (loses amount) and other is a seller (gains amount). Amount gained is equal to amount lost. Payment operation can be performed only when buyer has enough money (amount greater or equal to payment amount) in a payment currency.

Multiple pairs of accounts can request payment operations at the same time. One account can participate in multiple payment operations at the same time. There should not be any race conditions during multiple parallel operations.

Service should provide two additional features:

* listing of accounts
* listing of payments

Service should be able to work correctly with multiple instances running behind some kind of a load balancer.

## Limitations

Being a test task this service is developed with a set of limitations in mind:

* only two accounts can partitcipate in one payment operations (no exchange type orderbook trades)
* service uses shared database for all instances (SPOF, possible lock contention and performance bottleneck point). Alternative would be distributed consensus based payment operation. But it has a tricky implementation and should be tested VERY extensively because of multitude of failure modes
* no users, authentication and authorization concept introduced
* no database schema migration scaffolding
* database initialisation method (through default postgres image initdb hack) is not production ready
* no balance history table
* no delete operations supported