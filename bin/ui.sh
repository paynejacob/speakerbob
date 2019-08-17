#!/bin/sh
docker-compose up -d
pushd web
  yarn build --watch
popd