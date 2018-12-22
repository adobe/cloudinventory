#!/bin/sh

if [[ $TRAVIS_PULL_REQUEST == “false” ]] && [[ $TRAVIS_BRANCH == “master” ]]; then
  make docker-push
  make docker-push-arm
fi
