package orderRepoPostgres

import (
	"context"
	"wbL0/internal/models"
)

func (r *OrderPostgresRepository) GetOrderInfoByUid(ctx context.Context, orderUID string) (*models.Order, error) {
	const op = "OrderPostgresRepository.GetOrderInfo"
	order := &models.Order{}

	query := `SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, 
       	delivery_service, shardkey, sm_id, date_created, oof_shard 
		FROM orders WHERE order_uid = $1`
	err := r.pool.QueryRow(ctx, query, orderUID).Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
	if err != nil {
		r.log.Error("failed to get order info", "op", op, "orderUID", orderUID, "err", err)
		return nil, err
	}
	r.log.Info("order info retrieved", "op", op, "orderUID", orderUID)
	return order, nil
}

func (r *OrderPostgresRepository) GetAllFullOrders(ctx context.Context) ([]*models.FullOrder, error) {
	const op = "OrderPostgresRepository.GetAllFullOrders"
	var fullOrders []*models.FullOrder

	queryOrders := `SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, 
                           delivery_service, shardkey, sm_id, date_created, oof_shard 
                    FROM orders`
	rows, err := r.pool.Query(ctx, queryOrders)
	if err != nil {
		r.log.Error("failed to query orders", "op", op, "err", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		if err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
			&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
		); err != nil {
			r.log.Error("failed to scan order", "op", op, "err", err)
			continue
		}

		delivery, err := r.getDeliveryByOrderUID(ctx, order.OrderUID)
		if err != nil {
			r.log.Warn("failed to get delivery", "op", op, "orderUID", order.OrderUID, "err", err)
			continue
		}

		payment, err := r.getPaymentByOrderUID(ctx, order.OrderUID)
		if err != nil {
			r.log.Warn("failed to get payment", "op", op, "orderUID", order.OrderUID, "err", err)
			continue
		}

		items, err := r.getItemsByOrderUID(ctx, order.OrderUID)
		if err != nil {
			r.log.Warn("failed to get items", "op", op, "orderUID", order.OrderUID, "err", err)
			continue
		}

		fullOrders = append(fullOrders, &models.FullOrder{
			Order:    order,
			Delivery: *delivery,
			Payment:  *payment,
			Items:    items,
		})
	}

	r.log.Info("retrieved all full orders", "op", op, "count", len(fullOrders))
	return fullOrders, nil
}

func (r *OrderPostgresRepository) GetFullOrderByUID(ctx context.Context, orderUID string) (*models.FullOrder, error) {
	const op = "OrderPostgresRepository.GetFullOrderByUID"

	order, err := r.GetOrderInfoByUid(ctx, orderUID)
	if err != nil {
		r.log.Error("failed to get order info", "op", op, "orderUID", orderUID, "err", err)
		return nil, err
	}

	delivery, err := r.getDeliveryByOrderUID(ctx, orderUID)
	if err != nil {
		r.log.Error("failed to get delivery", "op", op, "orderUID", orderUID, "err", err)
		return nil, err
	}

	payment, err := r.getPaymentByOrderUID(ctx, orderUID)
	if err != nil {
		r.log.Error("failed to get payment", "op", op, "orderUID", orderUID, "err", err)
		return nil, err
	}

	items, err := r.getItemsByOrderUID(ctx, orderUID)
	if err != nil {
		r.log.Error("failed to get items", "op", op, "orderUID", orderUID, "err", err)
		return nil, err
	}

	r.log.Info("full order retrieved", "op", op, "orderUID", orderUID)
	return &models.FullOrder{
		Order:    *order,
		Delivery: *delivery,
		Payment:  *payment,
		Items:    items,
	}, nil
}

func (r *OrderPostgresRepository) getDeliveryByOrderUID(ctx context.Context, orderUID string) (*models.Delivery, error) {
	const op = "OrderPostgresRepository.getDeliveryByOrderUID"
	var delivery models.Delivery
	query := `SELECT order_uid, name, phone, zip, city, address, region, email 
              FROM delivery WHERE order_uid = $1`
	err := r.pool.QueryRow(ctx, query, orderUID).Scan(
		&delivery.OrderUID, &delivery.Name, &delivery.Phone, &delivery.Zip,
		&delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
	)
	if err != nil {
		r.log.Error("failed to get delivery", "op", op, "orderUID", orderUID, "err", err)
		return nil, err
	}
	return &delivery, nil
}

func (r *OrderPostgresRepository) getPaymentByOrderUID(ctx context.Context, orderUID string) (*models.Payment, error) {
	const op = "OrderPostgresRepository.getPaymentByOrderUID"
	var payment models.Payment
	query := `SELECT order_uid, transaction, request_id, currency, provider, amount, 
                    payment_dt, bank, delivery_cost, goods_total, custom_fee 
              FROM payment WHERE order_uid = $1`
	err := r.pool.QueryRow(ctx, query, orderUID).Scan(
		&payment.OrderUID, &payment.Transaction, &payment.RequestID, &payment.Currency,
		&payment.Provider, &payment.Amount, &payment.PaymentDt, &payment.Bank,
		&payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
	)
	if err != nil {
		r.log.Error("failed to get payment", "op", op, "orderUID", orderUID, "err", err)
		return nil, err
	}
	return &payment, nil
}

func (r *OrderPostgresRepository) getItemsByOrderUID(ctx context.Context, orderUID string) ([]models.Item, error) {
	const op = "OrderPostgresRepository.getItemsByOrderUID"
	var items []models.Item
	query := `SELECT order_uid, chrt_id, track_number, price, rid, name, sale, size, 
                    total_price, nm_id, brand, status 
              FROM items WHERE order_uid = $1`
	rows, err := r.pool.Query(ctx, query, orderUID)
	if err != nil {
		r.log.Error("failed to query items", "op", op, "orderUID", orderUID, "err", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		if err := rows.Scan(
			&item.OrderUID, &item.ChrtID, &item.TrackNumber, &item.Price,
			&item.Rid, &item.Name, &item.Sale, &item.Size, &item.TotalPrice,
			&item.NmID, &item.Brand, &item.Status,
		); err != nil {
			r.log.Warn("failed to scan item", "op", op, "orderUID", orderUID, "err", err)
			continue
		}
		items = append(items, item)
	}
	r.log.Info("items retrieved", "op", op, "orderUID", orderUID, "count", len(items))
	return items, nil
}
