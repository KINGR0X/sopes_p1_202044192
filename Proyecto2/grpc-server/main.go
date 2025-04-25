package main

import (
	"context"
	"flag"
	"fmt"
	"grpc-server/kafka"
	"grpc-server/structs"
	"log"
	"net"

	pb "grpc-server/proto"

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


	// === Enviar datos a Kafka === 
	// AUN SIN TESTEAR
	go kafka.SendData(structs.Tweet{
	    Description: in.GetDescription(),
	    Country:     in.GetCountry(),
	    Weather:     int(in.GetWeather()),
	}, "tweets-topic")

	return &pb.TweetResponse{
		Success: true,
	}, nil
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTweetServer(s, &server{})
	log.Printf("Server started on port %d", *port)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}