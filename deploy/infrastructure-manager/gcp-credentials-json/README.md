## GCP Credentials JSON (Service Account Key)

Deploy a GCP service account with JSON credentials for Elastic Agent GCP integration using GCP Infrastructure Manager.

This creates a service account with the necessary permissions and generates a JSON key file that can be used in the Elastic Agent GCP integration in Kibana.

### Prerequisites

1. GCP project with required permissions
2. `gcloud` CLI configured with your project

### Quick Deploy

#### Option 1: Cloud Shell (Recommended)

[![Open in Cloud Shell](https://gstatic.com/cloudssh/images/open-btn.svg)](https://shell.cloud.google.com/cloudshell/editor?cloudshell_git_repo=https://github.com/elastic/cloudbeat.git&cloudshell_git_branch=main&cloudshell_workspace=deploy/infrastructure-manager/gcp-credentials-json&show=terminal&ephemeral=true)

```bash
# For project-level monitoring (default)
./deploy_service_account.sh

# For organization-level monitoring
export ORG_ID="<YOUR_ORG_ID>"
./deploy_service_account.sh
```

#### Option 2: Local Terminal

```bash
cd deploy/infrastructure-manager/gcp-credentials-json

# For project-level monitoring (default)
./deploy_service_account.sh

# For organization-level monitoring
export ORG_ID="<YOUR_ORG_ID>"
./deploy_service_account.sh
```

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ORG_ID` | No | - | Organization ID for org-level monitoring |
| `DEPLOYMENT_NAME` | No | `elastic-agent-credentials` | Deployment name prefix |
| `LOCATION` | No | `us-central1` | GCP region for Infrastructure Manager |

### Resources Created

- Service account with `cloudasset.viewer` and `browser` roles
- Service account key (stored securely in Secret Manager)
- Secret Manager secret containing the JSON credentials
- IAM bindings (project or organization level)

### Output

After successful deployment, a `KEY_FILE.json` file is created containing the service account credentials.

**To use the credentials:**

1. Run `cat KEY_FILE.json` to view the key
2. Copy the entire JSON content
3. Paste it in the Elastic Agent GCP integration in Kibana
4. Save the key securely for future use

### Management

**View deployment:**
```bash
gcloud infra-manager deployments describe ${DEPLOYMENT_NAME} --location=${LOCATION}
```

**Delete deployment:**
```bash
gcloud infra-manager deployments delete ${DEPLOYMENT_NAME} --location=${LOCATION}
```

### Troubleshooting

**Common Issues:**

1. **Permission denied**: Ensure your account has the required IAM roles
2. **API not enabled**: The setup script enables required APIs automatically
3. **Organization scope fails**: Verify the ORG_ID is correct and you have org-level permissions

**Console:** [Infrastructure Manager Deployments](https://console.cloud.google.com/infra-manager/deployments)
