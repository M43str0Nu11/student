package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"orders-demo/internal/api"
	"orders-demo/internal/cache"
	"orders-demo/internal/db"
	"orders-demo/internal/nats"
)

func main() {
	fmt.Println("🚀 Orders Demo Service starting...")

	connStr := "postgres://orders_user:StrongP@ssw0rd@localhost:5432/orders_db?sslmode=disable"

	// Подключаем БД
	pg, err := db.New(connStr)
	if err != nil {
		log.Fatalf("Ошибка подключения к Postgres: %v", err)
	}
	defer pg.Pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pg.InitSchema(ctx); err != nil {
		log.Fatalf("Ошибка инициализации схемы: %v", err)
	}

	// Кэш
	c := cache.New()

	// Загружаем все данные из БД в кэш при старте
	data, err := pg.LoadAll(ctx)
	if err == nil {
		c.LoadAll(data)
		fmt.Printf("🔁 Кэш восстановлен: %d записей\n", len(data))
	} else {
		fmt.Println("⚠️ Не удалось загрузить данные из БД:", err)
	}

	// HTTP API
	h := api.NewHandler(c)

	// Запускаем (позже добавим NATS подписчика)
	go nats.StartSubscriber(pg, c)

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	fmt.Printf("🌐 HTTP server started on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, h.Router()))
}
