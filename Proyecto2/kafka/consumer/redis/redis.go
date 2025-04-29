package redis

import (
	"consumer/structs"
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	redisLock    = &sync.Mutex{}
	redisClient  *redis.Client
	weatherTypes = map[int]string{
		0: "rainy",
		1: "cloudy",
		2: "sunny",
	}
)

func InitRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "redis.sopes1.svc.cluster.local:6379",
		Password: "sopes999",
		DB:       0,
	})
}

func RedisInstance() *redis.Client {
	if redisClient == nil {
		redisLock.Lock()
		defer redisLock.Unlock()
		if redisClient == nil {
			redisClient = InitRedis()
		}
	}
	return redisClient
}

func Insert(value structs.Tweet) {
	ctx := context.Background()
	client := RedisInstance()

	weather, ok := weatherTypes[value.Weather]
	if !ok {
		weather = "unknown"
	}

	// Sorted Sets para cada tipo de clima
	client.ZIncrBy(ctx, fmt.Sprintf("weather:%s", weather), 1, value.Country)
	
	client.HIncrBy(ctx, fmt.Sprintf("country:%s", value.Country), weather, 1)
	client.HIncrBy(ctx, "weather:global", weather, 1)

}

func GetDataForGrafana(ctx context.Context) ([]map[string]interface{}, error) {
	client := RedisInstance()
	var result []map[string]interface{}

	// Obtener todos los países
	countries, err := client.Keys(ctx, "country:*").Result()
	if err != nil {
		return nil, err
	}

	for _, countryKey := range countries {
		country := countryKey[8:] // remove "country:" prefix
		weatherData, err := client.HGetAll(ctx, countryKey).Result()
		if err != nil {
			return nil, err
		}

		for weather, count := range weatherData {
			result = append(result, map[string]interface{}{
				"country": country,
				"weather": weather,
				"count":   count,
			})
		}
	}

	return result, nil
}