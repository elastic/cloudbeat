## Elastic Agent Infrastructure Manager (Terraform)

Deploy Elastic Agent for CIS GCP integration using GCP Infrastructure Manager. Creates a compute instance with Elastic Agent pre-installed and configured with necessary permissions.

### Prerequisites

1. Elastic Stack with Fleet Server deployed
2. GCP project with required permissions (Editor, IAM Admin)
3. Fleet URL and enrollment token from Kibana

### Quick Deploy

#### Option 1: Cloud Shell (Recommended)

[![Open in Cloud Shell](https://gstatic.com/cloudssh/images/open-btn.svg)](https://shell.cloud.google.com/cloudshell/editor?cloudshell_git_repo=https://github.com/elastic/cloudbeat.git&cloudshell_git_branch=main&cloudshell_workspace=deploy/infrastructure-manager/gcp-elastic-agent&show=terminal&ephemeral=true)

```bash
# Enable required APIs
gcloud services enable iam.googleapis.com config.googleapis.com compute.googleapis.com \
    cloudresourcemanager.googleapis.com cloudasset.googleapis.com
```

```bash
# Set deployment configuration
export ORG_ID=""  # Optional: Set to your organization ID for org-level monitoring
export PROJECT_ID="<YOUR_GCP_PROJECT_ID>"
export DEPLOYMENT_NAME="elastic-agent-cspm"
export FLEET_URL="<YOUR_FLEET_URL>"
export ENROLLMENT_TOKEN="<YOUR_TOKEN>"
export ELASTIC_AGENT_VERSION="<YOUR_AGENT_VERSION>"
export ZONE="us-central1-a"  # Change if needed

# Automatically set scope and parent_id based on ORG_ID
if [ -n "${ORG_ID}" ]; then
  export SCOPE="organizations"
  export PARENT_ID="${ORG_ID}"
else
  export SCOPE="projects"
  export PARENT_ID="${PROJECT_ID}"
fi

# Configure GCP project and location
gcloud config set project ${PROJECT_ID}
export LOCATION=$(echo ${ZONE} | sed 's/-[a-z]$//')  # Extract region from zone

# Deploy from local source (repo already cloned by Cloud Shell)
gcloud infra-manager deployments apply ${DEPLOYMENT_NAME} \
    --location=${LOCATION} \
    --service-account="projects/${PROJECT_ID}/serviceAccounts/$(gcloud projects describe ${PROJECT_ID} --format='value(projectNumber)')@cloudservices.gserviceaccount.com" \
    --local-source="." \
    --input-values="\
project_id=${PROJECT_ID},\
deployment_name=${DEPLOYMENT_NAME},\
zone=${ZONE},\
fleet_url=${FLEET_URL},\
enrollment_token=${ENROLLMENT_TOKEN},\
elastic_agent_version=${ELASTIC_AGENT_VERSION},\
scope=${SCOPE},\
parent_id=${PARENT_ID}"
```

**For organization-level monitoring:** Set `export ORG_ID="<YOUR_ORG_ID>"` before running the deploy command.

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
| `deployment_name` | Yes | - | Deployment name |
| `fleet_url` | Yes | - | Fleet Server URL |
| `enrollment_token` | Yes | - | Enrollment token (sensitive) |
| `elastic_agent_version` | Yes | - | Agent version (e.g., `8.15.0`) |
| `zone` | No | `us-central1-a` | GCP zone |
| `scope` | No | `projects` | `projects` or `organizations` |
| `parent_id` | Yes | - | Project ID or Organization ID |
| `service_account_name` | No | `""` | Existing SA (creates new if empty) |
| `allow_ssh` | No | `false` | Enable SSH firewall rule |

### Resources Created

- Compute instance (Ubuntu, n2-standard-4, 32GB disk)
- Service account with `cloudasset.viewer` and `browser` roles
- VPC network with auto-created subnets
- IAM bindings (project or organization level)
- Optional: SSH firewall rule

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
