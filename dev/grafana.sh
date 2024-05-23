#!/usr/bin/env bash

docker network create -d bridge forzatelemetry-network

docker create \
    -it --name forzatelemetry-grafana \
    --network forzatelemetry-network \
    --publish 127.0.0.1:3000:3000 \
    --mount type=bind,source="$(pwd)"/dev/provisioning,target=/etc/grafana/provisioning,readonly \
    --mount type=bind,source="$(pwd)"/grafana/dashboards,target=/var/lib/grafana/dashboards,readonly \
    --env GF_AUTH_ANONYMOUS_ENABLED=true \
    grafana/grafana-oss

docker start forzatelemetry-grafana

# stop the container on interupt
trap "docker stop forzatelemetry-grafana" SIGINT

docker logs forzatelemetry-grafana --tail 100 --follow
