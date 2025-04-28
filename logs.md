## Cliente go

```cmd
kubectl logs -l app=grpc-client-go -n sopes1 --tail=20 --follow
```

## server go

```cmd
kubectl logs -l app=grpc-server-go -n sopes1 --tail=20 --follow
```

## Monitorear Ingress

```cmd
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx | grep "POST /grpc-go"
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
