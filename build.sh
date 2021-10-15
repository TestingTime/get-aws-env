#!/bin/sh
docker run --rm -v "$(pwd):/app" -w '/app' golang:alpine go build -o get-aws-env-alpine