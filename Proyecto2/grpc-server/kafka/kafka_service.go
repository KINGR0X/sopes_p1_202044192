package kafka

import (
	"context"
	"encoding/json"
	"grpc-server/structs"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func SendData(data structs.Tweet, topic string) {
	partition := 0

	conn, err := kafka.DialLeader(context.Background(), "tcp", "my-cluster-kafka-bootstrap:9092", topic, partition)

	if err != nil {
		log.Printf("Error al conectar con Kafka: %v", err)
	}

	// Convertir el struct a JSON
	valueBytes, err := json.Marshal(data)
	
	if err != nil {
		log.Printf("Error al serializar datos a JSON: %v", err)
		
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	// Enviar el mensaje a Kafka
	_, err = conn.WriteMessages(
		kafka.Message{Value: valueBytes},
	)

	if err != nil {
		log.Fatal("Error al mandar el mensaje:", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("Error al cerrar la conexion:", err)
	}

	log.Println("Mensaje enviado")
}