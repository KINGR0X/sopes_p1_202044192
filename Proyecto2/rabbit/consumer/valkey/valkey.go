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
	return InsertWithContext(context.Background(), value)
}

func InsertWithContext(ctx context.Context, value structs.Tweet) error {
	client := ValkeyInstance()
	weather, ok := weatherTypes[value.Weather]
	if !ok {
		weather = "unknown"
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		// Iniciar transacción MULTI
		if err := client.Do(ctx, client.B().Multi().Build()).Error(); err != nil {
			lastErr = fmt.Errorf("failed to start transaction: %w", err)
			time.Sleep(retryDelay)
			continue
		}

		// Comandos de la transacción
		cmds := []valkey.Completed{
			client.B().Zincrby().Key(fmt.Sprintf("weather:%s", weather)).Increment(1).Member(value.Country).Build(),
			client.B().Hincrby().Key(fmt.Sprintf("country:%s", value.Country)).Field(weather).Increment(1).Build(),
			client.B().Hincrby().Key("weather:global").Field(weather).Increment(1).Build(),
		}

		// Encolar comandos
		for _, cmd := range cmds {
			if err := client.Do(ctx, cmd).Error(); err != nil {
				client.Do(ctx, client.B().Discard().Build()) // Descartar transacción
				lastErr = fmt.Errorf("failed to queue command: %w", err)
				time.Sleep(retryDelay)
				continue
			}
		}

		// Ejecutar transacción
		_, err := client.Do(ctx, client.B().Exec().Build()).ToArray()
		if err != nil {
			lastErr = fmt.Errorf("transaction failed: %w", err)
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

    // Usar MULTI/EXEC para lectura atómica
    if err := client.Do(ctx, client.B().Multi().Build()).Error(); err != nil {
        return nil, fmt.Errorf("failed to start transaction: %w", err)
    }

    // 1. Obtener claves de países
    keysCmd := client.Do(ctx, client.B().Keys().Pattern("country:*").Build())
    
    // 2. Obtener datos globales
    globalCmd := client.Do(ctx, client.B().Hgetall().Key("weather:global").Build())

    // Ejecutar transacción
    execRes, err := client.Do(ctx, client.B().Exec().Build()).ToArray()
    if err != nil {
        return nil, fmt.Errorf("transaction failed: %w", err)
    }

    // Procesar resultados
    if len(execRes) != 2 {
        return nil, fmt.Errorf("unexpected number of results: %d", len(execRes))
    }

    // Procesar datos globales
    globalData, err := globalCmd.AsStrMap()
    if err != nil {
        return nil, fmt.Errorf("failed to parse global data: %w", err)
    }

    for weather, count := range globalData {
        result = append(result, map[string]interface{}{
            "country": "global",
            "weather": weather,
            "count":   count,
        })
    }

    // Procesar países
    keys, err := keysCmd.AsStrSlice()
    if err != nil {
        return nil, fmt.Errorf("failed to parse country keys: %w", err)
    }

    for _, key := range keys {
        country := key[8:] // Eliminar "country:" del prefijo
        dataCmd := client.Do(ctx, client.B().Hgetall().Key(key).Build())
        countryData, err := dataCmd.AsStrMap()
        if err != nil {
            continue // O manejar el error según sea necesario
        }

        for weather, count := range countryData {
            result = append(result, map[string]interface{}{
                "country": country,
                "weather": weather,
                "count":   count,
            })
        }
    }

    return result, nil
}