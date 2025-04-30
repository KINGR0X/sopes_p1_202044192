package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "grpc-client/proto"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	kafkaAddr  = flag.String("kafka-addr", "grpc-server-service-kafka:50051", "the address to connect to Kafka gRPC server")
	rabbitAddr = flag.String("rabbit-addr", "grpc-server-service:50051", "the address to connect to RabbitMQ gRPC server")
	kafkaConn  *grpc.ClientConn
	rabbitConn *grpc.ClientConn
)

type Tweet struct {
	Descripcion string `json:"descripcion"`
	Country     string `json:"country"`
	Weather     int    `json:"weather"`
}

func initGRPC() error {
	var err error
	
	// Initialize Kafka connection
	kafkaConn, err = grpc.Dial(*kafkaAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka gRPC server: %v", err)
	}
	
	// Initialize RabbitMQ connection
	rabbitConn, err = grpc.Dial(*rabbitAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ gRPC server: %v", err)
	}
	
	return nil
}

func sendData(fiberCtx *fiber.Ctx) error {
	var body Tweet
	if err := fiberCtx.BodyParser(&body); err != nil {
		return fiberCtx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	fmt.Println("Received Tweet data:", body)

	// Convert weather
	var weather pb.Weather
	switch body.Weather {
	case 0:
		weather = pb.Weather_rainy
	case 1:
		weather = pb.Weather_cloudy
	case 2:
		weather = pb.Weather_sunny
	default:
		weather = pb.Weather_rainy
	}

	// Create request
	request := &pb.TweetRequest{
		Description: body.Descripcion,
		Country:     body.Country,
		Weather:     weather,
	}

	// Configurar contexto con timeout más generoso
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Send to both servers
	var responses []*pb.TweetResponse
	var errors []error

	// Send to Kafka server
	kafkaClient := pb.NewTweetClient(kafkaConn)
	kafkaResp, kafkaErr := kafkaClient.SendTweet(ctx, request)
	responses = append(responses, kafkaResp)
	errors = append(errors, kafkaErr)

	// Send to RabbitMQ server
	rabbitClient := pb.NewTweetClient(rabbitConn)
	rabbitResp, rabbitErr := rabbitClient.SendTweet(ctx, request)
	responses = append(responses, rabbitResp)
	errors = append(errors, rabbitErr)

	// Check for errors
	errorMessages := make([]string, 0)
	for _, err := range errors {
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
		}
	}

	if len(errorMessages) > 0 {
		return fiberCtx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"errors":  errorMessages,
			"message": "Some gRPC calls failed",
		})
	}

	return fiberCtx.JSON(fiber.Map{
		"success":  true,
		"message":  "Data sent to both servers successfully",
		"responses": responses,
	})
}

func main() {
	flag.Parse()

	// Initialize gRPC connections
	if err := initGRPC(); err != nil {
		log.Fatalf("failed to initialize gRPC connections: %v", err)
	}
	defer func() {
		kafkaConn.Close()
		rabbitConn.Close()
	}()

	app := fiber.New()
	app.Post("/grpc-go", sendData)

	fmt.Println("HTTP server listening on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}