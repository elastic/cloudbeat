## Elastic Agent Service Account Infrastructure (Terraform)

Deploy a target service account in your GCP environment for Elastic Cloud Connectors. This module creates the service account that Elastic's global service account will impersonate to perform security monitoring in your GCP project or organization.

**Note**: This module is intended for customers. Before deploying, you'll need the `global_service_account_email` from Elastic's `gcp-global-sa` deployment.

### Prerequisites

1. GCP project with Elastic integration enabled
2. Global service account email from Elastic (from `gcp-global-sa` module)
3. Elastic resource ID (deployment identifier)
4. Required permissions (see [Required Permissions](#required-permissions))

### Quick Deploy

#### Option 1: Cloud Shell (Recommended)

[![Open in Cloud Shell](https://gstatic.com/cloudssh/images/open-btn.svg)](https://shell.cloud.google.com/cloudshell/editor?cloudshell_git_repo=https://github.com/elastic/cloudbeat.git&cloudshell_git_branch=main&cloudshell_workspace=deploy/infrastructure-manager/gcp-service-account&show=terminal&ephemeral=true)

```bash
# Set required configuration
export GLOBAL_SERVICE_ACCOUNT_EMAIL="<ELASTIC_GLOBAL_SA_EMAIL>"  # From gcp-global-sa output
export ELASTIC_RESOURCE_ID="<YOUR_DEPLOYMENT_ID>"                # Your Elastic deployment ID

# Optional: For organization-level monitoring
# export ORGANIZATION_ID="<YOUR_ORG_ID>"

# Optional: Override defaults
# export DEPLOYMENT_NAME="elastic-agent-sa"  # Default: elastic-agent-sa
# export LOCATION="us-central1"                       # Default: us-central1

# Deploy using the deploy script
./deploy.sh
```

#### Option 2: GCP Console

1. Go to [Infrastructure Manager Console](https://console.cloud.google.com/infra-manager/deployments/create)
2. Configure:
   - **Source**: Git repository
   - **Repository URL**: `https://github.com/elastic/cloudbeat.git`
   - **Branch**: `main`
   - **Directory**: `deploy/infrastructure-manager/gcp-service-account`
   - **Location**: `us-central1`
3. Add input variables (see table below)
4. Click **Create**

### Input Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `project_id` | Yes | - | Your GCP Project ID |
| `global_service_account_email` | Yes | - | Email of Elastic's global service account |
| `elastic_resource_id` | Yes | - | Unique identifier for your Elastic deployment |
| `scope` | No | `projects` | Scope: `projects` or `organizations` |
| `parent_id` | Yes | - | Project ID or Organization ID (based on scope) |
| `target_service_account_name` | No | `elastic-agent-sa` | Target service account name |

### Resources Created

- Target Service Account
- IAM role bindings:
  - `roles/cloudasset.viewer` (project or organization level)
  - `roles/browser` (project or organization level)
- Service Account Impersonation permission (Global SA → Target SA)

### Outputs

After deployment, you'll receive these outputs:

- `target_service_account_email`: Email of the target service account (use this in Elastic Agent configuration)
- `target_service_account_id`: Unique ID of the service account
- `project_id`: Your GCP Project ID
- `scope`: Deployment scope (projects or organizations)

**Save the `target_service_account_email`** - this is required for your Elastic Agent configuration.

### Elastic Agent Configuration

After deployment, configure Elastic Agent with these values:

```yaml
gcp:
  project_id: "<YOUR_PROJECT_ID>"
  account_type: "single-account"     # Or "organization-account"
  credentials:
    service_account_email: "<target_service_account_email from output>"
  supports_cloud_connectors: true
```

Environment variables (set by Elastic platform):
```bash
CLOUD_CONNECTORS_GCP_GLOBAL_SERVICE_ACCOUNT="<global_service_account_email>"
CLOUD_CONNECTORS_GCP_WORKLOAD_POOL="<pool_name>"
CLOUD_CONNECTORS_GCP_WORKLOAD_PROVIDER="<provider_name>"
CLOUD_CONNECTORS_GCP_PROJECT_NUMBER="<project_number>"
CLOUD_CONNECTORS_ID_TOKEN_FILE="/path/to/oidc/token"
```

### Management

**View deployment:**
```bash
gcloud infra-manager deployments describe ${DEPLOYMENT_NAME} --location=${LOCATION}
```

**View outputs:**
```bash
gcloud infra-manager deployments describe ${DEPLOYMENT_NAME} --location=${LOCATION} --format="value(terraformOutputs)"
```

**Delete deployment:**
```bash
gcloud infra-manager deployments delete ${DEPLOYMENT_NAME} --location=${LOCATION}
```

### Required Permissions

The deployment service account needs these roles:
- `roles/iam.serviceAccountAdmin` - Create and manage service accounts
- `roles/resourcemanager.projectIamAdmin` - Manage IAM bindings
- `roles/config.admin` - Infrastructure Manager operations

For organization-level deployments, you also need:
- `roles/resourcemanager.organizationAdmin` - Manage organization IAM

### Architecture

```
OIDC Token (from Elastic platform)
    ↓
Workload Identity Federation (in Elastic's project)
    ↓
Global Service Account (Elastic-owned)
    ↓
Target Service Account (Customer-owned) ← THIS MODULE
    ↓
Customer's GCP Resources (monitored)
```

### Account Types

**Single Account** (default):
- Monitors a single GCP project
- Set `scope=projects` and `parent_id=<project_id>`

**Organization Account**:
- Monitors all projects in a GCP organization
- Set `ORGANIZATION_ID` environment variable
- Requires organization-level permissions

### Security

This module implements security best practices:
- **No service account keys** - Uses keyless impersonation
- **Least privilege** - Only grants Viewer and Asset Viewer roles
- **Resource isolation** - Each deployment gets unique identifiers
- **Conditional access** - Can be enhanced with IAM conditions (see code comments)

### Troubleshooting

**Common Issues:**

1. **Invalid global_service_account_email**: Verify the email matches the output from `gcp-global-sa` deployment
2. **Permission denied on organization**: Ensure you have `roles/resourcemanager.organizationAdmin` for org-level deployments
3. **IAM binding failed**: Check that the global service account exists and is accessible

**Verify global SA exists:**
```bash
gcloud iam service-accounts describe ${GLOBAL_SERVICE_ACCOUNT_EMAIL}
```

**Console:** [Infrastructure Manager Deployments](https://console.cloud.google.com/infra-manager/deployments)
