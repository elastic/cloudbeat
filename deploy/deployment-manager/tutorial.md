# Deploying the CIS GCP integration using Deployment Manager

## Script Overview

The script you are about to execute automates the creation of a new deployment, a compute instance, and the attachment of a service account with a custom role that holds the necessary permissions for running the CIS GCP integration.
The following steps are involved in the script execution:

### Step 1: Enabling Required APIs

Before deploying the CIS GCP integration, certain APIs must be enabled. These APIs include:

1. IAM: Required for creating a service account and a custom role to be attached to the compute instance.
2. Deployment Manager: Necessary for creating the deployment itself.
3. Compute: Used to create a new compute instance to host the Elastic agent.
4. Cloud Resource Manager: Enables querying information about the project being analyzed.
5. Cloud Asset Inventory: Facilitates data collection about assets linked to the project.


### Step 2: Adding Roles to the Default Service Account

The template used in this deployment creates a service account, custom IAM role, and service account bindings.
However, prior to the successful deployment, the Google APIs Service Agent service account must be granted the `Role Administrator` and `Project IAM Admin` roles. For more information on this account, you can refer to the [Google-managed service account documentation](https://cloud.google.com/iam/docs/maintain-custom-roles-deployment-manager).

Once the deployment is completed, the script will remove the roles from the default service account. If you wish to delete the deployment, you will need to manually add the roles back to the default service account.

## Step 3: Applying the Deployment Manager Templates

Upon executing the script, the following resources will be created:

1. A compute instance with the Elastic agent installed.
2. A service account with a custom role attached.
3. A role with appropriate permissions required for the integration.
4. A service account binding that associates the custom role with the compute instance.

In case the deployment encounters any issues and fails, the script will attempt to delete the deployment along with all the associated resources that were created during the process.
