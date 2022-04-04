create-kind-cluster:
  kind create cluster --config deploy/k8s/kind/kind-config.yaml

install-kind:
  brew install kind

setup-env: install-kind create-kind-cluster

# Vanilla

build-deploy-cloudbeat: build-cloudbeat load-cloudbeat-image delete-cloudbeat deploy-cloudbeat

build-deploy-cloudbeat-debug: build-cloudbeat-debug load-cloudbeat-image delete-cloudbeat-debug deploy-cloudbeat-debug

load-cloudbeat-image:
  kind load docker-image cloudbeat:latest --name kind-mono

build-cloudbeat:
  GOOS=linux go build -v && docker build -t cloudbeat .

deploy-cloudbeat:
  kubectl apply -f deploy/k8s/cloudbeat-ds.yaml -n kube-system

build-cloudbeat-debug:
  GOOS=linux CGO_ENABLED=0 go build -gcflags "all=-N -l" && docker build -f Dockerfile.debug -t cloudbeat .

deploy-cloudbeat-debug:
   kubectl apply -f deploy/k8s/cloudbeat-ds-debug.yaml -n kube-system

delete-cloudbeat:
  kubectl delete -f deploy/k8s/cloudbeat-ds.yaml -n kube-system

delete-cloudbeat-debug:
  kubectl delete -f deploy/k8s/cloudbeat-ds-debug.yaml -n kube-system


# EKS

build-deploy-eks-cloudbeat: login-aws build-cloudbeat publish-image-to-ecr delete-eks-cloudbeat deploy-eks-cloudbeat

login-aws:
  aws ecr get-login-password --region us-east-2 | docker login --username AWS --password-stdin 704479110758.dkr.ecr.us-east-2.amazonaws.com

publish-image-to-ecr:
  docker tag cloudbeat:latest 704479110758.dkr.ecr.us-east-2.amazonaws.com/cloudbeat:latest & docker push 704479110758.dkr.ecr.us-east-2.amazonaws.com/cloudbeat:latest

deploy-eks-cloudbeat:
  kubectl delete -f deploy/eks/cloudbeat-ds.yaml -n kube-system & kubectl apply -f deploy/eks/cloudbeat-ds.yaml -n kube-system

delete-eks-cloudbeat:
  kubectl delete -f deploy/eks/cloudbeat-ds.yaml -n kube-system


#General

logs-cloudbeat:
  kubectl logs -f --selector="k8s-app=cloudbeat" -n kube-system

logs-cloudbeat-file:
  kubectl logs -f --selector="k8s-app=cloudbeat" -n kube-system > cloudbeat-logs.ndjson

build-kibana-docker:
  node scripts/build --docker-images --skip-docker-ubi --skip-docker-centos -v

elastic-stack-up:
  elastic-package stack up --version=8.1.0-SNAPSHOT

elastic-stack-down:
  elastic-package stack down

ssh-cloudbeat:
    CLOUDBEAT_POD=$( kubectl get pods --no-headers -o custom-columns=":metadata.name" -n kube-system | grep "cloudbeat" ) && \
    kubectl exec --stdin --tty "${CLOUDBEAT_POD}" -n kube-system -- /bin/bash

expose-ports:
    CLOUDBEAT_POD=$( kubectl get pods --no-headers -o custom-columns=":metadata.name" -n kube-system | grep "cloudbeat" ) && \
    kubectl port-forward $CLOUDBEAT_POD -n kube-system 40000:40000 8080:8080