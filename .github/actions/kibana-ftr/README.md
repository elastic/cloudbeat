# Run Kibana FTR

This GitHub Action runs Kibana tests using the Functional Test Runner ([FTR](https://www.elastic.co/guide/en/kibana/current/development-tests.html#development-functional-tests)).

## Inputs

| Name                 | Description                                    | Required | Default                                                      |
|----------------------|------------------------------------------------|----------|--------------------------------------------------------------|
| `test_kibana_url`    | URL for the Kibana instance to test            | true     |                                                              |
| `test_es_url`        | URL for the Elasticsearch instance to test     | true     |                                                              |
| `es_version`         | Version of Elasticsearch to test against       | true     |                                                              |
| `kibana_ref`     | Kibana PR commit sha                           | false    | `main`                                                       |

## Usage

```yaml
name: Kibana UI Tests

on:
  workflow_dispatch:
    inputs:
      kibana_ref:
        type: string
        description: |
          Kibana PR commit sha
        required: false

jobs:
  run-kibana-tests:
    runs-on: ubuntu-22.04
    steps:
      - name: Check out the repo
        uses: actions/checkout@v4

      - name: Run Kibana Tests Action
        uses: ./.github/actions/kibana-ftr
        with:
          kibana_ref: ${{ github.event.inputs.kibana_ref || 'main' }}
          test_kibana_url: ${{ secrets.TEST_KIBANA_URL }}
          test_es_url: ${{ secrets.TEST_ES_URL }}
          es_version: ${{ env.ES_VERSION }}
```

## Details

This action performs the following steps:

1. **Set global variables**:
   - Sets the global variable `KIBANA_DIR` to `kibana`.

2. **Checkout Kibana Repository**:
   - Uses the `actions/checkout@v4` action to check out the Kibana repository at the specified `kibana_ref` or `main` to the `kibana` directory.

3. **Setup Node**:
   - Uses the `actions/setup-node@v4` action to set up the Node.js environment based on the `package.json`.

4. **Bootstrap Kibana**:
   - Runs the `yarn kbn bootstrap` command in the `kibana` directory to bootstrap the Kibana environment.

5. **Run FTR**:
   - Runs the Functional Test Runner (FTR) using the `x-pack/test/cloud_security_posture_functional/config.cloud.ts` configuration and the necessary environment variables.

## Notes

- Ensure that the `test_kibana_url` and `test_es_url` inputs are provided and valid. More information about these variables can be found [here](https://www.elastic.co/guide/en/kibana/current/development-tests.html#_running_functional_tests).
- The `kibana_ref` input can be omitted to default to the `main` branch.
