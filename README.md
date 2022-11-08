# Cloudbeat
[![Coverage Status](https://coveralls.io/repos/github/elastic/cloudbeat/badge.svg?branch=main)](https://coveralls.io/github/elastic/cloudbeat?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/elastic/cloudbeat)](https://goreportcard.com/report/github.com/elastic/cloudbeat)
[![Build Status](https://internal-ci.elastic.co/buildStatus/icon?job=cloudbeat%2Fcloudbeat-mbp%2Fmain)](https://internal-ci.elastic.co/job/cloudbeat/job/cloudbeat-mbp/job/main/)

### Cloudbeat evaluates cloud assets for security compliance and ships findings to Elasticsearch

## Table of contents
- [Prerequisites](#prerequisites)
- [Running Cloudbeat](#running-cloudbeat)
  - [Clean up](#clean-up)
  - [Remote Debugging](#remote-debugging)
  - [Skaffold Workflows](#skaffold-workflows)
- [Code guidelines](#code-guidelines)


## Prerequisites
1. [Hermit by Cashapp](https://cashapp.github.io/hermit/usage/get-started/)
2. Elasticsearch with the default username & password (`elastic` & `changeme`) running on the default port (`http://localhost:9200`)
3. Kibana with running on the default port (`http://localhost:5601`)
4. Install and configure [Elastic-Package](https://github.com/elastic/elastic-package)
5. Set up the local env:

- Install & activate hermit
  ```zsh
  curl -fsSL https://github.com/cashapp/hermit/releases/download/stable/install.sh | /bin/bash
  ```
	```zsh
  . ./bin/activate-hermit
  ```
- Run setup env recipe
  ```zsh
  just setup-env
  ```


>**Note**
>This will download and install hermit into `~/bin`. You should add this to your `$PATH` if it isn't already. Also consider to review documentation for automatic shell & IDE integration for your setup of choice.


## Running Cloudbeat
Load the elastic stack environment variables.
```zsh
eval "$(elastic-package stack shellinit)"
```

### Kubernetes Vanilla
Build & deploy cloudbeat:

```zsh
just build-deploy-cloudbeat
```

### Amazon Elastic Kubernetes Service (EKS)
Export AWS creds as env vars, kustomize will use these to populate your cloudbeat deployment.
```zsh
$ export AWS_ACCESS_KEY="<YOUR_AWS_KEY>" AWS_SECRET_ACCESS_KEY="<YOUR_AWS_SECRET>"
```

Set your default cluster to your EKS cluster
```zsh
 kubectl config use-context your-eks-cluster
```

Deploy cloudbeat on your EKS cluster
```zsh
just deploy-eks-cloudbeat
````
### Advanced

If you need to change the default values in the configuration(ES_HOST, ES_PORT, ES_USERNAME, ES_PASSWORD), you can
also create the deployment file yourself.

Vanilla
```zsh
just create-vanilla-deployment-file
```

EKS
```zsh
just create-eks-deployment-file
```

To validate check the logs:

### See logs
```zsh
just logs-cloudbeat
```

Now go and check out the data on your Kibana!

### Clean up

To stop this example and clean up the pod, run:
```zsh
just delete-cloudbeat
```
### Remote Debugging

Build & Deploy remote debug docker:

```zsh
just build-deploy-cloudbeat-debug
```

After running the pod, expose the relevant ports:
```zsh
just expose-ports
```

The app will wait for the debugger to connect before starting

```zsh
just logs-cloudbeat
```

Use your favorite IDE to connect to the debugger on `localhost:40000` (for example [Goland](https://www.jetbrains.com/help/go/attach-to-running-go-processes-with-debugger.html#step-3-create-the-remote-run-debug-configuration-on-the-client-computer))

Note: Check the jusfile for all available commands for build or deploy `$ just --summary`
</br>

### Skaffold Workflows
[Skaffold](https://skaffold.dev/) is a CLI tool that enables continuous development for K8s applications. Skaffold will initiate a file-system watcher and will continuously deploy cloudbeat to a local or remote K8s cluster. The skaffold workflows are defined in the [skaffold.yml](skaffold.yml) file.
[Kustomize](https://kustomize.io/) is used to overlay different config options. (current are cloudbeat vanilla & EKS)

#### Cloudbeat Vanilla:
Skaffold will initiate a watcher to build and re-deploy Cloudbeat every time a go file is saved and output logs to stdout
```zsh
skaffold dev
```

#### Cloudbeat EKS:
Export AWS creds as env vars, Skaffold & kustomize will use these to populate your k8s deployment.
```zsh
$ export AWS_ACCESS_KEY="<YOUR_AWS_KEY>" AWS_SECRET_ACCESS_KEY="<YOUR_AWS_SECRET>"
```
A [skaffold profile](https://skaffold.dev/docs/environment/profiles/) is configured for EKS, it can be activated via the following options

Specify the profile name using the `-p` flag
```zsh
skaffold -p eks dev
```

export the activation var prior to skaffold invocation, then proceed as usual.
```zsh
export SKF_MODE="CB_EKS"
skaffold dev
```
#### Additional commands:

Skaffold supports one-off commands (no continuous watcher) if you wish to build or deploy just once.
```zsh
skaffold build
skaffold deploy
```
Full CLI reference can be found [here](https://skaffold.dev/docs/references/cli/)
## Running Agent & Cloudbeat
Cloudbeat is only supported on managed elastic-agents. It means, that in order to run the setup, you will be required to have a Kibana running.
Create an agent policy and install the CSP integration. Now, when adding a new agent, you will get the K8s deployment instructions of elastic-agent.

### Update settings
Update cloudbeat settings on a running elastic-agent can be done by running the [script](/scripts/remote_edit_config.sh).
The script still requires a second step of trigerring the agent to re-run cloudbeat.
This can be done on Fleet UI by changing the agent log level.
Another option is through CLI on the agent by running
```
kill -9 `pidof cloudbeat`
```

### Local configuration changes
To update your local configuration of cloudbeat and control it, use 
```sh
mage config
```

In order to control the policy type you can pass the following environment variable
```sh
POLICY_TYPE=cloudbeat/cis_eks mage config
```

The default `POLICY_TYPE` is set to `cloudbeat/cis_k8s` on `_meta/config/cloudbeat.common.yml.tmpl`

## Code guidelines

### Pre-commit hooks

see [pre-commit](https://pre-commit.com/) package

- Install the package `brew install pre-commit`
- Then run `pre-commit install`
- Finally `pre-commit run --all-files --verbose`

### Editorconfig

see [editorconfig](https://editorconfig.org/#pre-installed) package
### Testing

Cloudbeat has a various sets of tests. This guide should help to understand how the different test suites work, how they are used and how new tests are added.

In general there are two major test suites:

- Unit tests written in Go
- Integration tests written in Python

The tests written in Go use the Go Testing package. The tests written in Python depend on pytest and require a compiled and executable binary from the Go code. The python test run a beat with a specific config and params and either check if the output is as expected or if the correct things show up in the logs.

Integration tests in Beats are tests which require an external system like Elasticsearch to test if the integration with this service works as expected. Beats provides in its testsuite docker containers and docker-compose files to start these environments but a developer can run the required services also locally.

#### Mocking

Cloudbeat uses [`mockery`](https://github.com/vektra/mockery) as its mocking test framework.
`Mockery` provides an easy way to generate mocks for golang interfaces.

Some tests use the new [expecter]((https://github.com/vektra/mockery#expecter-interfaces)) interface the library provides.
For example, given an interface such as

```go
type Requester interface {
	Get(path string) (string, error)
}
```
You can use the type-safe expecter interface as such:
```go
requesterMock := Requester{}
requesterMock.EXPECT().Get("some path").Return("result", nil)
requesterMock.EXPECT().
	Get(mock.Anything).
	Run(func(path string) { fmt.Println(path, "was called") }).
	// Can still use return functions by getting the embedded mock.Call
	Call.Return(func(path string) string { return "result for " + path }, nil)
```

Notes
- Place the test in the same package as the code it meant to test.
- File name should be aligned with the convention `original_file_mock`. For example: ecr_provider -> ecr_provider_mock.

Command example:
```
mockery --name=<interface_name> --with-expecter  --case underscore  --inpackage --recursive
```
