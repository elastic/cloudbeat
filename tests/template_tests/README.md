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

## Used by the template tests (keep when removing the old environment tooling)

- `tests/integrations_setup/`: `install_cnvm_integration.py`, `install_cspm_gcp_integration.py`,
  `agents_enrolled.py`, and the `configuration_fleet`, `fleet_api`, `package_policy`,
  `state_file_manager` code they import.
- `tests/` pytest suites selected with `-k cspm_gcp` and `-k cspm_aws`.
- `deploy/cloudformation/`, `deploy/deployment-manager/`.
