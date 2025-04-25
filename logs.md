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
