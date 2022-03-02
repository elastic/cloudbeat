# BEAT_NAME=cloudbeat
# BEAT_PATH=github.com/elastic/beats/v7/cloudbeat
# BEAT_GOPATH=$(firstword $(subst :, ,${GOPATH}))
# SYSTEM_TESTS=false
# TEST_ENVIRONMENT=false
# ES_BEATS_IMPORT_PATH=github.com/elastic/beats/v7
# ES_BEATS?=$(shell go list -m -f '{{.Dir}}' ${ES_BEATS_IMPORT_PATH})
# LIBBEAT_MAKEFILE=$(ES_BEATS)/libbeat/scripts/Makefile
# GOPACKAGES=$(shell go list ${BEAT_PATH}/... | grep -v /tools)
# GOBUILD_FLAGS=-i -ldflags "-X ${ES_BEATS_IMPORT_PATH}/libbeat/version.buildTime=$(NOW) -X ${ES_BEATS_IMPORT_PATH}/libbeat/version.commit=$(COMMIT_ID)"
# MAGE_IMPORT_PATH=github.com/magefile/mage
# NO_COLLECT=true
# CHECK_HEADERS_DISABLED=true

# # Path to the libbeat Makefile
# -include $(LIBBEAT_MAKEFILE)

# .PHONY: copy-vendor
# copy-vendor:
# 	mage vendorUpdate

# delete-pod:
# 	kubectl delete pod cloudbeat-demo

# build-docker:
# 	GOOS=linux go build && docker build -t cloudbeat .

# docker-image-load-minikube: build-docker
# 	minikube image load cloudbeat:latest

# docker-image-load-kind: build-docker
# 	kind load docker-image docker.elastic.co/beats/elastic-agent:8.1.0-SNAPSHOT --name single-host

# deploy-cloudbeat:
# 	kubectl apply -f deploy/k8s/cloudbeat-ds.yaml -n kube-system

# deploy-pod: delete-pod build-docker docker-image-load-minikube
# 	kubectl apply -f pod.yml

# build-deploy-docker: build-docker docker-image-load-kind deploy-cloudbeat

##############################################################################
# Variables used for various build targets.
##############################################################################

# Ensure the Go version in .go_version is installed and used.
GOROOT?=$(shell ./script/run_with_go_ver go env GOROOT)
GO:=$(GOROOT)/bin/go
export PATH:=$(GOROOT)/bin:$(PATH)

# By default we run tests with verbose output. This may be overridden, e.g.
# scripts may set GOTESTFLAGS=-json to format test output for processing.
GOTESTFLAGS?=-v

GOOSBUILD:=./build/$(shell $(GO) env GOOS)
APPROVALS=$(GOOSBUILD)/approvals
GENPACKAGE=$(GOOSBUILD)/genpackage
GOIMPORTS=$(GOOSBUILD)/goimports
GOLICENSER=$(GOOSBUILD)/go-licenser
GOLINT=$(GOOSBUILD)/golint
MAGE=$(GOOSBUILD)/mage
REVIEWDOG=$(GOOSBUILD)/reviewdog
STATICCHECK=$(GOOSBUILD)/staticcheck
ELASTICPACKAGE=$(GOOSBUILD)/elastic-package

PYTHON_ENV?=.
PYTHON_BIN:=$(PYTHON_ENV)/build/ve/$(shell $(GO) env GOOS)/bin
PYTHON=$(PYTHON_BIN)/python

APM_SERVER_VERSION=$(shell grep defaultBeatVersion cmd/version.go | cut -d'=' -f2 | tr -d '" ')

# Create a local config.mk file to override configuration,
# e.g. for setting "GOLINT_UPSTREAM".
-include config.mk

##############################################################################
# Rules for building and unit-testing apm-server.
##############################################################################

.DEFAULT_GOAL := cloudbeat

.PHONY: build-cloudbeat
apm-server:
	@$(GO) build -o $@ .

.PHONY: apm-server-oss
apm-server-oss:
	@$(GO) build -o $@

.PHONY: test
test:
	$(GO) test $(GOTESTFLAGS) ./...

.PHONY: system-test
system-test:
	@(cd systemtest; $(GO) test $(GOTESTFLAGS) -timeout=20m ./...)

.PHONY:
clean: $(MAGE)
	@$(MAGE) clean

##############################################################################
# Checks/tests.
##############################################################################

.PHONY: check-full
# check-full: update check golint staticcheck check-docker-compose
check-full: update

.PHONY: check-approvals
check-approvals: $(APPROVALS)
	@$(APPROVALS)

.PHONY: check
check: $(MAGE) check-fmt check-headers
	@$(MAGE) check

.PHONY: bench
bench:
	@$(GO) test -benchmem -run=XXX -benchtime=100ms -bench='.*' ./...

##############################################################################
# Rules for updating config files, etc.
##############################################################################

#update: go-generate add-headers build-package notice $(MAGE)
update: go-generate add-headers $(MAGE)
	@$(MAGE) update
	@go mod download all # make sure go.sum is complete

config: apm-server.yml apm-server.docker.yml
apm-server.yml apm-server.docker.yml: $(MAGE) magefile.go _meta/beat.yml
	@$(MAGE) config

.PHONY: go-generate
go-generate:
	@$(GO) generate .

notice: NOTICE.txt
NOTICE.txt: $(PYTHON) go.mod utils/go.mod
	@$(PYTHON) script/generate_notice.py . ./x-pack/apm-server

.PHONY: add-headers
add-headers: $(GOLICENSER)
ifndef CHECK_HEADERS_DISABLED
	@$(GOLICENSER)
#	@$(GOLICENSER) -license Elasticv2 x-pack
endif

## get-version : Get the apm server version
.PHONY: get-version
get-version:
	@echo $(APM_SERVER_VERSION)

##############################################################################
# Documentation.
##############################################################################

.PHONY: docs
docs:
	@rm -rf build/html_docs
	sh script/build_apm_docs.sh apm-server docs/index.asciidoc build

.PHONY: update-beats-docs
update-beats-docs: $(PYTHON)
	@$(PYTHON) script/copy-docs.py

##############################################################################
# Beats synchronisation.
##############################################################################

BEATS_VERSION?=main
BEATS_MODULE:=$(shell $(GO) list -m -f {{.Path}} all | grep github.com/elastic/beats)

.PHONY: update-beats
update-beats: update-beats-module update
	@echo --- Use this commit message: Update to elastic/beats@$(shell $(GO) list -m -f {{.Version}} $(BEATS_MODULE) | cut -d- -f3)

.PHONY: update-beats-module
update-beats-module:
	$(GO) get -d -u $(BEATS_MODULE)@$(BEATS_VERSION) && $(GO) mod tidy
	cp -f $$($(GO) list -m -f {{.Dir}} $(BEATS_MODULE))/.go-version .go-version
	find . -maxdepth 2 -name Dockerfile -exec sed -i'.bck' -E -e "s#(FROM golang):[0-9]+\.[0-9]+\.[0-9]+#\1:$$(cat .go-version)#g" {} \;
	sed -i'.bck' -E -e "s#(:go-version): [0-9]+\.[0-9]+\.[0-9]+#\1: $$(cat .go-version)#g" docs/version.asciidoc

##############################################################################
# Linting, style-checking, license header checks, etc.
##############################################################################

GOLINT_TARGETS?=$(shell $(GO) list ./...)
GOLINT_UPSTREAM?=origin/main
REVIEWDOG_FLAGS?=-conf=reviewdog.yml -f=golint -diff="git diff $(GOLINT_UPSTREAM)"
GOLINT_COMMAND=$(GOLINT) ${GOLINT_TARGETS} | grep -v "should have comment" | $(REVIEWDOG) $(REVIEWDOG_FLAGS)

.PHONY: golint
golint: $(GOLINT) $(REVIEWDOG)
	@output=$$($(GOLINT_COMMAND)); test -z "$$output" || (echo $$output && exit 1)

.PHONY: staticcheck
staticcheck: $(STATICCHECK)
	$(STATICCHECK) github.com/elastic/apm-server/...

.PHONY: check-changelogs
check-changelogs: $(PYTHON)
	$(PYTHON) script/check_changelogs.py

.PHONY: check-headers
check-headers: $(GOLICENSER)
ifndef CHECK_HEADERS_DISABLED
	@$(GOLICENSER) -d -exclude build -exclude x-pack -exclude internal/otel_collector
	@$(GOLICENSER) -d -exclude build -license Elasticv2 x-pack
endif

.PHONY: check-docker-compose
check-docker-compose: $(PYTHON_BIN)
	@PATH=$(PYTHON_BIN):$(PATH) ./script/check_docker_compose.sh $(BEATS_VERSION)

.PHONY: format-package build-package
format-package: $(ELASTICPACKAGE)
	@(cd apmpackage/apm; $(CURDIR)/$(ELASTICPACKAGE) format)
build-package: $(ELASTICPACKAGE)
	@rm -fr ./build/integrations/apm/* ./build/apmpackage
	@$(GO) run ./apmpackage/cmd/genpackage -o ./build/apmpackage -version=$(APM_SERVER_VERSION)
	@(cd ./build/apmpackage; $(CURDIR)/$(ELASTICPACKAGE) build && $(CURDIR)/$(ELASTICPACKAGE) check)

.PHONY: check-gofmt check-autopep8 gofmt autopep8
check-fmt: check-gofmt check-autopep8
fmt: gofmt autopep8
check-gofmt: $(GOIMPORTS)
	@PATH=$(GOOSBUILD):$(PATH) sh script/check_goimports.sh
gofmt: $(GOIMPORTS) add-headers
	@echo "fmt - goimports: Formatting Go code"
	@PATH=$(GOOSBUILD):$(PATH) GOIMPORTSFLAGS=-w sh script/goimports.sh
check-autopep8: $(PYTHON_BIN)
	@PATH=$(PYTHON_BIN):$(PATH) sh script/autopep8_all.sh --diff --exit-code
autopep8: $(PYTHON_BIN)
	@echo "fmt - autopep8: Formatting Python code"
	@PATH=$(PYTHON_BIN):$(PATH) sh script/autopep8_all.sh --in-place

##############################################################################
# Rules for creating and installing build tools.
##############################################################################

BIN_MAGE=$(GOOSBUILD)/bin/mage

# BIN_MAGE is the standard "mage" binary.
$(BIN_MAGE): go.mod
	$(GO) build -o $@ github.com/magefile/mage

# MAGE is the compiled magefile.
$(MAGE): magefile.go $(BIN_MAGE)
	$(BIN_MAGE) -compile=$@

$(GOLINT): go.mod
	$(GO) build -o $@ golang.org/x/lint/golint

$(GOIMPORTS): go.mod
	$(GO) build -o $@ golang.org/x/utils/cmd/goimports

$(STATICCHECK): utils/go.mod
	$(GO) build -o $@ -modfile=$< honnef.co/go/utils/cmd/staticcheck

$(GOLICENSER): utils/go.mod
	$(GO) build -o $@ -modfile=$< github.com/elastic/go-licenser

$(REVIEWDOG): utils/go.mod
	$(GO) build -o $@ -modfile=$< github.com/reviewdog/reviewdog/cmd/reviewdog

$(ELASTICPACKAGE): utils/go.mod
	$(GO) build -o $@ -modfile=$< -ldflags '-X github.com/elastic/elastic-package/internal/version.CommitHash=anything' github.com/elastic/elastic-package

$(PYTHON): $(PYTHON_BIN)
$(PYTHON_BIN): $(PYTHON_BIN)/activate
$(PYTHON_BIN)/activate: $(MAGE)
	@$(MAGE) pythonEnv
	@touch $@

.PHONY: $(APPROVALS)
$(APPROVALS):
	@$(GO) build -o $@ github.com/elastic/apm-server/approvaltest/cmd/check-approvals

##############################################################################
# Release manager.
##############################################################################

# Builds a snapshot release.
release-manager-snapshot: export SNAPSHOT=true
release-manager-snapshot: release

# Builds a snapshot release.
.PHONY: release-manager-release
release-manager-release: release

.PHONY: release

JAVA_ATTACHER_VERSION:=1.28.4
JAVA_ATTACHER_JAR:=apm-agent-attach-cli-$(JAVA_ATTACHER_VERSION)-slim.jar
JAVA_ATTACHER_SIG:=$(JAVA_ATTACHER_JAR).asc
JAVA_ATTACHER_BASE_URL:=https://repo1.maven.org/maven2/co/elastic/apm/apm-agent-attach-cli
JAVA_ATTACHER_URL:=$(JAVA_ATTACHER_BASE_URL)/$(JAVA_ATTACHER_VERSION)/$(JAVA_ATTACHER_JAR)
JAVA_ATTACHER_SIG_URL:=$(JAVA_ATTACHER_BASE_URL)/$(JAVA_ATTACHER_VERSION)/$(JAVA_ATTACHER_SIG)

APM_AGENT_JAVA_PUB_KEY:=apm-agent-java-public-key.asc

release: export PATH:=$(dir $(BIN_MAGE)):$(PATH)
release: $(MAGE) $(PYTHON) build/$(JAVA_ATTACHER_JAR) build/dependencies.csv
	$(MAGE) package

build/dependencies.csv: $(PYTHON) go.mod
	$(PYTHON) script/generate_notice.py ./x-pack/apm-server --csv $@

.imported-java-agent-pubkey:
	@gpg --import $(APM_AGENT_JAVA_PUB_KEY)
	@touch $@

build/$(JAVA_ATTACHER_SIG):
	curl -sSL $(JAVA_ATTACHER_SIG_URL) > $@

build/$(JAVA_ATTACHER_JAR): build/$(JAVA_ATTACHER_SIG) .imported-java-agent-pubkey
	curl -sSL $(JAVA_ATTACHER_URL) > $@
	gpg --verify $< $@
	@cp $@ build/java-attacher.jar
