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

**cambiar el host**

### Ingress

creado con ayuda de plantilla de helm

instalar el ingress-nginx

```cmd

helm upgrade --install ingress-nginx ingress-nginx \
 --repo https://kubernetes.github.io/ingress-nginx \
 --namespace ingress-nginx --create-namespace
```

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
