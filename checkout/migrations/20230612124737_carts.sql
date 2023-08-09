-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS carts (
    user_id BIGSERIAL,
    sku BIGINT NOT NULL,
    count INT NOT NULL,
    CONSTRAINT cart_id PRIMARY KEY (user_id,sku)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS carts;
-- +goose StatementEnd
