# Template tests

End-to-end checks for the cloud deployment templates that ship with cloudbeat
(CloudFormation and GCP Deployment Manager). They deploy a template, let the agent
enroll into a real Elastic Cloud stack and check that findings arrive.

Workflows:

- `.github/workflows/cloudformation-ci.yml`
- `.github/workflows/test-gcp-dm.yml`

## What is here

| File | Purpose |
|---|---|
| `create_stack.sh` | Creates the Elastic Cloud (ECH) deployment with `ecctl`, same sizes as the old Terraform `elk-stack` root, and exports the endpoints and credentials to the job. |
| `destroy_stack.sh` | Shuts the deployment down (run it in an `if: always()` step). Falls back to a lookup by name when the create step failed before exporting the ID. |
| `delete_gcp_env.sh` | Deletes the GCP Deployment Manager deployments after a run. |
| `fleet_setup.py` | Creates the agent and package policies in Kibana and writes the config file of the template deployment (`cnvm`, `cspm-gcp`), or waits until the agent is enrolled (`wait-agents`). |
| `test_fleet_setup.py` | Offline unit tests of `fleet_setup.py` (no Kibana, AWS or GCP needed). Run in the `Unit Test` job of `ci-pull_request.yml`. |
| `fixtures/` | Package metadata and golden package policies used by the unit tests. |

`ecctl` is installed by `.github/actions/ecctl` (pinned to a commit hash). The scripts read
`EC_API_KEY` from the environment; only the create and destroy steps set it.

## This repository is public

Endpoints and credentials of the deployment must never reach the job log or an artifact:

- `create_stack.sh` keeps ecctl's JSON output (it contains the generated credentials) in a
  `0600` temp file that is deleted on exit, and masks the password, the Kibana and
  Elasticsearch URLs (full and bare host) and the credential-bearing URLs with `::add-mask::`
  before exporting anything to `GITHUB_ENV`. The `elastic` username is not masked, because
  that would redact every log line containing the word.
- Do not add `set -x`, `ecctl --verbose` or `echo` of these values, and do not upload the
  ecctl output as an artifact.
- Secrets are set only on the steps that need them, not for the whole job.

## fleet_setup.py

Run from the `tests` directory with the Kibana settings in the environment (`create_stack.sh` exports them to the job):

```bash
cd tests
poetry run python -m template_tests.fleet_setup cnvm        # AWS CNVM: deploy/cloudformation/config.json + published template
poetry run python -m template_tests.fleet_setup cspm-gcp    # GCP CSPM: deploy/deployment-manager/config.json
poetry run python -m template_tests.fleet_setup wait-agents # waits for the agent of the last created policy
```

- `cnvm` downloads the CloudFormation template published with the integration, adds the AWS tags
  (`division`, `org`, `team`, `project`) to the EC2 instance, and keeps the repository copy as
  `elastic-agent-ec2-cnvm-orig.yml`. If the package has no `templateURL`, the repository template is used.
- `cspm-gcp` reads the credentials from `SERVICE_ACCOUNT_JSON_PATH` or `GOOGLE_APPLICATION_CREDENTIALS`.
- The enrollment token is masked with `::add-mask::` before it is written, and the config files and the state file
  (`.fleet_state.json`) are created with `0600` permissions. Do not log the Fleet or config data.
- Unit tests: `poetry run python -m pytest --noconftest -o addopts="" template_tests/test_fleet_setup.py`.

## Used by the template tests (do not remove)

- `tests/commonlib/kibana_api.py`, the Kibana client used by `fleet_setup.py`.
- `tests/` pytest suites selected with `-k cspm_gcp` and `-k cspm_aws`.
- `deploy/cloudformation/`, `deploy/deployment-manager/`.
