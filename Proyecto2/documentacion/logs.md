## Monitorear Ingress

```cmd
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx | grep "POST /input"
```

## Api-rust

```cmd
kubectl logs -l app=api-rust -n sopes1 --tail=20 --follow
```

## Cliente go

```cmd
kubectl logs -l app=grpc-client-go -n sopes1 --tail=20 --follow
```

## server go kafka

```cmd
kubectl logs -l app=grpc-server-go -n sopes1 --tail=20 --follow
```

## server go rabbit

```cmd
kubectl logs -l app=grpc-server-rabbit -n sopes1 --tail=20 --follow
```

## kafka consumer

```cmd
kubectl logs -l app=kafka-consumer -n sopes1 --tail=20 --follow
```

## rabbit consumer

```cmd
kubectl logs -l app=rabbitmq-consumer -n sopes1 --tail=20 --follow
```

## Obtener pods del namespace sopes1

```cmd
kubectl get pods -n sopes1
```

## Obtener pods del namespace kafka

```cmd
kubectl get pods -n kafka
```

## Eliminar deployment

```cmd
kubectl delete deployment -n sopes1 kafka-consumer
```

## Servicios del namespace de kafka

```cmd
kubectl get svc -n kafka
```

## revisar que valkey esta guardando datos

```cmd
kubectl exec -n sopes1 valkey-54c765597b-cbbk4 -- redis-cli -h valkey.sopes1.svc.cluster.local -p 6378 -a sopes999 HGETALL weather:global
```

## revisar que redis esta guardando datos

```cmd
kubectl exec -n sopes1 redis-9fb464557-h2sx6 -- redis-cli -h redis.sopes1.svc.cluster.local -p 6379 -a sopes999 HGETALL weather:global
```

## Eliminar deployment de redis

```cmd
kubectl delete -f ./k8s/redis.yaml -n sopes1
```

## Eliminar deployment de valkey

```cmd
kubectl delete -f ./k8s/valkey.yaml -n sopes1
```
