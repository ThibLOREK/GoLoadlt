#!/usr/bin/env bash
set -euo pipefail

docker compose -f deploy/docker/docker-compose.yml up -d
cp configs/.env.example .env || true
go run ./cmd/server
