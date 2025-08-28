package repo

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Arkine2054/l0/internal/models"
	"github.com/lib/pq"
	"log"
	"sync"
)

type repo struct {
	db    *sql.DB
	cache map[string]*models.Order
	mu    sync.RWMutex
}

type Repo interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id string) (*models.Order, error)
	WarmUpCache(ctx context.Context) error
	Close(ctx context.Context) error
}

func NewRepo(db *sql.DB) Repo {
	return &repo{
		db:    db,
		cache: make(map[string]*models.Order),
	}
}

func (r *repo) CreateOrder(ctx context.Context, order *models.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx error: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders 
		    (order_uid, track_number, entry, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.CustomerID, order.DeliveryService, order.ShardKey,
		order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return fmt.Errorf("insert order error: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO deliveries 
		    (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("insert delivery error: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO payments 
		    (order_uid, transaction, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.Payment.Transaction, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("insert payment error: %w", err)
	}

	for _, it := range order.Items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO items 
			    (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT (order_uid, rid) DO NOTHING`,
			order.OrderUID, it.ChrtID, it.TrackNumber, it.Price, it.RID, it.Name,
			it.Sale, it.Size, it.TotalPrice, it.NmID, it.Brand, it.Status)
		if err != nil {
			return fmt.Errorf("insert item error: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx error: %w", err)
	}

	// обновляем кэш
	r.mu.Lock()
	r.cache[order.OrderUID] = order
	r.mu.Unlock()

	return nil
}
func (r *repo) WarmUpCache(ctx context.Context) error {
	cache := make(map[string]*models.Order)

	rows, err := r.db.QueryContext(ctx, `
		SELECT order_uid, track_number, entry, locale, customer_id,
		       delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders
		ORDER BY date_created DESC
		LIMIT 100`)
	if err != nil {
		return fmt.Errorf("load orders error: %w", err)
	}
	defer rows.Close()

	orderIDs := []string{}
	for rows.Next() {
		o := &models.Order{}
		if err := rows.Scan(
			&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.CustomerID,
			&o.DeliveryService, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard,
		); err != nil {
			return fmt.Errorf("scan order error: %w", err)
		}
		cache[o.OrderUID] = o
		orderIDs = append(orderIDs, o.OrderUID)
	}

	if len(orderIDs) == 0 {
		return nil
	}

	delRows, err := r.db.QueryContext(ctx, `
		SELECT order_uid, name, phone, zip, city, address, region, email
		FROM deliveries
		WHERE order_uid = ANY($1)`, pq.Array(orderIDs))
	if err != nil {
		return fmt.Errorf("load deliveries error: %w", err)
	}
	defer delRows.Close()
	for delRows.Next() {
		var d models.Delivery
		var oid string
		if err := delRows.Scan(&oid, &d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email); err != nil {
			return fmt.Errorf("scan delivery error: %w", err)
		}
		if o, ok := cache[oid]; ok {
			o.Delivery = d
		}
	}

	payRows, err := r.db.QueryContext(ctx, `
		SELECT order_uid, transaction, currency, provider, amount, payment_dt,
		       bank, delivery_cost, goods_total, custom_fee
		FROM payments
		WHERE order_uid = ANY($1)`, pq.Array(orderIDs))
	if err != nil {
		return fmt.Errorf("load payments error: %w", err)
	}
	defer payRows.Close()
	for payRows.Next() {
		var p models.Payment
		var oid string
		if err := payRows.Scan(&oid, &p.Transaction, &p.Currency, &p.Provider,
			&p.Amount, &p.PaymentDT, &p.Bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee); err != nil {
			return fmt.Errorf("scan payment error: %w", err)
		}
		if o, ok := cache[oid]; ok {
			o.Payment = p
		}
	}

	itemRows, err := r.db.QueryContext(ctx, `
		SELECT order_uid, chrt_id, track_number, price, rid, name, sale, size,
		       total_price, nm_id, brand, status
		FROM items
		WHERE order_uid = ANY($1)`, pq.Array(orderIDs))
	if err != nil {
		return fmt.Errorf("load items error: %w", err)
	}
	defer itemRows.Close()
	for itemRows.Next() {
		var it models.Item
		var oid string
		if err := itemRows.Scan(&oid, &it.ChrtID, &it.TrackNumber, &it.Price,
			&it.RID, &it.Name, &it.Sale, &it.Size, &it.TotalPrice, &it.NmID,
			&it.Brand, &it.Status); err != nil {
			return fmt.Errorf("scan item error: %w", err)
		}
		if o, ok := cache[oid]; ok {
			o.Items = append(o.Items, it)
		}
	}

	log.Printf("WarmUpCache: loaded %d orders into cache", len(cache))
	return nil
}

func (r *repo) GetByID(ctx context.Context, id string) (*models.Order, error) {

	// сначала пытаемся достать из кэша
	r.mu.RLock()
	cached, ok := r.cache[id]
	r.mu.RUnlock()

	if ok {
		log.Printf("data loaded from cache: %v", id)
		return cached, nil
	}

	// если нет в кэше → идём в БД
	log.Printf("data loaded from database: %v", id)

	var order models.Order

	err := r.db.QueryRowContext(ctx, `SELECT order_uid, track_number, entry, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
	                      FROM orders WHERE order_uid=$1`, id).
		Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.CustomerID, &order.DeliveryService, &order.ShardKey, // <--- добавлено
			&order.SmID, &order.DateCreated, &order.OofShard)
	if err != nil {
		return nil, fmt.Errorf("error getting order by id: %w", err)
	}

	err = r.db.QueryRowContext(ctx,
		`SELECT name, phone, zip, city, address, region, email FROM deliveries WHERE order_uid=$1`, id).
		Scan(&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
			&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email)
	if err != nil {
		return nil, fmt.Errorf("error getting deliveries by id: %w", err)
	}

	err = r.db.QueryRowContext(ctx, `SELECT transaction, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
	                   FROM payments WHERE order_uid=$1`, id).
		Scan(&order.Payment.Transaction, &order.Payment.Currency, &order.Payment.Provider,
			&order.Payment.Amount, &order.Payment.PaymentDT, &order.Payment.Bank,
			&order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)
	if err != nil {
		return nil, fmt.Errorf("error getting payments by id: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
	                       FROM items WHERE order_uid=$1`, id)
	if err != nil {
		return nil, fmt.Errorf("error getting items by id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var it models.Item
		err = rows.Scan(&it.ChrtID, &it.TrackNumber, &it.Price, &it.RID, &it.Name,
			&it.Sale, &it.Size, &it.TotalPrice, &it.NmID, &it.Brand, &it.Status)
		if err != nil {
			log.Println(err)
		}
		order.Items = append(order.Items, it)

	}

	r.mu.Lock()
	r.cache[id] = &order
	r.mu.Unlock()

	return &order, nil
}

func (r *repo) Close(ctx context.Context) error {
	return r.db.Close()
}
