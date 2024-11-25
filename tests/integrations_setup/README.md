# Automated Manifest Generation

This package automates the generation and deployment of manifests for various integrations, including KSPM, CSPM, CNVM, Asset Inventory, and Wiz.

### Key Features:

- `Policy Creation`: Generates agent and package policies.
- `Kibana Installation`: Installs the policies directly in Kibana.
- `Manifest Generation`: Creates Elastic Agent manifests or scripts for deployment on environments like VMs or EC2 instances.

### Supported Integrations:

- `KSPM`: For unmanaged clusters and EKS.
- `CSPM`: Supports AWS, Azure, and GCP cloud providers.
- `CNVM`: Supports AWS cloud provider only.
- `Asset Inventory`: For AWS, Azure cloud providers.
- `Third-Party Integrations`:
  - `Wiz`: Includes configuration for CSPM and CNVM.

## Prerequisites

Before running the scripts, ensure that you have set the following environment variables:

### Common Variables for All Integration Scripts

- `ES_USER`: The username for the Elasticsearch instance.
- `ES_PASSWORD`: The password for the Elasticsearch instance.
- `KIBANA_URL`: The URL of the Kibana instance.
- `STACK_VERSION`: The version of the Elastic Stack being used.

### Integration-Specific Variables

Each integration has its own environment variables related to the package being installed. These can be found in the [configuration_fleet](./configuration_fleet.py) file.

Make sure to set these variables with the appropriate values based on your specific setup.

## Installation and Execution

Follow these steps to install the dependencies and execute the different scripts:

1. Open your terminal and navigate to the directory `tests` using the following command:

    ```bash
    cd tests
    ```

2. Install the dependencies using Poetry by running the command:

    ``` bash
    poetry install
    ```

3. To execute the KSPM unmanaged integration, use the following command:

    ``` bash
    poetry run python ./integrations_setup/install_kspm_unmanaged_integration.py
    ```

4. To execute the KSPM EKS integration, use the following command:

    ``` bash
    poetry run python ./integrations_setup//install_kspm_eks_integration.py
    ```

5. To execute the CSPM integration, use the following command:

    ``` bash
    poetry run python ./integrations_setup/install_cspm_integration.py
    ```

6. To execute the CNVM integration, use the following command:

    ``` bash
    poetry run python ./integrations_setup/install_cnvm_integration.py
    ```

7. To execute the Defend for Containers (D4C) integration, use the following command:

    ``` bash
    poetry run python ./integrations_setup/install_d4c_integration.py
    ```

8. To execute the CSPM GCP integration, use the following command:

    ``` bash
    poetry run python ./integrations_setup/install_cspm_gcp_integration.py
    ```

9. To execute the CSPM AZURE integration, use the following command:

    ``` bash
    poetry run python ./integrations_setup/install_cspm_azure_integration.py
    ```

10. To execute the AWS Asset Inventory integration, use the following command:

    ``` bash
    poetry run python ./integrations_setup/install_aws_asset_inventory_integration.py
    ```

11. To execute the Azure Asset Inventory integration, use the following command:

    ``` bash
    poetry run python ./integrations_setup/install_azure_asset_inventory_integration.py
    ```

12. To execute the Wiz integration, use the following command:

    ``` bash
    poetry run python ./integrations_setup/install_wiz_integration.py
    ```

13. To purge integrations, use the following command:

    ``` bash
    poetry run python ./integrations_setup/purge_integrations.py
    ```
