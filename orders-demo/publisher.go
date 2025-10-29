package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/nats-io/stan.go"
)

func main() {
	clusterID := "test-cluster"    // должно совпадать с подписчиком
	clientID := "order-publisher"  // уникальное имя
	channel := "orders"            // канал подписки
	url := "nats://localhost:4222" // адрес NATS Streaming

	// Подключаемся
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(url))
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	defer sc.Close()

	// Читаем JSON из файла
	data, err := ioutil.ReadFile("B:/orders-demo/model.json")
	if err != nil {
		log.Fatalf("Не удалось прочитать файл: %v", err)
	}

	// Публикуем в канал
	err = sc.Publish(channel, data)
	if err != nil {
		log.Fatalf("Ошибка публикации: %v", err)
	}

	fmt.Println("✅ JSON успешно опубликован в канал", channel)
}
