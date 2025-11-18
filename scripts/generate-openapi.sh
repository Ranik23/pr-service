#!/bin/sh
PORT=${SERVER_PORT:-8080}
sed "s|\${SERVER_PORT}|${PORT}|g" /spec/openapi.template.yml > /spec/openapi.yml