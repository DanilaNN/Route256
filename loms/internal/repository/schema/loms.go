package schema

type Order struct {
	UserId int64  `db:"user_id"`
	Sku    uint32 `db:"sku"`
	Count  uint16 `db:"count"`
	Status string `db:"order_status"`
}

type Orders struct {
	Orders []Order
}
