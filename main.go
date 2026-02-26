package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time" // Added this so time.Sleep works!
	"weather-service/internal/db"
	"weather-service/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedWeatherServiceServer
	repo *db.WeatherRepository
}

// The Interceptor Function
func apiKeyInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 1. Pull the "envelope" (metadata) out of the context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Metadata is missing")
	}

	// 2. Look for our specific key
	values := md["api-key"]
	if len(values) == 0 || values[0] != "SUPER-SECRET-123" {
		return nil, status.Error(codes.Unauthenticated, "Invalid or missing API key")
	}

	// 3. If the key is correct, call the actual function (handler)
	return handler(ctx, req)
}

// Unary RPC (Single request -> Single response)
func (s *server) GetCurrentWeather(ctx context.Context, req *pb.WeatherRequest) (*pb.WeatherResponse, error) {
	city := req.GetCity()
	// Use the repository instead of hardcoded values!
	weather, err := s.repo.GetWeatherByCity(ctx, city)
	if err != nil {
		// If DB can't find it, return a proper gRPC error
		return nil, status.Errorf(codes.NotFound, "City %s not found in database", city)
	}

	return weather, nil

	// // 1. Check for empty input
	// if city == "" {
	// 	log.Printf("Error: Received empty city request")
	// 	return nil, status.Error(codes.InvalidArgument, "City name cannot be empty")
	// }

	// // 2. Check for unsupported cities
	// if city == "Mars" {
	// 	log.Printf("Error: Mars is not supported")
	// 	return nil, status.Errorf(codes.NotFound, "Weather service not available for planet: %s", city)
	// }

	// log.Printf("Success: Sending weather for %s", city)
	// return &pb.WeatherResponse{
	// 	Temperature: 22.5,
	// 	Conditions:  "Sunny",
	// }, nil
}

// Server Streaming RPC (Single request -> Multiple responses)
func (s *server) StreamWeatherUpdates(req *pb.WeatherRequest, stream pb.WeatherService_StreamWeatherUpdatesServer) error {
	city := req.GetCity()
	log.Printf("Starting stream for city: %s", city)

	for i := 0; i < 5; i++ {
		update := &pb.WeatherResponse{
			Temperature: 20.0 + float32(i),
			Conditions:  "Monitoring live updates...",
		}

		if err := stream.Send(update); err != nil {
			log.Printf("Error sending to stream: %v", err)
			return err
		}

		log.Printf("Sent update %d for %s", i+1, city)
		time.Sleep(2 * time.Second)
	}

	log.Printf("Stream for %s completed successfully.", city)
	return nil
}
func (s *server) UpdateWeather(ctx context.Context, req *pb.UpdateWeatherRequest) (*pb.UpdateResponse, error) {
	// 1. Validation (Always validate your writes!)
	if req.GetCity() == "" {
		return nil, status.Error(codes.InvalidArgument, "City is required for update")
	}

	// 2. Call the Repository
	err := s.repo.SaveWeather(ctx, req.GetCity(), req.GetTemperature(), req.GetConditions())
	if err != nil {
		log.Printf("DB Error: %v", err)
		return nil, status.Error(codes.Internal, "Failed to save weather data")
	}

	return &pb.UpdateResponse{
		Message: fmt.Sprintf("Successfully updated weather for %s", req.GetCity()),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// 1. Connect to DB
	pool := db.ConnectDB()
	defer pool.Close()

	// 2. Initialize Repository
	weatherRepo := db.NewWeatherRepository(pool)

	// 3. Initialize Server with the Repo
	s := grpc.NewServer(
		grpc.UnaryInterceptor(apiKeyInterceptor),
	)

	pb.RegisterWeatherServiceServer(s, &server{repo: weatherRepo})
	reflection.Register(s)

	log.Printf("Server with Auth Shield running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
