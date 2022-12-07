#!/bin/sh
eks_dir="deploy/k8s/kustomize/overlays/cloudbeat-eks/"
cd $eks_dir
touch env.aws; echo "aws.key=$AWS_ACCESS_KEY\naws.secret=$AWS_SECRET_ACCESS_KEY" > env.aws
