package valkey

import (
	"consumer/structs"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/valkey-io/valkey-go"
)

var (
	valkeyLock    = &sync.Mutex{}
	valkeyClient  valkey.Client
	weatherTypes = map[int]string{
		0: "rainy",
		1: "cloudy",
		2: "sunny",
	}
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

// Resto del código permanece igual...
func Insert(value structs.Tweet) {
	ctx := context.Background()
	client := ValkeyInstance()

	weather, ok := weatherTypes[value.Weather]
	if !ok {
		weather = "unknown"
	}

	// Sorted Sets for each weather type
	client.Do(ctx, client.B().Zincrby().Key(fmt.Sprintf("weather:%s", weather)).Increment(1).Member(value.Country).Build())
	
	// Increment counters in hashes
	client.Do(ctx, client.B().Hincrby().Key(fmt.Sprintf("country:%s", value.Country)).Field(weather).Increment(1).Build())
	client.Do(ctx, client.B().Hincrby().Key("weather:global").Field(weather).Increment(1).Build())
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