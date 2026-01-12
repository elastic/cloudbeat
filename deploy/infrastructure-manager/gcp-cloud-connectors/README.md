## Elastic Agent Service Account Infrastructure (Terraform)

Deploy a target service account in your GCP environment for Elastic Cloud Connectors. This module creates the service account that Elastic's AWS role will impersonate to perform security monitoring in your GCP project or organization.

### Prerequisites

1. GCP project with required APIs enabled
2. Elastic's AWS role ARN: `arn:aws:iam::254766567737:role/cloud_connectors`
3. Required permissions (see [Required Permissions](#required-permissions))

### Quick Deploy

#### Option 1: Cloud Shell (Recommended)

[![Open in Cloud Shell](https://gstatic.com/cloudssh/images/open-btn.svg)](https://shell.cloud.google.com/cloudshell/editor?cloudshell_git_repo=https://github.com/elastic/cloudbeat.git&cloudshell_git_branch=main&cloudshell_workspace=deploy/infrastructure-manager/gcp-cloud-connectors&show=terminal&ephemeral=true)

```bash
# Set required configuration
export ELASTIC_RESOURCE_ID="<YOUR_DEPLOYMENT_ID>"  # Your Elastic deployment ID (must match AWS role session name)

# Optional: For organization-level monitoring
# export ORGANIZATION_ID="<YOUR_ORG_ID>"

# Optional: Override defaults
# export DEPLOYMENT_NAME="elastic-agent-sa"                                    # Default: elastic-agent-sa
# export LOCATION="us-central1"                                                # Default: us-central1
# export ELASTIC_ROLE_ARN="arn:aws:iam::254766567737:role/cloud_connectors"    # Default: Elastic's role

# Deploy using the deploy script
./deploy.sh
```

#### Option 2: GCP Console

1. Go to [Infrastructure Manager Console](https://console.cloud.google.com/infra-manager/deployments/create)
2. Configure:
   - **Source**: Git repository
   - **Repository URL**: `https://github.com/elastic/cloudbeat.git`
   - **Branch**: `main`
   - **Directory**: `deploy/infrastructure-manager/gcp-cloud-connectors`
   - **Location**: `us-central1`
3. Add input variables (see table below)
4. Click **Create**

### Input Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `project_id` | Yes | - | Your GCP Project ID |
| `elastic_resource_id` | Yes | - | Unique identifier for your Elastic deployment (must match AWS role session name) |
| `elastic_role_arn` | No | `arn:aws:iam::254766567737:role/cloud_connectors` | ARN of Elastic's AWS IAM role to trust |
| `scope` | No | `projects` | Scope: `projects` or `organizations` |
| `parent_id` | Yes | - | Project ID or Organization ID (based on scope) |
| `target_service_account_name` | No | `elastic-agent-sa` | Target service account name |

### Resources Created

- Workload Identity Pool (for AWS federation)
- Workload Identity Provider (AWS type)
- Target Service Account
- IAM role bindings:
  - `roles/cloudasset.viewer` (project or organization level)
  - `roles/browser` (project or organization level)
- Service Account Impersonation permission (AWS Role → GCP SA)

### Outputs

After deployment, you'll receive these outputs:

- `target_service_account_email`: Email of the target service account (use this in Elastic Agent configuration)
- `gcp_audience`: GCP audience URL for Workload Identity Federation

**Save the `target_service_account_email`** - this is required for your Elastic Agent configuration.

### Architecture

```
AWS Role (arn:aws:iam::254766567737:role/cloud_connectors)
    ↓ AWS STS GetCallerIdentity
Workload Identity Federation (AWS provider in customer's GCP project)
    ↓ Token exchange
Target Service Account (Customer-owned) ← THIS MODULE
    ↓ Impersonation
Customer's GCP Resources (monitored)
```

### How It Works

1. Elastic's service runs with the AWS role `arn:aws:iam::254766567737:role/cloud_connectors`
2. The AWS role calls GCP's STS token exchange endpoint with its AWS credentials
3. GCP's Workload Identity Federation validates the AWS identity and issues a GCP token
4. The GCP token allows impersonation of the customer's target service account
5. Elastic Agent uses the impersonated identity to read cloud assets

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
- **No service account keys** - Uses keyless AWS-to-GCP federation
- **Least privilege** - Only grants Viewer and Asset Viewer roles
- **Resource isolation** - Each deployment gets unique identifiers
- **Two-layer validation**:
  1. AWS role ARN must match `arn:aws:iam::254766567737:role/cloud_connectors`
  2. AWS role session name must match the `elastic_resource_id`
  
The attribute condition validates the full assumed-role ARN:
```
arn:aws:sts::254766567737:assumed-role/cloud_connectors/<elastic_resource_id>
```

This ensures only the specific Elastic deployment (identified by `elastic_resource_id`) can access the customer's GCP resources, even if multiple deployments share the same AWS role.

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
- `roles/iam.workloadIdentityPoolAdmin` - Create Workload Identity Pool and Provider
- `roles/resourcemanager.projectIamAdmin` - Manage IAM bindings
- `roles/config.admin` - Infrastructure Manager operations

For organization-level deployments, you also need:
- `roles/resourcemanager.organizationAdmin` - Manage organization IAM

### Troubleshooting

**Common Issues:**

1. **Permission denied on Workload Identity Pool creation**: Ensure you have `roles/iam.workloadIdentityPoolAdmin`
2. **Permission denied on organization**: Ensure you have `roles/resourcemanager.organizationAdmin` for org-level deployments
3. **IAM binding failed**: Check that the service account has required permissions

**Verify deployment:**
```bash
# Check Workload Identity Pool
gcloud iam workload-identity-pools list --location=global

# Check service account
gcloud iam service-accounts list --filter="email:elastic-agent-sa"
```

**Console:** [Infrastructure Manager Deployments](https://console.cloud.google.com/infra-manager/deployments)
