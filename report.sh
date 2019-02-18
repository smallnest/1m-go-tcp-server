#!/bin/bash

REPLICAS=$1

for (( c=0; c<${REPLICAS}; c++ ))
do
    docker logs 1mclient_$c|egrep "mean|stddev"|tail -3 >> metrics.log
done

grep "mean rate" metrics.log |awk '{s+=$5} END {print s}'