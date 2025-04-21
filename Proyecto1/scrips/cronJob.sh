#!/bin/bash

# Definir los comandos como listas de argumentos
commands=(
    "stress --cpu 1 --timeout 600"  # Consume CPU 
    "stress --vm 1 --vm-bytes 512M --timeout 600"  # Consume RAM 
    "stress --io 1 --timeout 600"  # Consume I/O 
    "stress --hdd 1 --hdd-bytes 1G --timeout 600"  # Consume Disco 
)

# Generar 10 contenedores de manera aleatoria
for i in {1..10}
do
    # Seleccionar un comando aleatorio
    command=${commands[$RANDOM % ${#commands[@]}]}

    # Generar un nombre único para el contenedor usando /dev/urandom
    container_name="container_$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 8 | head -n 1)"

    # Ejecutar el contenedor con el comando seleccionado
    docker run -d --name $container_name containerstack/alpine-stress $command
done