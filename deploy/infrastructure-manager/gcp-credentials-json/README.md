## GCP Credentials JSON (Service Account Key)

Deploy a GCP service account with JSON credentials for Elastic Agent GCP integration using GCP Infrastructure Manager.

This creates a service account with the necessary permissions and stores the JSON key in Secret Manager for use in the Elastic Agent GCP integration in Kibana.

### Prerequisites

1. GCP project with required permissions
2. `gcloud` CLI configured with your project

### Quick Deploy

#### Option 1: Cloud Shell (Recommended)

[![Open in Cloud Shell](https://gstatic.com/cloudssh/images/open-btn.svg)](https://shell.cloud.google.com/cloudshell/editor?cloudshell_git_repo=https://github.com/elastic/cloudbeat.git&cloudshell_git_branch=main&cloudshell_workspace=deploy/infrastructure-manager/gcp-credentials-json&show=terminal&ephemeral=true)

```bash
# For project-level monitoring (default)
./deploy.sh

# For organization-level monitoring
export ORG_ID="<YOUR_ORG_ID>"
./deploy.sh
```

#### Option 2: GCP Console

1. Go to [Infrastructure Manager Console](https://console.cloud.google.com/infra-manager/deployments/create)
2. Configure:
   - **Source**: Git repository
   - **Repository URL**: `https://github.com/elastic/cloudbeat.git`
   - **Branch**: `main`
   - **Directory**: `deploy/infrastructure-manager/gcp-credentials-json`
   - **Location**: `us-central1`
3. Add input variables:
   - `project_id`: Your GCP project ID
   - `resource_suffix`: Unique suffix (e.g., `abc123`)
   - `scope`: `projects` or `organizations`
   - `parent_id`: Project ID or Organization ID
4. Click **Create**

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ORG_ID` | No | - | Organization ID for org-level monitoring |
| `DEPLOYMENT_NAME` | No | `elastic-agent-credentials` | Deployment name prefix |
| `LOCATION` | No | `us-central1` | GCP region for Infrastructure Manager |

### Resources Created

- Service account with `cloudasset.viewer` and `browser` roles
- Service account key (stored securely in Secret Manager and saved locally)
- Secret Manager secret containing the JSON credentials
- IAM bindings (project or organization level)
- Local `KEY_FILE.json` with the service account credentials

### Output

After successful deployment, the script saves the service account credentials to `KEY_FILE.json` in the current directory.

**To use the credentials:**

1. Run `cat KEY_FILE.json` to view the service account key
2. Copy the entire JSON content
3. Paste it in the Elastic Agent GCP integration in Kibana

> **Note:** The key is also stored in Secret Manager for future access. The script outputs the `gcloud` command to retrieve it if needed.

### Required Permissions

The deployment service account needs these roles:
- `roles/iam.serviceAccountAdmin` - Create and manage service accounts
- `roles/iam.serviceAccountKeyAdmin` - Create service account keys
- `roles/resourcemanager.projectIamAdmin` - Manage project-level IAM bindings
- `roles/config.admin` - Infrastructure Manager operations
- `roles/storage.admin` - Store Terraform state
- `roles/secretmanager.admin` - Create and manage secrets

For organization-level deployments (when `ORG_ID` is set), you also need:
- `roles/iam.securityAdmin` - Manage organization IAM bindings (granted at organization level)

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
