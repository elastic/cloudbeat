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
# export DEPLOYMENT_NAME="elastic-agent-gcp"  # Default: elastic-agent-gcp
# export ZONE="us-central1-a"  # Default: us-central1-a
# export ELASTIC_ARTIFACT_SERVER="<CUSTOM_SERVER_URL>"  # Default: https://artifacts.elastic.co/downloads/beats/elastic-agent

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
| `elastic_artifact_server` | No | `https://artifacts.elastic.co/downloads/beats/elastic-agent` | Artifact server URL for downloading Elastic Agent |
| `zone` | No | `us-central1-a` | GCP zone |
| `scope` | No | `projects` | `projects` or `organizations` |
| `parent_id` | Yes | - | Project ID or Organization ID |
| `startup_validation_enabled` | No | `true` | Enable validation of startup script completion |
| `startup_timeout_seconds` | No | `600` | Maximum time to wait for startup (seconds) |

### Resources Created

- Compute instance (Ubuntu, n2-standard-4, 32GB disk)
- Service account with `cloudasset.viewer` and `browser` roles
- VPC network with auto-created subnets
- IAM bindings (project or organization level)

### Startup Validation

By default, Terraform waits for the startup script to complete and validates success:
- **Enabled**: Deployment fails if agent installation fails
- **Timeout**: 10 minutes (configurable via `startup_timeout_seconds`)
- **Requires**: `gcloud` CLI installed where Terraform runs

**Disable validation** (for testing or debugging):
```bash
# Via environment variable (for deploy.sh)
export STARTUP_VALIDATION_ENABLED=false
./deploy.sh

# Or pass to gcloud directly
gcloud infra-manager deployments apply ${DEPLOYMENT_NAME} \
  --location=${LOCATION} \
  --input-values="...,startup_validation_enabled=false"
```

**Guest Attributes Written**:

The startup script writes these attributes for monitoring:
- `elastic-agent/startup-status`: `"in-progress"`, `"success"`, or `"failed"`
- `elastic-agent/startup-error`: Error message (only when failed)
- `elastic-agent/startup-timestamp`: Completion timestamp (UTC)

Query manually:
```bash
gcloud compute instances get-guest-attributes ${INSTANCE_NAME} \
  --zone ${ZONE} \
  --query-path=elastic-agent/
```

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

**Check deployment status:**
```bash
# The instance name is based on the deployment name with a random suffix
# Format: elastic-agent-vm-<random-suffix>
# Example: elastic-agent-vm-0bc08b82

# Check startup script status via guest attributes
gcloud compute instances get-guest-attributes elastic-agent-vm-<suffix> \
  --zone ${ZONE} \
  --query-path=elastic-agent/startup-status

# Expected values:
# - "in-progress": Installation is running
# - "success": Installation completed successfully
# - "failed": Installation failed (check logs below)

# To find your instance name:
gcloud compute instances list --filter="name~^elastic-agent-vm-"
```

**Check agent logs (without SSH):**
```bash
# View serial console output (includes startup script execution)
gcloud compute instances get-serial-port-output ${INSTANCE_NAME} --zone ${ZONE}

# Filter for elastic-agent specific logs
gcloud compute instances get-serial-port-output ${INSTANCE_NAME} --zone ${ZONE} \
  | grep elastic-agent-setup
```

**Check agent logs (with SSH):**
```bash
gcloud compute ssh ${INSTANCE_NAME} --zone ${ZONE}
sudo journalctl -u google-startup-scripts.service
```

**Common Issues:**

1. **404 error downloading agent**: Check `ELASTIC_ARTIFACT_SERVER` and `STACK_VERSION` are correct
2. **Guest attributes show "failed"**: Check serial console logs for error details
3. **Guest attributes not available**: Guest attributes are enabled by default and populate during startup

**Console:** [Infrastructure Manager Deployments](https://console.cloud.google.com/infrastructure-manager/deployments)
