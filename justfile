# Variables
CLOUDBEAT_VERSION := ''
kustomizeVanillaOverlay := "deploy/kustomize/overlays/cloudbeat-vanilla"
kustomizeEksOverlay := "deploy/kustomize/overlays/cloudbeat-eks"
cspPoliciesPkg := "github.com/elastic/csp-security-policies"
hermitActivationScript := "bin/activate-hermit"

# General

create-kind-cluster kind='kind-multi':
  kind create cluster --config deploy/k8s/kind/{{kind}}.yml --wait 30s

setup-env: create-kind-cluster elastic-stack-connect-kind

linter-setup:
  source {{hermitActivationScript}} || true
  pre-commit install -f

# Vanilla

create-vanilla-deployment-file:
   kustomize build {{kustomizeVanillaOverlay}} --output deploy/k8s/cloudbeat-ds.yaml

build-deploy-cloudbeat: build-cloudbeat load-cloudbeat-image deploy-cloudbeat

build-load-both: build-deploy-cloudbeat load-pytest-kind

build-deploy-cloudbeat-debug: build-cloudbeat-debug load-cloudbeat-image deploy-cloudbeat

build-replace-cloudbeat: build-binary
  ./scripts/remote_replace_cloudbeat.sh

build-replace-bundle: build-opa-bundle
  ./scripts/remote_replace_bundle.sh

load-cloudbeat-image kind='kind-multi':
  kind load docker-image cloudbeat:latest --name {{kind}}

build-opa-bundle:
  mage BuildOpaBundle

build-binary:
  GOOS=linux go mod vendor
  GOOS=linux go build -v

build-cloudbeat: build-opa-bundle build-binary
  docker build -t cloudbeat .

deploy-cloudbeat:
  cp {{env_var('ELASTIC_PACKAGE_CA_CERT')}} {{kustomizeVanillaOverlay}}
  kubectl delete -k {{kustomizeVanillaOverlay}} -n kube-system & kubectl apply -k {{kustomizeVanillaOverlay}} -n kube-system
  rm {{kustomizeVanillaOverlay}}/ca-cert.pem

build-cloudbeat-debug: build-opa-bundle
  GOOS=linux go mod vendor
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
    CLOUDBEAT_POD=$( kubectl get pods -o=name -n kube-system | grep -m 1 "cloudbeat" ) && \
    kubectl logs -f "${CLOUDBEAT_POD}" -n kube-system

build-kibana-docker:
  node scripts/build --docker-images --skip-docker-ubi --skip-docker-centos -v

elastic-stack-up:
  elastic-package stack up -vd

elastic-stack-down:
  elastic-package stack down

elastic-stack-connect-kind kind='kind-multi':
  ./.ci/scripts/connect_kind.sh {{kind}}

ssh-cloudbeat:
    CLOUDBEAT_POD=$( kubectl get pods -o=name -n kube-system | grep -m 1 "cloudbeat" ) && \
    kubectl exec --stdin --tty "${CLOUDBEAT_POD}" -n kube-system -- /bin/bash

expose-ports:
    CLOUDBEAT_POD=$( kubectl get pods -o=name -n kube-system | grep -m 1 "cloudbeat" ) && \
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
ELK_STACK_VERSION := env_var('ELK_VERSION')
NAMESPACE := 'kube-system'
ECR_CLOUDBEAT_TEST := 'public.ecr.aws/z7e1r9l0/'

patch-cb-yml-tests:
  kubectl kustomize deploy/k8s/kustomize/test > tests/deploy/cloudbeat-pytest.yml

build-pytest-docker:
  cd tests; docker build -t {{TESTS_RELEASE}} .

load-pytest-kind kind='kind-multi': build-pytest-docker
  kind load docker-image {{TESTS_RELEASE}}:latest --name {{kind}}

load-pytest-eks:
  aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin public.ecr.aws/z7e1r9l0
  docker tag {{TESTS_RELEASE}}:latest {{ECR_CLOUDBEAT_TEST}}{{TESTS_RELEASE}}:latest
  docker push {{ECR_CLOUDBEAT_TEST}}{{TESTS_RELEASE}}:latest

deploy-tests-helm target range='' values_file='tests/deploy/values/ci.yml':
  helm upgrade --wait --timeout={{TIMEOUT}} --install --values {{values_file}} --set testData.marker={{target}} --set testData.range={{range}} --set elasticsearch.imageTag={{ELK_STACK_VERSION}} --set kibana.imageTag={{ELK_STACK_VERSION}} --namespace={{NAMESPACE}} {{TESTS_RELEASE}} tests/deploy/k8s-cloudbeat-tests/

purge-tests:
	helm del {{TESTS_RELEASE}} -n {{NAMESPACE}} & kubectl delete pvc --all -n {{NAMESPACE}}

gen-report:
  allure generate tests/allure/results --clean -o tests/allure/reports && cp tests/allure/reports/history/* tests/allure/results/history/. && allure open tests/allure/reports

run-tests target='default' kind='kind-multi':
  helm test {{TESTS_RELEASE}} -n {{NAMESPACE}} --kube-context kind-{{kind}} --timeout {{TESTS_TIMEOUT}} --logs

build-load-run-tests: build-pytest-docker load-pytest-kind run-tests

delete-local-helm-cluster kind='kind-multi':
  kind delete cluster --name {{kind}}

cleanup-create-local-helm-cluster target range='..': delete-local-helm-cluster create-kind-cluster build-cloudbeat load-cloudbeat-image
  just deploy-tests-helm tests/deploy/values/local-host.yml {{target}} {{range}}


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

run-test-target target range='..':
  echo 'Cleaning up cluster for running test target: {{target}}'
  just cleanup-create-local-helm-cluster {{target}} {{range}}

  echo 'Running test target: {{target}}'
  just build-load-run-tests &


run-test-targets range='..' +targets='file_system_rules k8s_object_rules process_api_server_rules process_controller_manager_rules process_etcd_rules process_kubelet_rules process_scheduler_rules':
  #!/usr/bin/env sh

  echo 'Running tests: {{targets}}'

  for TARGET in {{targets}}; do
    just run-test-target $TARGET {{range}}
    just collect-logs $TARGET
  done
