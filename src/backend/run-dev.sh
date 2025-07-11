#!/bin/bash

set -e

docker compose -f ../ci-runner/docker-compose.yml up -d
go build
PROFILE="TEST" RUN_NATIVELY=true USE_MOCK_EMAIL_CLIENT=true ./store
