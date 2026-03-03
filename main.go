package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"weather-service/internal/db"
	"weather-service/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedWeatherServiceServer
	repo db.WeatherStore
}

func apiKeyInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Metadata is missing")
	}

	values := md["api-key"]
	if len(values) == 0 || values[0] != "SUPER-SECRET-123" {
		return nil, status.Error(codes.Unauthenticated, "Invalid or missing API key")
	}

	return handler(ctx, req)
}

func (s *server) GetCurrentWeather(ctx context.Context, req *pb.WeatherRequest) (*pb.WeatherResponse, error) {
	city := req.GetCity()
	weather, err := s.repo.GetWeatherByCity(ctx, city)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "City %s not found in database", city)
	}

	return weather, nil
}

func (s *server) StreamWeatherUpdates(req *pb.WeatherRequest, stream pb.WeatherService_StreamWeatherUpdatesServer) error {
	city := req.GetCity()
	// ✅ FIXED: Used key "city"
	slog.Info("Starting stream", "city", city)

	for i := 0; i < 5; i++ {
		update := &pb.WeatherResponse{
			Temperature: 20.0 + float32(i),
			Conditions:  "Monitoring live updates...",
		}

		if err := stream.Send(update); err != nil {
			// ✅ FIXED: Used key "error"
			slog.Error("Error sending to stream", "error", err)
			return err
		}

		// ✅ FIXED: Used keys "count" and "city"
		slog.Info("Sent update", "count", i+1, "city", city)
		time.Sleep(2 * time.Second)
	}

	// ✅ FIXED: Used key "city"
	slog.Info("Stream completed successfully", "city", city)
	return nil
}

func (s *server) UpdateWeather(ctx context.Context, req *pb.UpdateWeatherRequest) (*pb.UpdateResponse, error) {
	if req.GetCity() == "" {
		return nil, status.Error(codes.InvalidArgument, "City is required for update")
	}

	err := s.repo.SaveWeather(ctx, req.GetCity(), req.GetTemperature(), req.GetConditions())
	if err != nil {
		// ✅ FIXED: Used key "error"
		slog.Error("DB Error", "error", err)
		return nil, status.Error(codes.Internal, "Failed to save weather data")
	}

	return &pb.UpdateResponse{
		Message: fmt.Sprintf("Successfully updated weather for %s", req.GetCity()),
	}, nil
}

func loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	slog.Info("gRPC Request",
		"method", info.FullMethod,
		"duration_ms", duration.Milliseconds(),
		"error", err,
	)

	return resp, err
}

func main() {
	// Initialize logger first!
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		// ✅ FIXED: Used key "error"
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	slog.Info("Service is starting", "version", "1.0.0", "port", 50051)

	pool := db.ConnectDB()
	defer pool.Close()

	weatherRepo := db.NewWeatherRepository(pool)

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor,
			apiKeyInterceptor,
		),
	)

	pb.RegisterWeatherServiceServer(s, &server{repo: weatherRepo})
	reflection.Register(s)

	// Setup Health Checks
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	slog.Info("Server running", "addr", ":50051")
	if err := s.Serve(lis); err != nil {
		// ✅ FIXED: Used key "error"
		slog.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
