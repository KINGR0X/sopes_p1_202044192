
for contenedor in $(docker ps -aq)
do

    docker rm -f $contenedor

done