## Elastic Agent Infrastructure Manager (Terraform)

Deploy Elastic Agent for CIS GCP integration using GCP Infrastructure Manager. Creates a compute instance with Elastic Agent pre-installed and configured with necessary permissions.

### Prerequisites

1. Elastic Stack with Fleet Server deployed
2. GCP project with required permissions (Editor, IAM Admin)
3. Fleet URL and enrollment token from Kibana

### Quick Deploy

#### Option 1: Cloud Shell (Recommended)

[![Open in Cloud Shell](https://gstatic.com/cloudssh/images/open-btn.svg)](https://shell.cloud.google.com/cloudshell/editor?cloudshell_git_repo=https://github.com/elastic/cloudbeat.git&cloudshell_workspace=deploy/infrastructure-manager&show=terminal&ephemeral=true)

```bash
# Enable required APIs
gcloud services enable iam.googleapis.com config.googleapis.com compute.googleapis.com \
    cloudresourcemanager.googleapis.com cloudasset.googleapis.com

# Set variables from current session
export PROJECT_ID=$(gcloud config get-value project)
export ZONE=$(gcloud config get-value compute/zone 2>/dev/null || echo "us-central1-a")
export LOCATION=$(echo $ZONE | sed 's/-[a-z]$//')  # Extract region from zone (e.g., us-central1-a -> us-central1)

# Deploy
gcloud infra-manager deployments apply elastic-agent-cspm \
    --location=${LOCATION} \
    --service-account="projects/${PROJECT_ID}/serviceAccounts/$(gcloud projects describe ${PROJECT_ID} --format='value(projectNumber)')@cloudservices.gserviceaccount.com" \
    --git-source-repo="https://github.com/elastic/cloudbeat.git" \
    --git-source-directory="deploy/infrastructure-manager" \
    --git-source-ref="main" \
    --input-values="\
project_id=${PROJECT_ID},\
deployment_name=elastic-agent-cspm,\
zone=${ZONE},\
fleet_url=YOUR_FLEET_URL,\
enrollment_token=YOUR_TOKEN,\
elastic_agent_version=8.19.0,\
scope=projects,\
parent_id=${PROJECT_ID},\
allow_ssh=false"
```

Replace `YOUR_FLEET_URL` and `YOUR_TOKEN` with your actual values.

For **organization-level** deployment, use `scope=organizations,parent_id=YOUR_ORG_ID`.

#### Option 2: GCP Console

1. Go to [Infrastructure Manager Console](https://console.cloud.google.com/infrastructure-manager/deployments/create)
2. Configure:
   - **Source**: Git repository
   - **Repository URL**: `https://github.com/elastic/cloudbeat.git`
   - **Branch**: `main`
   - **Directory**: `deploy/infrastructure-manager`
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
gcloud infra-manager deployments describe elastic-agent-cspm --location=us-central1
```

**Delete deployment:**
```bash
gcloud infra-manager deployments delete elastic-agent-cspm --location=us-central1
```

### Troubleshooting

**Check agent logs:**
```bash
gcloud compute ssh elastic-agent-cspm --zone us-central1-a
sudo journalctl -u google-startup-scripts.service
```

**Console:** [Infrastructure Manager Deployments](https://console.cloud.google.com/infrastructure-manager/deployments)
