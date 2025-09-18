# Variables
CLOUDBEAT_VERSION := ''
kustomizeVanillaOverlay := "deploy/kustomize/overlays/cloudbeat-vanilla"
kustomizeVanillaNoCertOverlay := "deploy/kustomize/overlays/cloudbeat-vanilla-nocert"
kustomizeEksOverlay := "deploy/kustomize/overlays/cloudbeat-eks"
kustomizeAwsOverlay := "deploy/kustomize/overlays/cloudbeat-aws"
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
  docker build -f deploy/Dockerfile -t cloudbeat . --platform=linux/$GOARCH

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
  GOOS=linux CGO_ENABLED=0 go build -gcflags "all=-N -l" && docker build -f deploy/Dockerfile.debug -t cloudbeat . --platform=linux/$GOARCH

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


#### GENERAL ####

logs-cloudbeat:
  CLOUDBEAT_POD=$( kubectl get pods -o=name -n kube-system | grep -m 1 "cloudbeat" ) && \
  kubectl logs -f "${CLOUDBEAT_POD}" -n kube-system

deploy-arm:
  deploy/azure/generate_dev_template.py --deploy

deploy-cloudformation:
  cd deploy/cloudformation && go run .

deploy-asset-inventory-cloudformation:
  cd deploy/asset-inventory-cloudformation && go run .

deploy-dm:
  .deploy/deployment-manager/deploy.sh

delete-dm name:
  gcloud deployment-manager deployments delete {{name}} -q

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


delete-cloud-env prefix ignore-prefix="" interactive="true":
  # delete all cloud environments that start with {{prefix}} and do not start with {{ignore-prefix}}
  # ask for confirmation before deleting each environment: {{interactive}}
  cd deploy/test-environments && \
  terraform init && \
  pwd && ./delete_env.sh --prefix {{prefix}} --ignore-prefix '{{ignore-prefix}}' --interactive {{interactive}}


#### MOCKS #####

# generate new and update existing mocks from golang interfaces
# and update the license header
generate-mocks:
  mockery --config=.mockery.yaml
  mage AddLicenseHeaders

# run to validate no mocks are missing
validate-mocks:
  # delete and re-generate files to check nothing is different / missing
  find . -name '*mock.go' -exec rm -f {} \;
  just generate-mocks
  git diff --exit-code
  git ls-files --exclude-standard --others | grep -qE 'mock_.*go' && exit 1 || exit 0

#### TESTS ####

TESTS_RELEASE := 'cloudbeat-test'
TIMEOUT := '1200s'
TESTS_TIMEOUT := '60m'
ELK_STACK_VERSION := env_var('ELK_VERSION')
NAMESPACE := 'kube-system'
ECR_CLOUDBEAT_TEST := 'public.ecr.aws/z7e1r9l0/'

patch-cb-yml-tests:
  kubectl kustomize deploy/k8s/kustomize/test > tests/test_environments/cloudbeat-pytest.yml

build-pytest-docker:
  cd tests; docker build -t {{TESTS_RELEASE}} .

load-pytest-kind kind='kind-multi': build-pytest-docker
  kind load docker-image {{TESTS_RELEASE}}:latest --name {{kind}}

load-pytest-eks:
  aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin public.ecr.aws/z7e1r9l0
  docker tag {{TESTS_RELEASE}}:latest {{ECR_CLOUDBEAT_TEST}}{{TESTS_RELEASE}}:latest
  docker push {{ECR_CLOUDBEAT_TEST}}{{TESTS_RELEASE}}:latest

deploy-tests-helm target values_file='tests/test_environments/values/ci.yml' range='':
  helm upgrade --wait --timeout={{TIMEOUT}} --install --values {{values_file}} --set testData.marker='{{target}}' --set testData.range={{range}} --set elasticsearch.imageTag={{ELK_STACK_VERSION}} --set kibana.imageTag={{ELK_STACK_VERSION}} --namespace={{NAMESPACE}} {{TESTS_RELEASE}} tests/test_environments/k8s-cloudbeat-tests/

apply-k8s-test-objects:
  kubectl apply -f tests/test_environments/k8s-objects-all-cases.yml && kubectl wait --for=condition=Ready --timeout={{TIMEOUT}} pod --all -n {{NAMESPACE}}

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
  just deploy-tests-helm {{target}} tests/test_environments/values/ci.yml {{range}}
