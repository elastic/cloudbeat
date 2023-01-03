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

### Update default configurations

If you need to change the default values in the configuration(`ES_HOST`, `ES_PORT`, `ES_USERNAME`, `ES_PASSWORD`), you
can also create the deployment file yourself:

Self-Managed Kubernetes

```zsh
just create-vanilla-deployment-file
```

Self-Managed Kubernetes wthout certificate

```zsh
just create-vanilla-deployment-file-nocert
```

EKS

```zsh
just create-eks-deployment-file
```

### Clean up

To stop this example and clean up the pod, run:

```zsh
just delete-cloudbeat
```

Or when running without certificate

```zsh
just delete-cloudbeat-nocert
```

### Remote Debugging

Build & Deploy remote debug docker:

```zsh
just build-deploy-cloudbeat-debug
```

Or without certificate

```zsh
just build-deploy-cloudbeat-debug-nocert
```

After running the pod, expose the relevant ports:

```zsh
just expose-ports
```

The app will wait for the debugger to connect before starting

> **Note**
> Use your favorite IDE to connect to the debugger on `localhost:40000` (for
> example [Goland](https://www.jetbrains.com/help/go/attach-to-running-go-processes-with-debugger.html#step-3-create-the-remote-run-debug-configuration-on-the-client-computer))

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

Some tests use the new [expecter](<(https://github.com/vektra/mockery#expecter-interfaces)>) interface the library provides.
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

### Running CI Locally - GitHub Act

#### Overview

[GitHub act tool](https://github.com/nektos/act) allows execution of GitHub actions locally.

When you `run` act it reads in your GitHub Actions from `.github/workflows/` and determines the set of actions that need to be run. It uses the Docker API to either pull or build the necessary images, as defined in your workflow files and finally determines the execution path based on the dependencies that were defined.

Full information can be found in tool's [readme docs](https://github.com/nektos/act/blob/master/README.md).

#### Prerequisites

- `act` depends on `docker` to run workflows.

#### Installation

##### [Homebrew](https://brew.sh/) (Linux/macOS)

[![homebrew version](https://img.shields.io/homebrew/v/act)](https://github.com/Homebrew/homebrew-core/blob/master/Formula/act.rb)

```shell
brew install act
```

or if you want to install version based on latest commit, you can run below (it requires compiler to be installed but Homebrew will suggest you how to install it, if you don't have it):

```shell
brew install act --HEAD
```

#### Configuration

You can provide default configuration flags to `act` by either creating a `./.actrc` or a `~/.actrc` file. Any flags in the files will be applied before any flags provided directly on the command line. For example, a file like below will always use the `nektos/act-environments-ubuntu:18.04` image for the `ubuntu-latest` runner:

```sh
# sample .actrc file
-P ubuntu-20.04=catthehacker/ubuntu:act-20.04
```

Additionally, act supports loading environment variables from an `.env` file. The default is to look in the working directory for the file but can be overridden by:

```sh
act --env-file my.env
```

`.env`:

```env
MY_ENV_VAR=MY_ENV_VAR_VALUE
MY_2ND_ENV_VAR="my 2nd env var value"
```

#### Secrets

To run `act` with secrets, you can enter them interactively, supply them as environment variables or load them from a file. The following options are available for providing secrets:

- `act -s MY_SECRET=somevalue` - use `somevalue` as the value for `MY_SECRET`.
- `act -s MY_SECRET` - check for an environment variable named `MY_SECRET` and use it if it exists. If the environment variable is not defined, prompt the user for a value.
- `act --secret-file my.secrets` - load secrets values from `my.secrets` file.
  - secrets file format is the same as `.env` format

#### Additional Environment

##### GitHub Upload Artifact

[GithHub upload-artifact](https://github.com/actions/upload-artifact) action requires simple server running on local machine.

The solution is described in [this](https://github.com/nektos/act/issues/329) issue.

In project's root folder need to create `docker-compose.yml`:

```yaml
artifact-server:
  image: ghcr.io/jefuller/artifact-server:latest
  environment:
    AUTH_KEY: foo
  ports:
    - "8080:8080"
```

Then update `.actrc`:

```
--env ACTIONS_CACHE_URL=http://localhost:8080/
--env ACTIONS_RUNTIME_URL=http://localhost:8080/
--env ACTIONS_RUNTIME_TOKEN=foo
```

Then start artifact server:

`docker-compose up`
