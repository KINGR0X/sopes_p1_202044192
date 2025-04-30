package main

import (
	"consumer/structs"
	"consumer/valkey"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	maxWorkers        = 10               // Limitar el número de workers concurrentes
	reconnectDelay    = 5 * time.Second  // Tiempo de espera para reconexión
	processingTimeout = 30 * time.Second // Tiempo máximo para procesar un mensaje
)

func main() {
	// Configurar el número máximo de CPUs
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Canal para workers con límite
	workers := make(chan struct{}, maxWorkers)

	for {
		conn, err := connectToRabbitMQ()
		if err != nil {
			log.Printf("Failed to connect to RabbitMQ, retrying in %v: %v", reconnectDelay, err)
			time.Sleep(reconnectDelay)
			continue
		}

		ch, q, err := setupChannelAndQueue(conn)
		if err != nil {
			conn.Close()
			time.Sleep(reconnectDelay)
			continue
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
			log.Printf("Failed to register consumer: %v", err)
			ch.Close()
			conn.Close()
			time.Sleep(reconnectDelay)
			continue
		}

		log.Printf("Successfully connected to RabbitMQ. Waiting for messages...")

		for d := range msgs {
			workers <- struct{}{} // Adquirir un worker
			go func(delivery amqp.Delivery) {
				defer func() { <-workers }() // Liberar el worker

				processMessage(delivery)
			}(d)
		}

		// Si llegamos aquí, el canal de mensajes se cerró
		log.Printf("Message channel closed, reconnecting...")
		ch.Close()
		conn.Close()
	}
}

func connectToRabbitMQ() (*amqp.Connection, error) {
	return amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
}

func setupChannelAndQueue(conn *amqp.Connection) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	// Configurar QoS para prefetch
	err = ch.Qos(
		10,    // prefetch count (número de mensajes a prefetchear)
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		ch.Close()
		return nil, amqp.Queue{}, err
	}

	q, err := ch.QueueDeclare(
		"tweets-queue", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		ch.Close()
		return nil, amqp.Queue{}, err
	}

	return ch, q, nil
}

func processMessage(d amqp.Delivery) {
    ctx, cancel := context.WithTimeout(context.Background(), processingTimeout)
    defer cancel()

    // Procesar el mensaje
    log.Printf("Received a message: %s", d.Body)

    // Enviar a Valkey con el contexto
    if err := valkeyInsert(ctx, d.Body); err != nil {
        log.Printf("Failed to insert to Valkey: %v", err)
        // Puedes decidir si reintentar o no aquí
        return
    }

    // Acusar recibo del mensaje
    if err := d.Ack(false); err != nil {
        log.Printf("Error acknowledging message: %v", err)
    }
}

func valkeyInsert(ctx context.Context, data []byte) error {
    var jsonData structs.Tweet

    if err := json.Unmarshal(data, &jsonData); err != nil {
        return fmt.Errorf("failed to unmarshal message: %w", err)
    }

    return valkey.InsertWithContext(ctx, jsonData)
}