#!/bin/bash

docker build -t agrirouter-go-sdk-oapi-gen -f tools/oapi/Dockerfile tools/oapi

ENTRYPOINT="oapi-codegen" tools/dockerized.sh agrirouter-go-sdk-oapi-gen --config /app/internal/oapi/oapi_codegen_client.yaml openapi.yaml
ENTRYPOINT="oapi-codegen" tools/dockerized.sh agrirouter-go-sdk-oapi-gen --config /app/internal/oapi/oapi_codegen_models.yaml openapi.yaml
ENTRYPOINT="oapi-codegen" tools/dockerized.sh agrirouter-go-sdk-oapi-gen --config /app/internal/tests/test_server/oapi_codegen_server.yaml openapi.yaml



