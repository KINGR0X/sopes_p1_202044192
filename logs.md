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

## server go

```cmd
kubectl logs -l app=grpc-server-go -n sopes1 --tail=20 --follow
```

## kafka consumer

```cmd
kubectl logs -l app=kafka-consumer -n sopes1 --tail=20 --follow
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
