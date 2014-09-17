#!/bin/bash
set -x
docker stop api
docker rm api
docker run -i --rm --name api api


