package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"grpc_server_rabbit/structs"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
    conn *amqp.Connection
    ch   *amqp.Channel
    mu   sync.Mutex
)

// Initialize establece la conexión con RabbitMQ
func Initialize(url string) error {
    mu.Lock()
    defer mu.Unlock()
    
    var err error
    conn, err = amqp.Dial(url)
    if err != nil {
        return fmt.Errorf("failed to connect: %v", err)
    }

    ch, err = conn.Channel()
    if err != nil {
        conn.Close()
        return fmt.Errorf("failed to open channel: %v", err)
    }
    
    if err := ch.Confirm(false); err != nil {
        return fmt.Errorf("failed to put channel in confirm mode: %v", err)
    }
    
    return nil
}

// SendData envía datos a una cola de RabbitMQ
func SendData(data structs.Tweet, queueName string) error {
    mu.Lock()
    defer mu.Unlock()
    
    // Verificar y reconectar si es necesario
    if conn == nil || conn.IsClosed() {
        if err := Initialize("amqp://guest:guest@rabbitmq:5672/"); err != nil {
            return fmt.Errorf("reconnect failed: %v", err)
        }
    }

    // Declarar cola
    _, err := ch.QueueDeclare(
        queueName,
        true,  // durable
        false, // autoDelete
        false, // exclusive
        false, // noWait
        nil,   // args
    )
    if err != nil {
        return fmt.Errorf("queue declare failed: %v", err)
    }

    // Convertir el struct a JSON
    body, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("json marshal failed: %v", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Publicar el mensaje
    err = ch.PublishWithContext(ctx,
        "",        // exchange
        queueName, // routing key
        false,     // mandatory
        false,     // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        })
    if err != nil {
        return fmt.Errorf("publish failed: %v", err)
    }

    log.Printf("Mensaje enviado exitosamente a RabbitMQ queue: %s", queueName)
    return nil
}

// Close cierra las conexiones
func Close() {
    mu.Lock()
    defer mu.Unlock()
    
    if ch != nil {
        ch.Close()
        ch = nil
    }
    if conn != nil {
        conn.Close()
        conn = nil
    }
}