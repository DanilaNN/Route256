package schema

type CartItem struct {
	UserID int64  `db:"user_id"`
	Sku    uint32 `db:"sku"`
	Count  uint32 `db:"count"`
}
