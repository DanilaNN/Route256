package repository

import (
	"context"
	"route256/loms/internal/domain/models"
	"route256/loms/internal/repository/schema"

	"github.com/Masterminds/squirrel"
	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

var (
	OrdersAllColumns     = []string{"order_id", "user_id", "order_status"}
	OrderItemsAllColumns = []string{"order_id", "sku", "count"}
)

const (
	tableNameUserOrders = "orders"
	tableNameOrderItems = "order_items"
)

// TODO sq.StatementBuilder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar) im main
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (r *Repository) InsertOrderItems(ctx context.Context, order models.Order) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository/InsertOrderItems")
	defer span.Finish()

	query := psql.Insert(tableNameOrderItems).Columns(OrderItemsAllColumns...)

	for _, item := range order.Items {
		query = query.Values(order.OrderId, item.Sku, item.Count)
	}

	rawSQL, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Query(ctx, rawSQL, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) CreateOrder(ctx context.Context, order models.Order) (uint64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository/CreateOrder")
	defer span.Finish()

	query := psql.Insert(tableNameUserOrders).Columns("user_id").
		Values(order.UserId).
		Suffix("RETURNING order_id")

	rawSQL, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	var orderId uint64
	err = r.pool.QueryRow(ctx, rawSQL, args...).Scan(&orderId)
	if err != nil {
		return 0, err
	}
	order.OrderId = orderId

	err = r.InsertOrderItems(ctx, order)
	if err != nil {
		return 0, err
	}

	return orderId, nil
}

func (r *Repository) GetOrder(ctx context.Context, orderID int64) (models.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository/GetOrder")
	defer span.Finish()

	query := psql.Select("user_id, sku, count, order_status").From(tableNameUserOrders).
		Join(tableNameOrderItems + " USING (order_id)").
		Where(squirrel.Eq{"orders.order_id": orderID})

	rawSQL, args, err := query.ToSql()
	if err != nil {
		return models.Order{}, err
	}

	orders := make([]schema.Order, 0)
	err = pgxscan.Select(ctx, r.pool, &orders, rawSQL, args...)
	if err != nil {
		return models.Order{}, err
	}

	var res models.Order
	if len(orders) != 0 {
		res.UserId = orders[0].UserId
		res.Status = orders[0].Status
		res.OrderId = uint64(orderID)
	}
	for _, order := range orders {
		res.Items = append(res.Items, models.OrderItem{Sku: uint64(order.Sku), Count: order.Count})
	}

	return res, nil
}

func (r *Repository) SetOrderStatus(ctx context.Context, orderID int64, status string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository/SetOrderStatus")
	defer span.Finish()

	query := psql.Update(tableNameUserOrders).Set("order_status", status).
		Where(squirrel.Eq{"order_id": orderID})

	rawSQL, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, rawSQL, args...)
	if err != nil {
		return err
	}

	return nil
}
