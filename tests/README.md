# Cloudbeat Automated Tests Framework (C-ATF)

>

This project provides a framework for developing component and integration tests for Cloudbeat.

## Getting started

This guide provides installations for macOS users via [Homebrew](https://brew.sh/).
For other platforms, please go forward through instructions provided in links and select relevant installation:

Install [Allure](https://docs.qameta.io/allure/#_installing_a_commandline) commandline tool

```shell
brew install allure
```

After install allure commandline you will be able to display raw test results in pretty html format.

## System Under Test (SUT) Setup

Before performing steps below verify that **just** tool is installed and your root folder in commandline is cloudbeat.

1. Create Kubernetes cluster

    ```shell
    just create-kind-cluster
    ```

2. Build cloudbeat and upload docker image to kind

    ```shell
    just build-cloudbeat
    just load-cloudbeat-image
    ```

3. Install elasticsearch and start cloudbeat

    ```shell
    just deploy-tests-helm-ci pre_merge
    ```

This command will install elasticsearch one node instance in kubernetes cluster and prepare configuration
for executing **pre_merge** marker tests.

### Dependencies

Automated test framework is built on top of the following:

- [pytest](https://docs.pytest.org/en/7.1.x/) -  Python testing framework.
- [poetry](https://python-poetry.org/docs/) - Dependency management for Python.
- [allure](https://docs.qameta.io/allure/#_pytest) - Generates html test reports.

### Prerequisites

Install [poetry](https://python-poetry.org/docs/#installation) python package manager

```shell
curl -sSL https://raw.githubusercontent.com/python-poetry/poetry/master/get-poetry.py | python -
```

**Note**: Installing poetry process will also initiate python installation if not installed yet.

### Setting up Dev

After forking the repo:

```shell
git clone https://github.com/<yourname>/cloudbeat.git
cd cloudbeat/tests
poetry install
```

### Configuring IDE

Tests development may be done using any IDE like [visual studio code](https://code.visualstudio.com/)
or [pycharm](https://www.jetbrains.com/pycharm/).

#### Configure Interpreter

Pycharm has integration with poetry. We recommend using it as interpreter for tests development:

- Navigate to PyCharm -> Preferences -> Project -> Project Interpreter
- Select Gear icon near Python interpreter text box
- Select Poetry environment

#### Tests Project as Root

Since this test project is a standalone project in cloudbeat repo, in pycharm you can configure it as standalone repo:

- PyCharm -> Preferences -> Project -> Project Structure
- Write **Project Structure** in the search box
- Select tests folder as content root

or alternatively, right-click the tests folder and select **Mark Directory as** -> **Sources Root**

#### Run / Debug Configuration:

- Open Run/Debug Configuration dialog
- Select pytest as runner
- For target select script path radio button and select script to be executed
- In Additional arguments define: -s -v --alluredir=./reports
- In Python interpreter select Poetry

### Project Structure

The project main folders are:

- commonlib - contains helper functions used for tests development.
- deploy - contains helm charts for deploying ELK, cloudbeat, and tests docker.
- product - contains cloudbeat tests, for example cloudbeat behavior and functional tests.
- integration - contains cloudbeat integration tests.
- project root content - contains project and tests configuration files.

### Adding New Tests

#### Conventions

- Test file shall start with **test_** prefix, for example test_login.py.
- Test method shall start with **test_** prefix, for example test_login_success().
- Add test marker for a test method. Framework markers are defined in **pyproject.toml** file,
  section **[tool.pytest.ini_options] markers**
- SetUp and TearDown actions for a test method are defined using pytest.fixture.
- Global SetUp and TearDown actions are defined in **conftest.py** file.

#### Test Folders

- Product tests folder is **product/tests**.</br>
- Integration tests folder is **intergration/tests**.

### Building

Test framework output is a docker image that encapsulates python framework and tests.
In order to build tests docker image ensure that docker desktop application is running.

```shell
cd tests
docker build -t cloudbeat-tests .
```

The command above will build docker image tagged as **cloudbeat-tests**, '.' - means search for **Dockerfile** in the
current folder.

Execute the following command

```shell
docker images
```

The cloudbeat-tests:latest image shall appear in docker images list.

### Uploading Image

For loading test's docker image to kind cluster execute

```shell
just load-pytest-kind
```

## Tests Execution

Tests execution depends on the developers needs and currently this framework supports the following modes:

1. Dev Mode - Writing test and executing tests on dev machine
2. Integration Mode (Production) - Writing tests on dev machine, building test's docker image, and executing tests in
   kubernetes cluster.

### Dev Mode

To run all test targets with just cloudbeat, without testing against Kibana or Elasticsearch, run

```
just run-test-targets
```

Note that this will create and destroy the test cluster several times. Logs can be found in the `test-logs` directory
and test results can be found in `tests/allure/results`.

----

Before running tests verify that **System Under Test (SUT) Setup** is done and running.
Since elasticsearch is deployed inside cluster, for reaching it from outside execute the following command:

```shell
kubectl port-forward svc/elasticsearch-master -n kube-system 9200
```

For kibana:

```shell
kubectl port-forward svc/cloudbeat-test-kibana -n kube-system 5601
```

IDE: Execute tests by IDE runner (assume run/debug setup done before) -> click on play/debug button
Terminal: Ensure that virtualenv is activated and then execute

```shell
poetry run pytest -s -v -m ci_cloudbeat --alluredir=./reports
allure serve ./reports
```

### Integration Mode

1. Build tests docker image and upload to kubernetes cluster

    ```shell
    just build-test-docker
    just load-tests-image-kind
    ```

2. If test suite is not deployed initiate:

    ```shell
    just deploy-tests-helm-ci pre_merge
    ```

3. Execute tests

    ```shell
    just run-tests
    ```

4. Investigate reports

    ```shell
    allure serve <path to raw result data>
    ```

## CI-CD Workflows

Current usage of test project is in the following ci flows:

- cloudbeat-ci
    - build cloudbeat-tests docker image
    - load cloudbeat-tests docker image to kind cluster
    - deploy tests helm chart
    - execute product and integration cloudbeat tests

