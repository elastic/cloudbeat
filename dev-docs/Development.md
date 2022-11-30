# Development

### Code guidelines
For Golang, we try to follow [Google's code style](https://google.github.io/styleguide/go/)
For Python we try to follow [PEP8](https://peps.python.org/pep-0008/) style guid

### Pre-commit hooks

We use [pre-commit](https://pre-commit.com/) package to enforce our pre commit hooks.
To install:

```zsh
brew install pre-commit # Install the package
pre-commit install # install the pre commits hooks
pre-commit run --all-files --verbose # run it!
```

### Update Cloudbeat configuration on a running Elastic-Agent
Update cloudbeat configuration on a running elastic-agent can be done by running the [script](/scripts/remote_edit_config.sh).
The script still requires a second step of triggering the agent to re-run cloudbeat.
This can be done on Fleet UI by changing the agent log level.

### Local configuration changes
To update your local configuration of cloudbeat, use `mage config`, for example to control the policy type you can pass the following environment variable
```zsh
POLICY_TYPE=cloudbeat/cis_eks mage config
```

The default `POLICY_TYPE` is set to `cloudbeat/cis_k8s` on [`_meta/config/cloudbeat.common.yml.tmpl`](_meta/config/cloudbeat.common.yml.tmpl)

### Testing

Cloudbeat has a various sets of tests. This guide should help to understand how the different test suites work, how they are used and how new tests are added.

In general there are two major test suites:

- Unit tests written in Go
- Integration tests written in Python (using pytest)

The tests written in Go use the Go Testing package. The tests written in Python depend on pytest and require a compiled and executable binary from the Go code. The python test run a beat with a specific config and params and either check if the output is as expected or if the correct things show up in the logs.

Integration tests in Beats are tests which require an external system like Elasticsearch to test if the integration with this service works as expected. Beats provides in its testsuite docker containers and docker-compose files to start these environments but a developer can run the required services also locally.

For more information, see our [testing docs](/tests/README.md)

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
