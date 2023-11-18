#!/bin/sh
eks_dir="deploy/k8s/kustomize/overlays/cloudbeat-eks/"
cd "$eks_dir" || exit
touch env.aws
printf "aws.key=%s\naws.secret=%s\n" "$AWS_ACCESS_KEY" "$AWS_SECRET_ACCESS_KEY" >env.aws
