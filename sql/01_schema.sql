create user wallet with encrypted password 'wallet';
create database wallet with owner wallet;

\c wallet wallet

create table currencies (
    id serial primary key,
    name varchar not null unique
);

comment on table currencies is 'Currencies dictionary';

create table accounts (
    id bigserial primary key,
    name varchar not null unique,
    currency_id integer not null references currencies(id),
    amount numeric(30,15) not null, -- crazy magnitude and precision because crypto 🤑
    constraint accounts_balance_check check (amount >= 0)
);

comment on table accounts is 'Accounts with their corresponding balances';

create table payments (
    id bigserial primary key,
    currency_id integer not null references currencies(id),
    amount numeric(30,15) not null,
    buyer_account_id bigserial not null references accounts(id),
    seller_account_id bigserial not null references accounts(id),
    operation_timestamp timestamp not null default now(),
    constraint payments_amount_check check (amount > 0),
    constraint payments_diff_account_check check (buyer_account_id != seller_account_id)
);

comment on table payments is 'Payments log table';
