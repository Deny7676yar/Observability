#! /bin/bash
set -e

while true
do
    RAND_LEN="$(($RANDOM % 15 + 1))"
    RAND_STR=$(openssl rand -hex $RAND_LEN)
    echo $RAND_STR
    curl http://localhost:9000/process?line="$RAND_STR"
    sleep 0.05
done