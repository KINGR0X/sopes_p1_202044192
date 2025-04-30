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

    // Operación atómica para incrementar los contadores
    pipe := client.TxPipeline()
    
    // Incrementar ranking por país para este clima
    pipe.ZIncrBy(ctx, fmt.Sprintf("weather:%s", weather), 1, value.Country)
    
    // Incrementar contador específico para este país+clima
    pipe.HIncrBy(ctx, fmt.Sprintf("country:%s", value.Country), weather, 1)
    
    // Incrementar contador global para este clima
    pipe.HIncrBy(ctx, "weather:global", weather, 1)
    
    _, err := pipe.Exec(ctx)
    if err != nil {
        fmt.Printf("Error executing Redis transaction: %v\n", err)
    }
}