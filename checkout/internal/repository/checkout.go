package postgres

import (
	"context"
	"route256/checkout/internal/domain/models"
	"route256/checkout/internal/repository/postgress/tx"
	"route256/checkout/internal/repository/schema"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/opentracing/opentracing-go"
)

// TODO sq.StatementBuilder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar) im main
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type PGRepository struct {
	provider tx.DBProvider
}

func New(provider tx.DBProvider) *PGRepository {
	return &PGRepository{provider: provider}
}

// var (
// 	cartsAllColumns = []string{"user_id", "sku", "count"}
// )

const (
	tableNameCarts = "carts"
)

func (r *PGRepository) AddCart(ctx context.Context, userOrder models.UserOrderItem) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository/AddCart")
	defer span.Finish()

	db := r.provider.GetDB(ctx)

	query := `
INSERT INTO carts("user_id", "sku", "count") VALUES 
	($1, $2, $3)
ON CONFLICT ("user_id", "sku") DO UPDATE 
	SET count=carts.count+excluded.count;
`

	_, err := db.Exec(ctx, query, userOrder.User, userOrder.Item.Sku, userOrder.Item.Count)
	return err
}

func (r *PGRepository) GetSkuCountInCart(ctx context.Context, userOrder models.UserOrderItem) (uint32, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository/GetSkuCountInCart")
	defer span.Finish()

	db := r.provider.GetDB(ctx)

	query := psql.Select("count").
		From(tableNameCarts).
		Where(sq.Eq{"user_id": userOrder.User, "sku": userOrder.Item.Sku})

	rawSQL, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	var count uint32
	err = db.QueryRow(ctx, rawSQL, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *PGRepository) DecreaseSkuCountInCart(ctx context.Context, userOrder models.UserOrderItem, delta int32) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository/DecreaseSkuCountInCart")
	defer span.Finish()

	db := r.provider.GetDB(ctx)

	query := psql.Update(tableNameCarts).
		Set("count", sq.Expr("count - ?", delta)).
		Where(sq.Eq{"user_id": userOrder.User, "sku": userOrder.Item.Sku})

	rawSQL, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, rawSQL, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *PGRepository) DeleteUserSku(ctx context.Context, userOrder models.UserOrderItem) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository/DeleteUserSku")
	defer span.Finish()

	db := r.provider.GetDB(ctx)

	query := psql.Delete(tableNameCarts).
		Where(sq.Eq{"user_id": userOrder.User, "sku": userOrder.Item.Sku})

	rawSQL, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, rawSQL, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *PGRepository) GetCart(ctx context.Context, userID int64) (models.ProductCarts, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository/GetCart")
	defer span.Finish()

	db := r.provider.GetDB(ctx)

	query := psql.Select("sku", "count").
		From(tableNameCarts).Where(sq.Eq{"user_id": userID})

	rawSQL, args, err := query.ToSql()
	if err != nil {
		return models.ProductCarts{}, err
	}

	var items []schema.CartItem

	err = pgxscan.Select(ctx, db, &items, rawSQL, args...)
	if err != nil {
		return models.ProductCarts{}, err
	}
	var res models.ProductCarts
	for _, item := range items {
		res.Carts = append(res.Carts,
			models.ProductCart{
				Sku:   uint64(item.Sku),
				Count: uint16(item.Count)})
	}

	return res, nil
}

func (r *PGRepository) DeleteCart(ctx context.Context, userID int64) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository/DeleteCart")
	defer span.Finish()

	db := r.provider.GetDB(ctx)

	query := psql.Delete(tableNameCarts).Where(sq.Eq{"user_id": userID})

	rawSQL, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, rawSQL, args...)
	if err != nil {
		return err
	}

	return nil
}
