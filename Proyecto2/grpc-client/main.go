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
	addr = flag.String("addr", "grpc-server-service-kafka:50051", "the address to connect to")
	grpcConn *grpc.ClientConn
)

type Tweet struct {
	Descripcion string `json:"descripcion"`
	Country     string `json:"country"`
	Weather     int    `json:"weather"`
}

func initGRPC() error {
	var err error
	grpcConn, err = grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return err
}

func sendData(fiberCtx *fiber.Ctx) error {
	var body Tweet
	if err := fiberCtx.BodyParser(&body); err != nil {
		return fiberCtx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	fmt.Println("Received Tweet data:", body)

	// Configurar contexto con timeout más generoso
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convertir weather
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

	// Enviar mensaje
	c := pb.NewTweetClient(grpcConn)
	response, err := c.SendTweet(ctx, &pb.TweetRequest{
		Description: body.Descripcion,
		Country:     body.Country,
		Weather:     weather,
	})

	if err != nil {
		return fiberCtx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gRPC error: %v", err),
		})
	}

	return fiberCtx.JSON(fiber.Map{
		"success":  true,
		"message":  "Data sent to server successfully",
		"response": response,
	})
}

func main() {
	flag.Parse()

	// Inicializar conexión gRPC
	if err := initGRPC(); err != nil {
		log.Fatalf("failed to connect to gRPC server: %v", err)
	}
	defer grpcConn.Close()

	app := fiber.New()
	app.Post("/grpc-go", sendData)

	fmt.Println("HTTP server listening on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}