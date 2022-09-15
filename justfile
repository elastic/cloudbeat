default:
  @just --list

# Variables

kustomizeVanillaOverlay := "deploy/kustomize/overlays/cloudbeat-vanilla"
kustomizeEksOverlay := "deploy/kustomize/overlays/cloudbeat-eks"

create-kind-cluster:
  kind create cluster --config deploy/k8s/kind/kind-config.yml --wait 30s

install-kind:
  brew install kind

setup-env: install-kind create-kind-cluster

# Vanilla

create-vanilla-deployment-file:
   kustomize build {{kustomizeVanillaOverlay}} --output deploy/k8s/cloudbeat-ds.yaml

build-deploy-cloudbeat: build-cloudbeat load-cloudbeat-image deploy-cloudbeat

build-deploy-cloudbeat-debug: build-cloudbeat-debug load-cloudbeat-image deploy-cloudbeat

load-cloudbeat-image:
  kind load docker-image cloudbeat:latest --name kind-mono

build-cloudbeat:
  GOOS=linux go mod vendor
  GOOS=linux go build -v && docker build -t cloudbeat .

deploy-cloudbeat:
  cp {{env_var('ELASTIC_PACKAGE_CA_CERT')}} {{kustomizeVanillaOverlay}}
  kubectl delete -k {{kustomizeVanillaOverlay}} -n kube-system & kubectl apply -k {{kustomizeVanillaOverlay}} -n kube-system
  rm {{kustomizeVanillaOverlay}}/ca-cert.pem

build-cloudbeat-debug:
  GOOS=linux CGO_ENABLED=0 go build -gcflags "all=-N -l" && docker build -f Dockerfile.debug -t cloudbeat .

delete-cloudbeat:
  kubectl delete -k {{kustomizeVanillaOverlay}} -n kube-system

# EKS
create-eks-deployment-file:
    kustomize build {{kustomizeEksOverlay}} --output deploy/eks/cloudbeat-ds.yaml

deploy-eks-cloudbeat:
  kubectl delete -k {{kustomizeEksOverlay}} -n kube-system & kubectl apply -k {{kustomizeEksOverlay}} -n kube-system

#General

logs-cloudbeat:
  kubectl logs -f --selector="k8s-app=cloudbeat" -n kube-system

logs-cloudbeat-file:
  kubectl logs -f --selector="k8s-app=cloudbeat" -n kube-system > cloudbeat-logs.ndjson

build-kibana-docker:
  node scripts/build --docker-images --skip-docker-ubi --skip-docker-centos -v

elastic-stack-up:
  elastic-package stack up -vd

elastic-stack-down:
  elastic-package stack down

elastic-stack-connect-kind:
  ID=$( docker ps --filter name=kind-mono-control-plane --format "{{{{.ID}}" ) && \
  docker network connect elastic-package-stack_default $ID

ssh-cloudbeat:
    CLOUDBEAT_POD=$( kubectl get pods --no-headers -o custom-columns=":metadata.name" -n kube-system | grep "cloudbeat" ) && \
    kubectl exec --stdin --tty "${CLOUDBEAT_POD}" -n kube-system -- /bin/bash

expose-ports:
    CLOUDBEAT_POD=$( kubectl get pods --no-headers -o custom-columns=":metadata.name" -n kube-system | grep "cloudbeat" ) && \
    kubectl port-forward $CLOUDBEAT_POD -n kube-system 40000:40000 8080:8080


#### TESTS ####

TEST_POD := 'test-pod-v1'
TESTS_RELEASE := 'cloudbeat-test'
TEST_LOGS_DIRECTORY := 'test-logs'
POD_STATUS_UNKNOWN := 'Unknown'
POD_STATUS_PENDING := 'Pending'
POD_STATUS_RUNNING := 'Running'
TIMEOUT := '1200s'
TESTS_TIMEOUT := '60m'
VERSION := '$(make -s get-version)-SNAPSHOT'
NAMESPACE := 'kube-system'


patch-cb-yml-tests:
  kubectl kustomize deploy/k8s/kustomize/test > tests/deploy/cloudbeat-pytest.yml

build-pytest-docker:
  cd tests; docker build -t {{TESTS_RELEASE}} .

load-pytest-kind:
  kind load docker-image {{TESTS_RELEASE}}:latest --name kind-mono

deploy-tests-helm-ci target range='0..':
  helm upgrade --wait --timeout={{TIMEOUT}} --install --values tests/deploy/values/ci.yml --set testData.marker={{target}} --set testData.range={{range}} --set elasticsearch.imageTag={{VERSION}} -n {{NAMESPACE}} {{TESTS_RELEASE}}  tests/deploy/k8s-cloudbeat-tests/

deploy-tests-helm-ci-agent target range='0..':
  helm upgrade --wait --timeout={{TIMEOUT}} --install --values tests/deploy/values/ci-sa-agent.yml --set testData.marker={{target}} --set testData.range={{range}} --set elasticsearch.imageTag={{VERSION}} -n {{NAMESPACE}} {{TESTS_RELEASE}}  tests/deploy/k8s-cloudbeat-tests/

deploy-local-tests-helm target range='0..':
  helm upgrade --wait --timeout={{TIMEOUT}} --install --values tests/deploy/values/local-host.yml --set testData.marker={{target}} --set testData.range={{range}} --set elasticsearch.imageTag={{VERSION}} -n {{NAMESPACE}} {{TESTS_RELEASE}}  tests/deploy/k8s-cloudbeat-tests/

purge-pvc:
  kubectl delete -f tests/deploy/pvc-deleter.yaml -n {{NAMESPACE}} & kubectl apply -f tests/deploy/pvc-deleter.yaml -n {{NAMESPACE}}

purge-tests:
	helm del {{TESTS_RELEASE}} -n {{NAMESPACE}}

purge-tests-full: purge-tests purge-pvc

gen-report:
  allure generate tests/allure/results --clean -o tests/allure/reports && cp tests/allure/reports/history/* tests/allure/results/history/. && allure open tests/allure/reports

run-tests:
  helm test {{TESTS_RELEASE}} -n {{NAMESPACE}}

run-tests-ci:
  #!/usr/bin/env bash
  helm test {{TESTS_RELEASE}} -n {{NAMESPACE}} --kube-context kind-kind-mono --timeout {{TESTS_TIMEOUT}} --logs 2>&1 | tee test.log
  result_code=${PIPESTATUS[0]}
  SUMMARY=$(cat test.log | sed -n '/summary/,/===/p')
  echo "summary<<EOF" >> "$GITHUB_ENV"
  echo "$SUMMARY" >> "$GITHUB_ENV"
  echo "EOF" >> "$GITHUB_ENV"
  exit $result_code

build-load-run-tests: build-pytest-docker load-pytest-kind run-tests

delete-local-helm-cluster:
  kind delete cluster --name kind-mono

cleanup-create-local-helm-cluster target range='0..': delete-local-helm-cluster create-kind-cluster build-cloudbeat load-cloudbeat-image
  just deploy-local-tests-helm {{target}} {{range}}

# TODO(DaveSys911): Move scripts out of JUSTFILE: https://github.com/elastic/security-team/issues/4291
test-pod-status:
  #!/usr/bin/env sh

  if [ ${STATUS=`kubectl get pod -n kube-system test-pod-v1 --template {{{{.status.phase}}`} ]; then
    echo $STATUS
  else
    echo {{POD_STATUS_UNKNOWN}}
  fi

collect-logs target:
  #!/usr/bin/env sh

  echo 'Collecting logs for target {{target}}...'

  LOG_FILE={{TEST_LOGS_DIRECTORY}}/{{target}}.log
  LOG_FILE_TMP={{TEST_LOGS_DIRECTORY}}/{{target}}.log.tmp

  mkdir -p {{TEST_LOGS_DIRECTORY}}
  echo '' > $LOG_FILE
  
  STATUS={{POD_STATUS_UNKNOWN}}
  while [ $STATUS = {{POD_STATUS_UNKNOWN}} ] || [ $STATUS = {{POD_STATUS_PENDING}} ] || [ $STATUS = {{POD_STATUS_RUNNING}} ]; do
    sleep 5

    STATUS=`just test-pod-status`
    if [ $STATUS = {{POD_STATUS_UNKNOWN}} ]; then
      continue
    fi

    kubectl logs test-pod-v1 -n kube-system 2>&1 > $LOG_FILE_TMP

    if [ `stat -c%s "${LOG_FILE_TMP}"` -gt `stat -c%s "${LOG_FILE}"` ]; then
      cp $LOG_FILE_TMP $LOG_FILE
      echo "Wrote logs to ${LOG_FILE}"
    fi
  done

  rm $LOG_FILE_TMP
  echo 'Done collecting logs for target {{target}}.'

run-test-target target range='0..':
  echo 'Cleaning up cluster for running test target: {{target}}'
  just cleanup-create-local-helm-cluster {{target}} {{range}}

  echo 'Running test target: {{target}}'
  just build-load-run-tests &


run-test-targets range='0..' +targets='file_system_rules k8s_object_rules process_api_server_rules process_controller_manager_rules process_etcd_rules process_kubelet_rules process_scheduler_rules':
  #!/usr/bin/env sh

  echo 'Running tests: {{targets}}'

  for TARGET in {{targets}}; do
    just run-test-target $TARGET {{range}}
    just collect-logs $TARGET
  done

