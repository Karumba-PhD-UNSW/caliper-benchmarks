#!/usr/bin/env bash
sudo npx caliper benchmark run  --caliper-bind-sut fabric:1.4.1 --caliper-workspace . --caliper-benchconfig benchmarks/samples/fabric/energypie/config.yaml --caliper-networkconfig networks/fabric/fabric-v1.4.1/4org1peergoleveldb/fabric-go.yaml --"$1" 
