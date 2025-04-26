package redis

import (
	"consumer/structs"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

var redisLock = &sync.Mutex{}

var redisClient *redis.Client

func InitRedis() *redis.Client {

	host := "redis"
	port := "6379"
	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: "sopes999",
		DB:       0,
	})

	return client
}

func RedisInstance() *redis.Client {

	if redisClient == nil {
		redisLock.Lock()
		defer redisLock.Unlock()

		if redisClient == nil {
			fmt.Println("Creating single redis instance now.")
			redisClient = InitRedis()
		} else {
			fmt.Println("Single instance already created.")
		}

	} else {
		fmt.Println("Single instance already created.")
	}

	return redisClient
}

func Insert(value structs.Tweet) {
	ctx := context.Background()
	client := RedisInstance()

	weather := "sunny"

	switch value.Weather {
	case 1:
		weather = "rainy"
	case 2:
		weather = "cloudy"
	case 3:
		weather = "snowy"
	}

	// Incrementar contador por país y clima
	newValue, err := client.HIncrBy(ctx, value.Country, weather, 1).Result()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Nuevo valor de %s en %s: %d\n", value.Country, weather, newValue)

	// Incrementar contador total por país
	newValue, err = client.HIncrBy(ctx, value.Country, "total", 1).Result()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Nuevo valor de %s en %s: %d\n", value.Country, "total", newValue)


}