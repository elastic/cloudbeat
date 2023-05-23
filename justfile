# Variables
CLOUDBEAT_VERSION := ''
kustomizeVanillaOverlay := "deploy/kustomize/overlays/cloudbeat-vanilla"
kustomizeVanillaNoCertOverlay := "deploy/kustomize/overlays/cloudbeat-vanilla-nocert"
kustomizeEksOverlay := "deploy/kustomize/overlays/cloudbeat-eks"
kustomizeAwsOverlay := "deploy/kustomize/overlays/cloudbeat-aws"
cspPoliciesPkg := "github.com/elastic/csp-security-policies"
hermitActivationScript := "bin/activate-hermit"

# use env var if available
export LOCAL_GOARCH := `go env GOARCH`

create-kind-cluster kind='kind-multi':
  kind create cluster --config deploy/k8s/kind/{{kind}}.yml --wait 30s

linter-setup:
  source {{hermitActivationScript}} || true
  pre-commit install -f

create-vanilla-deployment-file:
  @echo "Make sure to run 'eval \$(elastic-package stack shellinit --shell \$(basename \$SHELL))'"
  cp {{env_var('ELASTIC_PACKAGE_CA_CERT')}} {{kustomizeVanillaOverlay}}
  kustomize build {{kustomizeVanillaOverlay}} --output deploy/k8s/cloudbeat-ds.yaml

create-vanilla-deployment-file-nocert:
  kustomize build {{kustomizeVanillaNoCertOverlay}} --output deploy/k8s/cloudbeat-ds-nocert.yaml

create-aws-deployment-file:
  kustomize build {{kustomizeAwsOverlay}} --output deploy/aws/cloudbeat-ds.yaml

build-deploy-cloudbeat kind='kind-multi' $GOARCH=LOCAL_GOARCH:
  just build-cloudbeat-docker-image $GOARCH
  just load-cloudbeat-image {{kind}}
  just deploy-cloudbeat

build-deploy-cloudbeat-nocert $GOARCH=LOCAL_GOARCH:
  just build-cloudbeat-docker-image $GOARCH
  just load-cloudbeat-image
  just deploy-cloudbeat-nocert

build-deploy-cloudbeat-debug $GOARCH=LOCAL_GOARCH:
  just build-cloudbeat-debug $GOARCH
  just load-cloudbeat-image
  just deploy-cloudbeat

build-deploy-cloudbeat-debug-nocert $GOARCH=LOCAL_GOARCH:
  just build-cloudbeat-debug $GOARCH
  just load-cloudbeat-image
  just deploy-cloudbeat-nocert

# Builds cloudbeat binary and replace it in the agents (in the current context - i.e. `kubectl config current-context`)
# Set GOARCH to the desired values amd64|arm64
build-replace-cloudbeat $GOARCH=LOCAL_GOARCH: build-opa-bundle
  just build-binary $GOARCH
  # replace the cloudbeat binary in the agents
  ./scripts/remote_replace_cloudbeat.sh
  # Replace bundle in the agents
  ./scripts/remote_replace_bundle.sh

load-cloudbeat-image kind='kind-multi':
  kind load docker-image cloudbeat:latest --name {{kind}}

build-opa-bundle:
  go mod vendor
  mage BuildOpaBundle

# Builds cloudbeat binary
# Set GOARCH to the desired values amd64|arm64, by default, it will use local arch
build-binary $GOARCH=LOCAL_GOARCH:
  @echo "Building cloudbeat binary for linux/$GOARCH"
  GOOS=linux go mod vendor
  GOOS=linux go build -v

# For backwards compatibility
alias build-cloudbeat := build-cloudbeat-docker-image

# Builds cloudbeat docker image with the OPA bundle included
# Set GOARCH to the desired values amd64|arm64. by default, it will use local arch
build-cloudbeat-docker-image $GOARCH=LOCAL_GOARCH: build-opa-bundle
  just build-binary $GOARCH
  @echo "Building cloudbeat docker image for linux/$GOARCH"
  docker build -t cloudbeat . --platform=linux/$GOARCH

deploy-cloudbeat:
  cp {{env_var('ELASTIC_PACKAGE_CA_CERT')}} {{kustomizeVanillaOverlay}}
  kubectl delete -k {{kustomizeVanillaOverlay}} -n kube-system & kubectl apply -k {{kustomizeVanillaOverlay}} -n kube-system
  rm {{kustomizeVanillaOverlay}}/ca-cert.pem

deploy-cloudbeat-aws:
  kubectl delete -k {{kustomizeAwsOverlay}} -n kube-system || true && kubectl apply -k {{kustomizeAwsOverlay}} -n kube-system

deploy-cloudbeat-nocert:
  kubectl delete -k {{kustomizeVanillaNoCertOverlay}} -n kube-system & kubectl apply -k {{kustomizeVanillaNoCertOverlay}} -n kube-system

# Builds cloudbeat docker image with the OPA bundle included and the debug flag
build-cloudbeat-debug $GOARCH=LOCAL_GOARCH: build-opa-bundle
  GOOS=linux go mod vendor
  GOOS=linux CGO_ENABLED=0 go build -gcflags "all=-N -l" && docker build -f Dockerfile.debug -t cloudbeat . --platform=linux/$GOARCH

delete-cloudbeat:
  cp {{env_var('ELASTIC_PACKAGE_CA_CERT')}} {{kustomizeVanillaOverlay}}
  kubectl delete -k {{kustomizeVanillaOverlay}} -n kube-system

delete-cloudbeat-nocert:
  kubectl delete -k {{kustomizeVanillaNoCertOverlay}} -n kube-system

# EKS
create-eks-deployment-file:
  @echo "Make sure to run 'eval \$(elastic-package stack shellinit --shell \$(basename \$SHELL))'"
  cp {{env_var('ELASTIC_PACKAGE_CA_CERT')}} {{kustomizeEksOverlay}}
  kustomize build {{kustomizeEksOverlay}} --output deploy/eks/cloudbeat-ds.yaml

deploy-eks-cloudbeat:
  kubectl delete -k {{kustomizeEksOverlay}} -n kube-system & kubectl apply -k {{kustomizeEksOverlay}} -n kube-system

#General

logs-cloudbeat:
  CLOUDBEAT_POD=$( kubectl get pods -o=name -n kube-system | grep -m 1 "cloudbeat" ) && \
  kubectl logs -f "${CLOUDBEAT_POD}" -n kube-system

deploy-cloudformation:
  cd deploy/cloudformation && go run .

build-kibana-docker:
  node scripts/build --docker-images --skip-docker-ubi --skip-docker-centos -v

elastic-stack-up:
  elastic-package stack up -vd

elastic-stack-down:
  elastic-package stack down

elastic-stack-connect-kind kind='kind-multi':
  ./scripts/connect_kind.sh {{kind}}

elastic-stack-disconnect-kind kind='kind-multi':
  ./scripts/connect_kind.sh {{kind}} disconnect

ssh-cloudbeat:
  CLOUDBEAT_POD=$( kubectl get pods -o=name -n kube-system | grep -m 1 "cloudbeat" ) && \
  kubectl exec --stdin --tty "${CLOUDBEAT_POD}" -n kube-system -- /bin/bash

expose-ports:
  CLOUDBEAT_POD=$( kubectl get pods -o=name -n kube-system | grep -m 1 "cloudbeat" ) && \
  kubectl port-forward $CLOUDBEAT_POD -n kube-system 40000:40000 8080:8080

#### MOCKS #####

# generate new and update existing mocks from golang interfaces
# and update the license header
generate-mocks:
  mockery --dir . --inpackage --all --with-expecter --case underscore --recursive --exclude vendor
  mage AddLicenseHeaders

# run to validate no mocks are missing
validate-mocks:
  ./scripts/validate-mocks.sh


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

deploy-tests-helm target values_file='tests/deploy/values/ci.yml' range='':
  helm upgrade --wait --timeout={{TIMEOUT}} --install --values {{values_file}} --set testData.marker={{target}} --set testData.range={{range}} --set elasticsearch.imageTag={{ELK_STACK_VERSION}} --set kibana.imageTag={{ELK_STACK_VERSION}} --namespace={{NAMESPACE}} {{TESTS_RELEASE}} tests/deploy/k8s-cloudbeat-tests/

purge-tests:
  helm del {{TESTS_RELEASE}} -n {{NAMESPACE}} & kubectl delete pvc --all -n {{NAMESPACE}}

gen-report:
  allure generate tests/allure/results --clean -o tests/allure/reports && cp tests/allure/reports/history/* tests/allure/results/history/. && allure open tests/allure/reports

run-tests target='default' context='kind-kind-multi':
  helm test {{TESTS_RELEASE}} -n {{NAMESPACE}} --kube-context {{context}} --timeout {{TESTS_TIMEOUT}} --logs

build-load-run-tests: build-pytest-docker load-pytest-kind run-tests

delete-kind-cluster kind='kind-multi':
  kind delete cluster --name {{kind}}

cleanup-create-local-helm-cluster target range='..' $GOARCH=LOCAL_GOARCH: delete-kind-cluster create-kind-cluster
  just build-cloudbeat-docker-image $GOARCH
  just load-cloudbeat-image
  just deploy-tests-helm {{target}} tests/deploy/values/ci.yml {{range}}


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
