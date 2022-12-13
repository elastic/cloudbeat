#!/usr/bin/env bash
set -euxo pipefail
IMAGE_SUFFIX=$1 # ${{ github.run_id }}
ECR_REGISTRY=$2 # aws ecr path


load_and_push_image () {
    docker load -i /tmp/"$1"-"${IMAGE_SUFFIX}".tar
    docker tag "$2" $ECR_REGISTRY/"$2"
    docker push $ECR_REGISTRY/"$2"
}

load_and_push_image cloudbeat "cloudbeat:latest" &
load_and_push_image pytest "cloudbeat-test:latest"