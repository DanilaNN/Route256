package models

type Product struct {
	Name  string
	Price int32
}

type Stock struct {
	WarehouseID int64
	Count       uint64
}

type Skus []uint32

type OrderItem struct {
	Sku   uint64
	Count uint16
}

type UserOrderItem struct {
	User int64
	Item OrderItem
}
type Order struct {
	UserId int64
	Items  []OrderItem
}

type ProductCart struct {
	Sku   uint64
	Count uint16
	Name  string
	Price uint32
}

type ProductCarts struct {
	Carts      []ProductCart
	TotalPrice uint32
}
