#!/bin/bash
# Use this script to wait for rabbitmq to really be ready
port=${1:-5672} # can override port from command line
set -e
until timeout 1 bash -c "cat < /dev/null > /dev/tcp/localhost/${port}"; do
  >&2 echo "Rabbit MQ not up yet on localhost:${port}"
  sleep 1
done

echo "Rabbit MQ is up"