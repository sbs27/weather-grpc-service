package db

import (
	"context"
	"weather-service/pb"

	"github.com/jackc/pgx/v5/pgxpool"
)

// WeatherRepository handles all database operations
type WeatherRepository struct {
	Pool *pgxpool.Pool
}

// NewWeatherRepository creates a new repo instance
func NewWeatherRepository(pool *pgxpool.Pool) *WeatherRepository {
	return &WeatherRepository{Pool: pool}
}

// GetWeatherByCity queries the DB for a specific city's data
func (r *WeatherRepository) GetWeatherByCity(ctx context.Context, city string) (*pb.WeatherResponse, error) {
	var temp float32
	var cond string

	// Standard SQL query
	query := "SELECT temperature, conditions FROM weather WHERE city_name = $1"

	err := r.Pool.QueryRow(ctx, query, city).Scan(&temp, &cond)
	if err != nil {
		return nil, err // This will return a 'no rows in result set' error if city isn't found
	}

	return &pb.WeatherResponse{
		Temperature: temp,
		Conditions:  cond,
	}, nil
}
func (r *WeatherRepository) SaveWeather(ctx context.Context, city string, temp float32, cond string) error {
	// THE UPSERT QUERY
	query := `
		INSERT INTO weather (city_name, temperature, conditions, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (city_name) 
		DO UPDATE SET 
			temperature = EXCLUDED.temperature,
			conditions = EXCLUDED.conditions,
			updated_at = NOW();
	`
	_, err := r.Pool.Exec(ctx, query, city, temp, cond)
	return err
}
