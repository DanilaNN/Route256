package models

type OrderItem struct {
	Sku   uint64
	Count uint16
}

type Order struct {
	OrderId uint64
	Status  string
	UserId  int64
	Items   []OrderItem
}

type Stock struct {
	WarehouseID int64
	Count       uint64
}
