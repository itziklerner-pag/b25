#!/bin/bash
# Minimal SSH tunnel - only essential ports
# Run this on your LOCAL machine

ssh -N \
  -L 3000:localhost:3000 \
  -L 3001:localhost:3001 \
  -L 9090:localhost:9090 \
  mm@66.94.120.149
