## About

The "Send Slack Notification" GitHub Action facilitates communication between GitHub workflows and Slack channels.
This action is designed to retrieve Vault credentials, compose Slack messages, and deliver notifications.
Working with Vault requires logging in using a GitHub token, which can be created following [these docs](https://github.com/elastic/infra/tree/master/docs/vault).


Cloudbeat secrets are stored under `secret/csp-team/ci/`. This action uses `slack-bot-token` and `slack-users` secrets.


To add a new Slack user ID, make sure that [infra-vault-tools](https://github.com/elastic/infra/blob/master/flavortown/infra-vault-tools/README.md) are installed. Add the user using the following command:

```bash
vault-append secret/csp-team/ci/slack-users <replace_by_github_user> <replace_by_slack_user_id>
```
___

- [About](#about)
- [Usage](#usage)
  - [Configuration](#configuration)
- [Customizing](#customizing)
  - [inputs](#inputs)

## Usage

This action accepts inputs such as Vault credentials, Slack channel details, and message content. Customize your Slack notifications based on your needs, whether it's a simple text message or a richer payload using Slack's formatting capabilities.

### Configuration

```yaml
---
name: example

on:
  push:
    branches:
      - main
      - "[0-9]+.[0-9]+"

jobs:
  test-slack-message:
    runs-on: ubuntu-latest
    steps:
      - name: Check out the repo
        uses: actions/checkout@v4

      - name: Publish to Slack channel
        if: always()
        uses: ./.github/actions/slack-notification
        env:
          RUN_URL: "${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}"
          JOB_STATUS_COLOR: "${{ job.status == 'success' && '#36a64f' || '#D40E0D' }}"
        with:
          vault-url: ${{ secrets.VAULT_ADDR }}
          vault-role-id: ${{ secrets.CSP_VAULT_ROLE_ID }}
          vault-secret-id: ${{ secrets.CSP_VAULT_SECRET_ID }}
          slack-channel: "#example-channel"
          slack-payload: |
            {
              "attachments": [
                {
                  "color": "${{ env.JOB_STATUS_COLOR }}",
                  "blocks": [
                    {
                        "type": "section",
                        "text": {
                            "type": "mrkdwn",
                            "text": "${{ github.workflow }} job <${{env.RUN_URL}}|${{ inputs.prefix }}> triggered by `${{github.actor}}`"
                        }
                    }
                  ]
                }
              ]
            }


```

## Customizing

### inputs

Following inputs can be used as `step.with` keys:

| Name             | Type     | Required | Description                                                       |
|------------------|----------|----------|-------------------------------------------------------------------|
| `vault-role-id`  | String   | yes      | The Vault role id.                                                |
| `vault-secret-id`| String   | yes      | The Vault secret id.                                              |
| `vault-url`      | String   | yes      | The Vault URL to connect to.                                      |
| `slack-channel`  | String   | no       | Slack channel id or channel name. Default: #cloud-sec-qa-alerts   |
| `slack-message`  | String   | no       | Posting a simple plain text message.                              |
| `slack-payload`  | String   | no       | Posting a rich message using Block Kit.                           |
| `mask-secrets`   | String   | no       | Masking secrets in the logs. Default: 'true'                      |
| `url-encoded`    | String   | no       | URL-encoded message. Default: 'true'                              |
