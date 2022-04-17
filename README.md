# Cloudbeat 
![Coverage](https://img.shields.io/badge/Coverage-48.3%25-yellow)
[![Go Report Card](https://goreportcard.com/badge/github.com/elastic/cloudbeat)](https://goreportcard.com/report/github.com/elastic/cloudbeat)
[![Build Status](https://internal-ci.elastic.co/buildStatus/icon?job=cloudbeat%2Fcloudbeat-mbp%2Fmain)](https://internal-ci.elastic.co/job/cloudbeat/job/cloudbeat-mbp/job/main/)

## Table of contents
- [Prerequisites](#prerequisites)
- [Running Cloudbeat](#running-cloudbeat)
  - [Clean up](#clean-up)
  - [Remote Debugging](#remote-debugging)
- [Code guidelines](#code-guidelines)


## Prerequisites
1. [Just command runner](https://github.com/casey/just)
2. Elasticsearch with the default username & password (`elastic` & `changeme`) running on the default port (`http://localhost:9200`)
3. Kibana with running on the default port (`http://localhost:5601`)
4. Set up the local env:

```zsh
just setup-env
```

## Running Cloudbeat

Build & deploy cloudbeat:

```zsh
just build-deploy-cloudbeat
```

To validate check the logs:

```zsh
just logs-cloudbeat
```

Now go and check out the data on your Kibana! Make sure to add a kibana dataview `logs-cis_kubernetes_benchmark.findings-*`

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

## Code guidelines

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