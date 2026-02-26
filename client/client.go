package main

import (
	"context"
	"io"
	"log"
	"time"

	"weather-service/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func main() {
	// 1. Connect to the server
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewWeatherServiceClient(conn)

	// --- Execute Unary Call ---
	log.Println("--- Calling Unary GetCurrentWeather ---")
	callUnary(client)

	// --- Execute Streaming Call ---
	log.Println("\n--- Calling Server Stream WeatherUpdates ---")
	callStream(client)

	log.Println("--- Calling UpdateWeather for London ---")
	callUpdate(client)
}

func callUnary(c pb.WeatherServiceClient) {
	// 1. Create a base context
	ctx := context.Background()

	// 2. Wrap it with the Secret API Key
	ctx = metadata.AppendToOutgoingContext(ctx, "api-key", "SUPER-SECRET-123")

	// 3. Add a timeout for safety
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	res, err := c.GetCurrentWeather(ctx, &pb.WeatherRequest{City: "London"})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			log.Printf("ERROR [%v]: %v", st.Code(), st.Message())
		}
		return
	}
	log.Printf("SUCCESS: Temp is %.1f", res.GetTemperature())
}

func callStream(c pb.WeatherServiceClient) {
	// No timeout here because we want to watch the stream for 10 seconds
	stream, err := c.StreamWeatherUpdates(context.Background(), &pb.WeatherRequest{City: "London"})
	if err != nil {
		log.Fatalf("Stream error: %v", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			log.Println("SERVER DONE: Stream closed.")
			break
		}
		if err != nil {
			log.Fatalf("Recv error: %v", err)
		}
		log.Printf("STREAM UPDATE: Temp is now %.1f°C", res.GetTemperature())
	}
}
func callUpdate(c pb.WeatherServiceClient) {
	// 1. Prepare the "Envelopes" (Metadata/Security)
	ctx := metadata.AppendToOutgoingContext(context.Background(), "api-key", "SUPER-SECRET-123")

	// 2. Pack the data
	req := &pb.UpdateWeatherRequest{
		City:        "London",
		Temperature: 15.0,
		Conditions:  "Cloudy with a chance of tea",
	}

	// 3. Send it!
	res, err := c.UpdateWeather(ctx, req)
	if err != nil {
		log.Fatalf("Could not update: %v", err)
	}

	// 4. Celebrate!
	log.Printf("SERVER SAYS: %s", res.GetMessage())
}
