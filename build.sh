#!/bin/sh
docker run --rm -v "$(pwd):/app" -w '/app' golang:alpine go build -o get-aws-env-alpine

docker run --rm -v "$(pwd):/app" -w '/app' golang:bullseye go build -o get-aws-env-amd64