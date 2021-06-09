#!/usr/bin/env bash

set -e

echo "building Frontend"
pushd frontend/
npm i
npm run build -- --env baseUrl=/json
popd

echo "building Backend"
go build ./cmd/server

echo "run ./server to start a server"
echo "go to http://localhost:8080/static"
