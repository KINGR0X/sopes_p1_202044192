# Proyecto 2 sopes 1

## Iniciar el cluster

```cmd
gcloud container clusters create proyecto2 --num-nodes=4 --region=us-west1-a --tags=allin,allout --machine-type=e2-medium --no-enable-network-policy --disk-size=25GB --disk-type pd-standard
```

## Deployments

### Crear namespace

```cmd
kubectl apply -f ./k8s/namespace.yaml
```

### Ingress

creado con ayuda de plantilla de helm

instalar el ingress-nginx

```cmd

helm upgrade --install ingress-nginx ingress-nginx \
 --repo https://kubernetes.github.io/ingress-nginx \
 --namespace ingress-nginx --create-namespace
```

verificar host, aparecera como LoadBalancer Ingress

```cmd
kubectl describe service ingress-nginx-controller --namespace ingress-nginx
```

**cambiar el host en ingress.yaml**

Configurar ingress

```cmd
kubectl apply -f ./k8s/ingress.yaml
```

### Go gRPC client

```cmd
kubectl apply -f ./k8s/grpc_client.yaml
```

## gRPC server

```cmd
kubectl apply -f ./k8s/grpc_server_kafka.yaml
```

## kafka deployment

Desplegado con chart de streamzi

```cmd
kubectl create -f 'https://strimzi.io/install/latest?namespace=sopes1' -n sopes1
```

kafka consumer

```cmd
kubectl apply -f ./k8s/consumer.yaml
```

## Redis deployment

```cmd
kubectl apply .f ./k8s/redis.yaml
```
