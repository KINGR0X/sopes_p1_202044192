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

## Grafana

grafana con helm

```cmd
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

Instalacion de grafana

```cmd
helm install my-grafana grafana/grafana
```

Exponer el servicio de grafana

```cmd
kubectl expose service my-grafana --type=NodePort --target-port=3000 --name=grafana-node-port-service
```

Obtener la ip del nodo

```cmd
kubectl get nodes -o wide
```

Obtner la contraseña de grafana

```cmd
kubectl get secret --namespace default my-grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
```
