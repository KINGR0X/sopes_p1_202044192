# Documentacion

### ¿Cómo funciona Kafka?

Apache Kafka es una paltaforma de streaming distribuido diseñado para manejar grandes volumenes de datos en tiempo real. El funcionamiento de kafka es el siguiente: los productores escriben eventos en kafka, estos datos se publican en topicos, cada topico esta dividio en particiones lo cual pertmite escabilidad, kafka replica las particiones en multiples brokers para la tolerancia a fallos. Los consumer leen los mensajes de los topicos, dichos consumers se organizan en grupos consumer groups donde cada particion es asignada a un unico consumidr dentro del grupo para asi distribuir la carga de procesamiento. Los consumidores leen los mensajes en orden y los procesa en tiempo real para almacenarlos en bases de datos.

Kafka retiene los mensajes por un tiempo, lo cual permite reanudar el consumo en caso de que un consumidor flle.

### ¿Cómo difiere Valkey de Redis?

Redis utiliza single-threaded para la mayoria de operaciones, pero se puede usar multi-threading en las ultimas versiones. Por su parte Valkey en sus ultimas versiones threading I/O lo cual mejorar el paralelismo del sistemas mejorando el rendimiento y la latencia comparado a sus versiones anteriores.

Valket proporciona metricas por slot para monitoreo detallado, mientras que Redis ofrece monitoreo basico.

### ¿Es mejor gRPC que HTTP?

Depende del proyecto, gRPC es más usado para microservicios, o aplicaciones de streaming donde no se permite una alta latencia. HTTP esta mejor adoptado por los navegadores web, permitiendo una mejor integración con las diferentes api, por lo cual es mejor para desarollo web

### ¿Hubo una mejora al utilizar dos replicas en los deployments de API REST y gRPC?

Si, al usar dos replicas la información que llega atravez del Locust la informacion se puede distribuir de mejor manera.

## Deployments del proyecto

### Iniciar el cluster

Comando para iniciar el cluster

```cmd
gcloud container clusters create proyecto2 --num-nodes=4 --region=us-west2-a --tags=allin,allout --machine-type=e2-medium --no-enable-network-policy --disk-size=100GB --disk-type=pd-standard
```

### Crear namespace

```cmd
kubectl apply -f ./k8s/namespace.yaml
```

### Ingress

El Ingress es el encargado de controlar el trafico de entrada al cluster de kubernetes redirigiendo el ingreso a el cliente de rust, el cual fue creado con ayuda de plantilla de helm.

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

## Go gRPC client

```cmd
kubectl apply -f ./k8s/grpc_client.yaml
```

### gRPC rabbit server

```cmd
kubectl apply -f ./k8s/grpc_server_rabbit.yaml
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

## kafka consumer

```cmd
kubectl apply -f ./k8s/consumer.yaml
```

## rabbit consumer

```cmd
kubectl apply -f ./k8s/consumer_rabbit.yaml
```

## Redis deployment

```cmd
kubectl apply -f ./k8s/redis.yaml
```

## valkey deployment

```cmd
kubectl apply -f ./k8s/valkey.yaml
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

## Crear dashboard en grafana con los datos de valkey

Connections -> Add new connection -> redis

addres: valkey.sopes1.svc.cluster.local:6378
Password: sopes999
save & test

## Ejecutar el servicio de Locus local

```cmd
python3 -m venv venv
source venv/bin/activate
pip3 install locust

python app.py http://34.94.158.179.nip.io
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
