-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS order_items (
    order_id BIGINT,
    sku BIGINT,
    count INT,
    CONSTRAINT order_item_key PRIMARY KEY (order_id,sku)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS order_items;
-- +goose StatementEnd
