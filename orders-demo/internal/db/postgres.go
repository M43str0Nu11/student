package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"orders-demo/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func New(connString string) (*Postgres, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	return &Postgres{Pool: pool}, nil
}

// InitSchema создаёт таблицу, если она отсутствует
func (p *Postgres) InitSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS orders (
		order_uid TEXT PRIMARY KEY,
		payload JSONB NOT NULL,
		created_at TIMESTAMPTZ DEFAULT now()
	);`
	_, err := p.Pool.Exec(ctx, schema)
	return err
}

// SaveOrder сохраняет заказ в БД
func (p *Postgres) SaveOrder(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	_, err = p.Pool.Exec(ctx,
		`INSERT INTO orders (order_uid, payload) VALUES ($1, $2)
		 ON CONFLICT (order_uid) DO UPDATE SET payload = EXCLUDED.payload`,
		order.OrderUID, data)
	return err
}

// LoadAll загружает все заказы при старте (для восстановления кэша)
func (p *Postgres) LoadAll(ctx context.Context) (map[string]model.CachedOrder, error) {
	rows, err := p.Pool.Query(ctx, `SELECT order_uid, payload FROM orders`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cache := make(map[string]model.CachedOrder)
	for rows.Next() {
		var id string
		var payload []byte
		if err := rows.Scan(&id, &payload); err != nil {
			return nil, err
		}
		var order model.Order
		if err := json.Unmarshal(payload, &order); err != nil {
			fmt.Printf("⚠️ Ошибка при Unmarshal: %v\n", err)
			continue
		}
		cache[id] = model.CachedOrder{Order: order, PayloadRaw: payload, LoadedAt: time.Now()}
	}
	return cache, nil
}
