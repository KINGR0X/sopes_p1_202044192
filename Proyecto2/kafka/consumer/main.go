package main

import (
	"consumer/redis"
	"consumer/structs"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func main() {

	topic := "tweets-topic"
	broker := "my-cluster-kafka-bootstrap:9092"

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		Topic:       topic,
		Partition:   0,
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.LastOffset,
		GroupID:     uuid.New().String(),
	})

	for {
		// Leer el mensaje
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println("Error al leer el mensake", err)
			break
		}

		// Imprimir el mensaje
		fmt.Printf("mensaje %d: %s= %s\n", m.Offset, string(m.Key), string(m.Value))

		//mandando a redis
		fmt.Println("mandando a redis")
		redistInsert(m.Value)

		// Commit el mensaje
		err = r.CommitMessages(context.Background(), m)
		if err != nil {
			log.Println("Error al commit", err)
		}
	}

	if err := r.Close(); err != nil {
		log.Fatal("Error al cerrar el lector")
	}
}

func redistInsert(data []byte) {

	var jsonData structs.Tweet

	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		fmt.Printf("Failed to unmarshal message: %s", err)
		return
	}

	// go rutina para insertar en Redis
	go redis.Insert(jsonData)
}