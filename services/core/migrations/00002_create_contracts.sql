-- +goose Up
CREATE TABLE contracts (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    counterparty text        NOT NULL CHECK (char_length(counterparty) BETWEEN 1 AND 200),
    amount_cents bigint      NOT NULL CHECK (amount_cents > 0),
    currency     text        NOT NULL CHECK (currency IN ('UAH', 'EUR', 'USD')),
    signed_at    date        NOT NULL,
    ends_at      date        CONSTRAINT contracts_ends_at_check CHECK (ends_at >= signed_at),
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER contracts_set_updated_at BEFORE UPDATE ON contracts
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE contracts;
