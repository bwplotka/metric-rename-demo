#!/bin/zsh

#DOCKER_TAG=v0.1.0
#DOCKER_IMAGE=quay.io/bwplotka/my-app

pushd my-org/my-app || exit 1
  GOOS=linux GOARCH=amd64 go build -o ../../.build/linux-amd64/my-app . || exit 1
popd || exit 1

echo "Building ${DOCKER_IMAGE}:${DOCKER_TAG}"
docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} . || exit 1

if [[ ${DOCKER_PUSH} == "yes" ]]; then
  docker push ${DOCKER_IMAGE}:${DOCKER_TAG}
fi
