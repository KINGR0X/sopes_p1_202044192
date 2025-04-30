package valkey

import (
	"consumer/structs"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/valkey-io/valkey-go"
)

var (
	valkeyLock    = &sync.Mutex{}
	valkeyClient  valkey.Client
	valkeyPool    = &sync.Pool{
		New: func() interface{} {
			return InitValkey()
		},
	}
	weatherTypes = map[int]string{
		0: "rainy",
		1: "cloudy",
		2: "sunny",
	}
	maxRetries = 3
	retryDelay = 1 * time.Second
)


func InitValkey() valkey.Client {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{"valkey.sopes1.svc.cluster.local:6378"},
		Password:    "sopes999",
	})
	if err != nil {
		log.Fatalf("Failed to connect to Valkey: %v", err)
	}
	return client
}

func ValkeyInstance() valkey.Client {
	if valkeyClient == nil {
		valkeyLock.Lock()
		defer valkeyLock.Unlock()
		if valkeyClient == nil {
			valkeyClient = InitValkey()
		}
	}
	return valkeyClient
}

func Insert(value structs.Tweet) error {
	ctx := context.Background()
	client := ValkeyInstance()

	weather, ok := weatherTypes[value.Weather]
	if !ok {
		weather = "unknown"
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		// Sorted Sets for each weather type
		_, err := client.Do(ctx, client.B().Zincrby().Key(fmt.Sprintf("weather:%s", weather)).Increment(1).Member(value.Country).Build()).AsInt64()
		if err != nil {
			lastErr = err
			time.Sleep(retryDelay)
			continue
		}
		
		// Increment counters in hashes
		_, err = client.Do(ctx, client.B().Hincrby().Key(fmt.Sprintf("country:%s", value.Country)).Field(weather).Increment(1).Build()).AsInt64()
		if err != nil {
			lastErr = err
			time.Sleep(retryDelay)
			continue
		}

		_, err = client.Do(ctx, client.B().Hincrby().Key("weather:global").Field(weather).Increment(1).Build()).AsInt64()
		if err != nil {
			lastErr = err
			time.Sleep(retryDelay)
			continue
		}

		return nil
	}

	return fmt.Errorf("failed after %d retries: %v", maxRetries, lastErr)
}

func GetDataForGrafana(ctx context.Context) ([]map[string]interface{}, error) {
	client := ValkeyInstance()
	var result []map[string]interface{}

	// Get all countries
	keysCmd := client.Do(ctx, client.B().Keys().Pattern("country:*").Build())
	keys, err := keysCmd.AsStrSlice()
	if err != nil {
		return nil, err
	}

	for _, countryKey := range keys {
		country := countryKey[8:] 
		hashCmd := client.Do(ctx, client.B().Hgetall().Key(countryKey).Build())
		weatherData, err := hashCmd.AsStrMap()
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

func InsertWithContext(ctx context.Context, value structs.Tweet) error {
    client := ValkeyInstance()

    weather, ok := weatherTypes[value.Weather]
    if !ok {
        weather = "unknown"
    }

    var lastErr error
    for i := 0; i < maxRetries; i++ {
        // Sorted Sets for each weather type
        _, err := client.Do(ctx, client.B().Zincrby().Key(fmt.Sprintf("weather:%s", weather)).Increment(1).Member(value.Country).Build()).AsInt64()
        if err != nil {
            lastErr = err
            time.Sleep(retryDelay)
            continue
        }
        
        // Increment counters in hashes
        _, err = client.Do(ctx, client.B().Hincrby().Key(fmt.Sprintf("country:%s", value.Country)).Field(weather).Increment(1).Build()).AsInt64()
        if err != nil {
            lastErr = err
            time.Sleep(retryDelay)
            continue
        }

        _, err = client.Do(ctx, client.B().Hincrby().Key("weather:global").Field(weather).Increment(1).Build()).AsInt64()
        if err != nil {
            lastErr = err
            time.Sleep(retryDelay)
            continue
        }

        return nil
    }

    return fmt.Errorf("failed after %d retries: %v", maxRetries, lastErr)
}