#!/bin/bash

# Don't fail silently when a step doesn't succeed
set -e

if [ -z "$TRAVIS" ]; then
    echo "this script is intended to be run only on travis" >&2
    exit 1
fi

function goreleaser() {
    curl -sL https://git.io/goreleaser | bash
}

function gcr_push() {
    openssl aes-256-cbc -K $encrypted_222a2009ef6d_key -iv $encrypted_222a2009ef6d_iv -in heptio-images-c14f11347d8b.json.enc -out heptio-images-c14f11347d8b.json -d
    gcloud auth activate-service-account --key-file heptio-images-c14f11347d8b.json
    # https://github.com/travis-ci/travis-ci/issues/9905
    unset GIT_HTTP_USER_AGENT
    IMAGE_BRANCH="$BRANCH" DOCKER="gcloud docker -- " make container push
}

if [ ! -z "$TRAVIS_TAG" ]; then
    if [ "$(./sonobuoy version --short)" != "$TRAVIS_TAG" ]; then
        echo "sonobuoy version does not match tagged version!" >&2
        echo "sonobuoy short version is $(./sonobuoy version --short)" >&2
        echo "tag is $TRAVIS_TAG" >&2
        echo "sonobuoy full version info is $(./sonobuoy version)" >&2
        exit 1
    fi

    goreleaser --skip-validate
    gcr_push
fi

if [ "$BRANCH" == "master" ]; then
    gcr_push
fi