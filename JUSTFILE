create-kind-cluster:
  kind create cluster --config deploy/k8s/kind/kind-config.yaml

install-kind:
  brew install kind

setup-env: install-kind create-kind-cluster

load-cloudbeat-image:
  kind load docker-image cloudbeat:latest --name kind-mono

build-cloudbeat:
  GOOS=linux go build -v && docker build -t cloudbeat .

deploy-cloudbeat:
  kubectl delete -f deploy/k8s/kustomize/base/cloudbeat-ds.yml -n kube-system & kubectl apply -f deploy/k8s/kustomize/base/cloudbeat-ds.yml -n kube-system

build-deploy-cloudbeat: build-cloudbeat load-cloudbeat-image deploy-cloudbeat

build-cloudbeat-debug:
  GOOS=linux CGO_ENABLED=0 go build -gcflags "all=-N -l" && docker build -f Dockerfile.debug -t cloudbeat .

deploy-cloudbeat-debug:
  kubectl delete -f deploy/k8s/cloudbeat-ds-debug.yaml -n kube-system & kubectl apply -f deploy/k8s/cloudbeat-ds-debug.yaml -n kube-system

build-deploy-cloudbeat-debug: build-cloudbeat-debug load-cloudbeat-image deploy-cloudbeat-debug

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


#### TESTS ####

TESTS_RELEASE := "cloudbeat-tests"
TIMEOUT := "1200s"

patch-cb-yml-tests:
  kubectl kustomize deploy/k8s/kustomize/test > tests/deploy/cloudbeat-pytest.yml

build-pytest-docker:
  cd tests; docker build -t cloudbeat-test .

load-pytest-kind:
  kind load docker-image cloudbeat-test:latest --name kind-mono

deploy-tests-helm:
  helm upgrade --wait --timeout={{TIMEOUT}} --install --values tests/deploy/values/ci.yml --namespace kube-system {{TESTS_RELEASE}}  tests/deploy/k8s-cloudbeat-tests/

purge-tests:
	helm del {{TESTS_RELEASE}} -n kube-system

gen-report:
  allure generate tests/allure/results --clean -o tests/allure/reports && cp tests/allure/reports/history/* tests/allure/results/history/. && allure open tests/allure/reports

run-tests:
  helm test cloudbeat-tests --namespace kube-system