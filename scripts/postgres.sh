#!/usr/bin/env bash

docker network create -d bridge forzatelemetry-network
docker volume create forzatelemetry-postgres

docker create \
    -it --name forzatelemetry-postgres \
    --network forzatelemetry-network \
    --mount source=forzatelemetry-postgres,target=/var/lib/postgresql/data \
    -p 127.0.0.1:5432:5432 \
    -e POSTGRES_PASSWORD=owncgbwpwmyyiq postgres

docker start forzatelemetry-postgres

# stop the container on interupt
trap "docker stop forzatelemetry-postgres" SIGINT

docker logs forzatelemetry-postgres --tail 100 --follow
