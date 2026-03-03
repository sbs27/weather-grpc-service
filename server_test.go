package main

import (
	"context"
	"fmt"
	"testing"
	"weather-service/internal/db/mocks"
	"weather-service/pb"

	"github.com/stretchr/testify/assert" // We need this library!
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetCurrentWeather_Success(t *testing.T) {
	// 1. SETUP: Create the Fake Repo
	mockRepo := &mocks.MockWeatherRepo{
		GetResult: &pb.WeatherResponse{
			Temperature: 25.0,
			Conditions:  "Sunny Mock",
		},
	}

	// 2. SETUP: Create the Server with the Fake Repo
	s := &server{repo: mockRepo}

	// 3. EXECUTE: Call the function we want to test
	req := &pb.WeatherRequest{City: "London"}
	res, err := s.GetCurrentWeather(context.Background(), req)

	// 4. ASSERT: Did we get what we expected?
	assert.NoError(t, err)
	assert.Equal(t, float32(25.0), res.Temperature)
	assert.Equal(t, "Sunny Mock", res.Conditions)
}

func TestGetCurrentWeather_NotFound(t *testing.T) {
	// 1. SETUP: Create a Mock that returns an ERROR
	mockRepo := &mocks.MockWeatherRepo{
		GetResult: nil,
		GetError:  fmt.Errorf("city not found in DB"),
	}

	s := &server{repo: mockRepo}

	// 2. EXECUTE
	req := &pb.WeatherRequest{City: "UnknownCity"}
	res, err := s.GetCurrentWeather(context.Background(), req)

	// 3. ASSERT
	assert.Nil(t, res)   // Response should be empty
	assert.Error(t, err) // There should be an error

	// 4. SENIOR CHECK: Verify it's a gRPC "NotFound" status code
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}
