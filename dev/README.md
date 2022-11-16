## Code guidelines

### Pre-commit hooks

see [pre-commit](https://pre-commit.com/) package

```zsh
brew install pre-commit # Install the package
pre-commit install # install the pre commits hooks
pre-commit run --all-files --verbose # run it!
```

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
