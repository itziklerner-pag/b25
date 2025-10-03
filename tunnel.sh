#!/bin/bash
# SSH Tunnel to B25 VPS
# Run this on your LOCAL machine

ssh -N \
  -L 3000:localhost:3000 \
  -L 3001:localhost:3001 \
  -L 8000:localhost:8000 \
  -L 8080:localhost:8080 \
  -L 8081:localhost:8081 \
  -L 8082:localhost:8082 \
  -L 8083:localhost:8083 \
  -L 8084:localhost:8084 \
  -L 8085:localhost:8085 \
  -L 8086:localhost:8086 \
  -L 9090:localhost:9090 \
  -L 9093:localhost:9093 \
  -L 9100:localhost:9100 \
  -L 9101:localhost:9101 \
  -L 9102:localhost:9102 \
  -L 9103:localhost:9103 \
  -L 9104:localhost:9104 \
  -L 9105:localhost:9105 \
  -L 9106:localhost:9106 \
  -L 50051:localhost:50051 \
  -L 50052:localhost:50052 \
  -L 50053:localhost:50053 \
  -L 50054:localhost:50054 \
  -L 50055:localhost:50055 \
  -L 50056:localhost:50056 \
  mm@66.94.120.149
