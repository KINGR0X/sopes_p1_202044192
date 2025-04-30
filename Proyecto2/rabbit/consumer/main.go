package main

import (
	"consumer/structs"
	"consumer/valkey" // Cambiado de redis a valkey
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"tweets-queue", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			// Process the message
			log.Printf("Received a message: %s", d.Body)
			
			// Send to Valkey
			log.Println("Sending to Valkey")
			valkeyInsert(d.Body)
			
			// Acknowledge the message
			if err := d.Ack(false); err != nil {
				log.Printf("Error acknowledging message: %v", err)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func valkeyInsert(data []byte) { 
	var jsonData structs.Tweet

	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		log.Printf("Failed to unmarshal message: %s", err)
		return
	}


	go valkey.Insert(jsonData)  
}