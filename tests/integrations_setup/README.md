# Template test integration installers

Small installers used by the CloudFormation and GCP Deployment Manager template tests
(`.github/workflows/cloudformation-ci.yml`, `.github/workflows/test-gcp-dm.yml`). Each one creates the agent and
package policies in Kibana and writes the config file consumed by the template deployment.

| Script | Purpose |
|--------|---------|
| `install_cnvm_integration.py` | Installs CNVM and writes `deploy/cloudformation/config.json` |
| `install_cspm_gcp_integration.py` | Installs CSPM GCP and writes `deploy/deployment-manager/config.json` |
| `agents_enrolled.py` | Waits until the expected agents are enrolled |

Supporting code: `configuration_fleet.py` (environment variables), `package_policy.py`, `state_file_manager.py`,
and the `fleet_api` package.

## Prerequisites

Set the common variables: `ES_USER`, `ES_PASSWORD`, `KIBANA_URL`, `STACK_VERSION`. Integration-specific variables
are listed in [configuration_fleet](./configuration_fleet.py).

## Usage

```bash
cd tests
poetry install
poetry run python ./integrations_setup/install_cnvm_integration.py
poetry run python ./integrations_setup/install_cspm_gcp_integration.py
```
