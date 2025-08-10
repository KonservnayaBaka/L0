package orderRepoPostgres

import (
	"context"
	"wbL0/internal/models"
)

func (r *OrderPostgresRepository) SaveOrderDataTx(ctx context.Context, tx PgxTx, order *models.Order) error {
	const op = "OrderPostgresRepository.SaveOrderDataTx"

	query := `INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := tx.Exec(ctx, query,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	if err != nil {
		r.log.Error("failed to save order data", "op", op, "orderUID", order.OrderUID, "err", err)
		return err
	}
	r.log.Info("order data saved", "op", op, "orderUID", order.OrderUID)
	return nil
}

func (r *OrderPostgresRepository) SaveDeliveryDataTx(ctx context.Context, tx PgxTx, delivery *models.Delivery) error {
	const op = "OrderPostgresRepository.SaveDeliveryDataTx"

	query := `INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := tx.Exec(ctx, query,
		delivery.OrderUID,
		delivery.Name,
		delivery.Phone,
		delivery.Zip,
		delivery.City,
		delivery.Address,
		delivery.Region,
		delivery.Email,
	)
	if err != nil {
		r.log.Error("failed to save delivery data", "op", op, "orderUID", delivery.OrderUID, "err", err)
		return err
	}
	r.log.Info("delivery data saved", "op", op, "orderUID", delivery.OrderUID)
	return nil
}

func (r *OrderPostgresRepository) SavePaymentDataTx(ctx context.Context, tx PgxTx, payment *models.Payment) error {
	const op = "OrderPostgresRepository.SavePaymentDataTx"

	query := `INSERT INTO payment (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := tx.Exec(ctx, query,
		payment.OrderUID,
		payment.Transaction,
		payment.RequestID,
		payment.Currency,
		payment.Provider,
		payment.Amount,
		payment.PaymentDt,
		payment.Bank,
		payment.DeliveryCost,
		payment.GoodsTotal,
		payment.CustomFee,
	)
	if err != nil {
		r.log.Error("failed to save payment data", "op", op, "orderUID", payment.OrderUID, "err", err)
		return err
	}
	r.log.Info("payment data saved", "op", op, "orderUID", payment.OrderUID)
	return nil
}

func (r *OrderPostgresRepository) SaveItemsDataTx(ctx context.Context, tx PgxTx, item *models.Item) error {
	const op = "OrderPostgresRepository.SaveItemsDataTx"

	query := `INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	_, err := tx.Exec(ctx, query,
		item.OrderUID,
		item.ChrtID,
		item.TrackNumber,
		item.Price,
		item.Rid,
		item.Name,
		item.Sale,
		item.Size,
		item.TotalPrice,
		item.NmID,
		item.Brand,
		item.Status,
	)
	if err != nil {
		r.log.Error("failed to save item data", "op", op, "orderUID", item.OrderUID, "err", err)
		return err
	}
	r.log.Info("item data saved", "op", op, "orderUID", item.OrderUID)
	return nil
}
