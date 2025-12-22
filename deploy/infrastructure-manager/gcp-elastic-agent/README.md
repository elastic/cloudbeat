## Elastic Agent Infrastructure Manager (Terraform)

Deploy Elastic Agent for CIS GCP integration using GCP Infrastructure Manager. Creates a compute instance with Elastic Agent pre-installed and configured with necessary permissions.

### Prerequisites

1. Elastic Stack with Fleet Server deployed
2. GCP project with required permissions (see [Required Permissions](#required-permissions))
3. Fleet URL and enrollment token from Kibana

### Quick Deploy

#### Option 1: Cloud Shell (Recommended)

[![Open in Cloud Shell](https://gstatic.com/cloudssh/images/open-btn.svg)](https://shell.cloud.google.com/cloudshell/editor?cloudshell_git_repo=https://github.com/elastic/cloudbeat.git&cloudshell_git_branch=main&cloudshell_workspace=deploy/infrastructure-manager/gcp-elastic-agent&show=terminal&ephemeral=true)

Setup once
```bash
# Enable required APIs
gcloud services enable iam.googleapis.com config.googleapis.com compute.googleapis.com \
    cloudresourcemanager.googleapis.com cloudasset.googleapis.com

# Create a service-account to run infra-manager scripts
gcloud iam service-accounts create infra-manager-deployer \
    --display-name="Infra Manager Deployment Account"

# Grant permissions to manage resources and Infrastructure Manager state
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:infra-manager-deployer@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/compute.admin"

gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:infra-manager-deployer@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/iam.serviceAccountAdmin"

gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:infra-manager-deployer@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/resourcemanager.projectIamAdmin"

gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:infra-manager-deployer@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/config.admin"
```

Deploy
```bash
# Set required configuration
export FLEET_URL="<YOUR_FLEET_URL>"
export ENROLLMENT_TOKEN="<YOUR_TOKEN>"
export STACK_VERSION="<YOUR_AGENT_VERSION>"

# Optional: Set these to override defaults
# export ORG_ID="<YOUR_ORG_ID>"  # For org-level monitoring
# export DEPLOYMENT_NAME="elastic-agent-cspm"  # Default: elastic-agent-cspm
# export ZONE="us-central1-a"  # Default: us-central1-a

# Deploy using the deploy script
./deploy.sh
```

#### Option 2: GCP Console

1. Go to [Infrastructure Manager Console](https://console.cloud.google.com/infrastructure-manager/deployments/create)
2. Configure:
   - **Source**: Git repository
   - **Repository URL**: `https://github.com/elastic/cloudbeat.git`
   - **Branch**: `main`
   - **Directory**: `deploy/infrastructure-manager/gcp-elastic-agent`
   - **Location**: `us-central1`
3. Add input variables (see table below)
4. Click **Create**

### Input Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `project_id` | Yes | - | GCP Project ID |
| `fleet_url` | Yes | - | Fleet Server URL |
| `enrollment_token` | Yes | - | Enrollment token (sensitive) |
| `elastic_agent_version` | Yes | - | Agent version (e.g., `8.15.0`) |
| `zone` | No | `us-central1-a` | GCP zone |
| `scope` | No | `projects` | `projects` or `organizations` |
| `parent_id` | Yes | - | Project ID or Organization ID |

### Resources Created

- Compute instance (Ubuntu, n2-standard-4, 32GB disk)
- Service account with `cloudasset.viewer` and `browser` roles
- VPC network with auto-created subnets
- IAM bindings (project or organization level)

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

**Check agent logs:**
```bash
gcloud compute ssh ${DEPLOYMENT_NAME} --zone ${ZONE}
sudo journalctl -u google-startup-scripts.service
```

**Console:** [Infrastructure Manager Deployments](https://console.cloud.google.com/infrastructure-manager/deployments)
