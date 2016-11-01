#!/bin/bash

VERSION=$(cat ./api/version.go | grep "var VERSION" | awk ' { print $4 } ' | sed s/\"//g)

docker build -t arkadiko .
docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
docker tag arkadiko:latest tfgco/arkadiko:$VERSION.$TRAVIS_BUILD_NUMBER
docker push tfgco/arkadiko:$VERSION.$TRAVIS_BUILD_NUMBER

DOCKERHUB_LATEST=$(python get_latest_tag.py)

if [ "$DOCKERHUB_LATEST" != "$VERSION.$TRAVIS_BUILD_NUMBER" ]; then
  echo "Last version is not in docker hub!"
  echo "docker hub: $DOCKERHUB_LATEST, expected: $VERSION.$TRAVIS_BUILD_NUMBER"
  exit 1
fi
