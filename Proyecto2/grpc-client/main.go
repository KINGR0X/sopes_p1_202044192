package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	pb "grpc-client/proto"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// conexión con el servidor gRPC
	addr = flag.String("addr", "grpc_server_service_kafka:50051", "the address to connect to")
)

type Tweet struct {
	Descripcion string `json:"descripcion"`
	Country     string `json:"country"`
	Weather     int    `json:"weather"`
}

func sendData(fiberCtx *fiber.Ctx) error {
	// parse del Json recibido 
	var body Tweet

	if err := fiberCtx.BodyParser(&body); err != nil {
		return fiberCtx.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	fmt.Println("Received Tweet data:", body)

	// Conexión al servidor gRPC
	address := *addr

	// Enviar datos al servidor gRPC
	var wg sync.WaitGroup
	wg.Add(1)
	
	var response *pb.TweetResponse
	var grpcErr error
	
	go func() {
		defer wg.Done()

		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			grpcErr = fmt.Errorf("did not connect to %s: %v", address, err)
			return
		}
		defer conn.Close()

		c := pb.NewTweetClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		var weather pb.Weather
		// cambiar el valor de weather a un valor de la enumeración Weather
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

		// Enviar el mensaje al servidor gRPC
		r, err := c.SendTweet(ctx, &pb.TweetRequest{ 
			Description: body.Descripcion,
			Country:     body.Country,
			Weather:     weather,
		})

		// Si hay un error
		if err != nil {
			grpcErr = fmt.Errorf("error from %s: %v", address, err)
			return
		}

		response = r
		fmt.Printf("Received response from gRPC server %s: %v\n", address, r)
	}()

	wg.Wait()

	if grpcErr != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": grpcErr.Error(),
		})
	}

	// ver si el mensaje llego bien
	return fiberCtx.JSON(fiber.Map{
		"success":  true,
		"message":  "Data sent to server successfully",
		"response": response, // Ahora se usa la variable response
	})
}

func main() {
	flag.Parse()
	app := fiber.New()
	app.Post("/weather", sendData)

	fmt.Println("HTTP server listening on :8080")
	err := app.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
}