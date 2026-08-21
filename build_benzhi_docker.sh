#!/usr/bin/env bash
set -euo pipefail
docker build --platform linux/amd64 -f benzhi.Dockerfile -t go-farm-production-benzhi:bug032 .
