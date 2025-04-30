package main

import (
	"context"
	"flag"
	"fmt"
	"grpc_server_rabbit/rabbitmq"
	"grpc_server_rabbit/structs"
	"log"
	"net"

	pb "grpc_server_rabbit/proto"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedTweetServer
}

func (s *server) SendTweet(_ context.Context, in *pb.TweetRequest) (*pb.TweetResponse, error) {
    log.Printf("Received Tweet: %v", in)
	log.Printf("Description: %s", in.GetDescription())
	log.Printf("Country: %s", in.GetCountry())
	log.Printf("Weather: %d", in.GetWeather())

    if err := rabbitmq.SendData(structs.Tweet{
        Description: in.GetDescription(),
        Country:     in.GetCountry(),
        Weather:     int(in.GetWeather()),
    }, "tweets-queue"); err != nil {
        log.Printf("Failed to send to RabbitMQ: %v", err)
        return &pb.TweetResponse{Success: false}, nil
    }
    
    return &pb.TweetResponse{Success: true}, nil
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Inicializar conexión a RabbitMQ
	err = rabbitmq.Initialize("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()

	s := grpc.NewServer()
	pb.RegisterTweetServer(s, &server{})
	log.Printf("Server started on port %d", *port)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}