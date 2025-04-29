# Proyecto 2 sopes 1

## Iniciar el cluster

us-west1-a

```cmd
gcloud container clusters create proyecto2 --num-nodes=4 --region=us-west2-a --tags=allin,allout --machine-type=e2-medium --no-enable-network-policy --disk-size=100GB --disk-type=pd-standard
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

## Api_rust

```cmd
kubectl apply -f ./k8s/rust.yaml
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
kubectl create -f 'https://strimzi.io/install/latest?namespace=kafka' -n kafka
```

```cmd
kubectl apply -f https://strimzi.io/examples/latest/kafka/kraft/kafka-single-node.yaml -n kafka
```

kafka consumer

```cmd
kubectl apply -f ./k8s/consumer.yaml
```

## Redis deployment

```cmd
kubectl apply -f ./k8s/redis.yaml
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

escoger cualquier IP externa de los nodos enumerados por el comando anterior

**_ej: 34.19.107.66_**

obtener el port de grafana

```cmd
kubectl get svc grafana-node-port-service
```

**_80:30511/TCP_**

direccion para acceder a grafana

**http://34.19.107.66:30511/login**

Obtner la contraseña de grafana

```cmd
kubectl get secret --namespace default my-grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
```

## Crear dashboard en grafana con los datos de redis

Connections -> Add new connection -> redis

addres: redis.sopes1.svc.cluster.local:6379
Password: sopes999
save & test

## Peticiones de redis en grafana

Obtener datos de un país:

```cmd
HGETALL country:GT
```

Obtener datos globales:

```cmd
HGETALL weather:global
```

## Ejecutar el servicio de Locus local

```cmd
python3 -m venv venv
source venv/bin/activate
pip3 install locust

locust -f app.py --headless -u 10 -r 10 -t 10000 --host http://34.94.179.180.nip.io
```

## Limpiar redis

acceder al pod

```cmd
kubectl exec -it "pod" -n sopes1 -- /bin/bash
```

comando para borrar la base de datos

```cmd
redis-cli -a sopes999 FLUSHDB
```
