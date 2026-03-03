package mocks

import (
	"context"
	"weather-service/pb"
)

// MockWeatherRepo is our "Actor"
type MockWeatherRepo struct {
	// We can store "fake" responses here
	GetResult *pb.WeatherResponse
	GetError  error
}

func (m *MockWeatherRepo) GetWeatherByCity(ctx context.Context, city string) (*pb.WeatherResponse, error) {
	return m.GetResult, m.GetError
}

func (m *MockWeatherRepo) SaveWeather(ctx context.Context, city string, temp float32, cond string) error {
	return nil // Just pretend it worked!
}
