CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    balance NUMERIC(15,2) DEFAULT 0.00,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE credits (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    amount NUMERIC(15,2) NOT NULL,
    rate NUMERIC(5,2) NOT NULL,
    term_months INT NOT NULL,
    monthly_payment NUMERIC(15,2) NOT NULL,
    remaining NUMERIC(15,2) NOT NULL,
    next_payment_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'active'
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_account_id UUID REFERENCES accounts(id),
    to_account_id UUID REFERENCES accounts(id),
    credit_id UUID REFERENCES credits(id),
    amount NUMERIC(15,2) NOT NULL,
    type VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE cards (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id),
    encrypted_number BYTEA NOT NULL,
    hmac_number TEXT NOT NULL,
    encrypted_expiry BYTEA NOT NULL,
    cvv_hash TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    last_four CHAR(4) NOT NULL
);

CREATE TABLE payment_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_id UUID NOT NULL REFERENCES credits(id),
    due_date DATE NOT NULL,
    amount NUMERIC(15,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    paid_at TIMESTAMPTZ,
    penalty_applied BOOLEAN DEFAULT FALSE
);