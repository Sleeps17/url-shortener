#!/bin/bash

NETWORK_NAME="proxynet"

if ! docker network inspect $NETWORK_NAME &> /dev/null; then
    echo "Creating network $NETWORK_NAME"
    docker network create $NETWORK_NAME
else
    echo "Network $NETWORK_NAME already exists"
fi
