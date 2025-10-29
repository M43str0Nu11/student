package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"orders-demo/internal/cache"
	"orders-demo/internal/db"
	"orders-demo/internal/model"

	"github.com/nats-io/stan.go"
)

// StartSubscriber запускает подписчика NATS Streaming, который принимает JSON-заказы
func StartSubscriber(pg *db.Postgres, c *cache.Cache) {
	clusterID := "test-cluster" // имя кластера (должно совпадать с тем, что указано при запуске контейнера NATS)
	clientID := "order-subscriber"
	channel := "orders" // название канала подписки
	url := "nats://localhost:4222"

	// Подключение к NATS Streaming
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(url))
	if err != nil {
		log.Fatalf("Ошибка подключения к NATS Streaming: %v", err)
	}
	defer sc.Close()

	fmt.Println("📡 Подписчик подключён к NATS Streaming...")

	// Подписка на канал "orders"
	_, err = sc.Subscribe(channel, func(msg *stan.Msg) {
		fmt.Printf("📨 Получено сообщение: %s\n", msg.Data)

		var order model.Order
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			fmt.Println("⚠️ Ошибка парсинга JSON:", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Сохраняем заказ в БД
		if err := pg.SaveOrder(ctx, order); err != nil {
			fmt.Println("❌ Ошибка сохранения в БД:", err)
			return
		}

		// Добавляем в кэш
		c.Set(order.OrderUID, model.CachedOrder{
			Order:      order,
			PayloadRaw: msg.Data,
			LoadedAt:   time.Now(),
		})

		fmt.Printf("💾 Заказ %s сохранён и добавлен в кэш\n", order.OrderUID)
	}, stan.DeliverAllAvailable())

	if err != nil {
		log.Fatalf("Ошибка подписки: %v", err)
	}

	// Чтобы сервис не завершался
	select {}
}
