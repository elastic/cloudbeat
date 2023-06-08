# Automated Manifest Generation for KSPM and CSPM

This package provides scripts that automate the generation of KSPM (Kubernetes Security Posture Management) manifests, both for unmanaged and EKS (Elastic Kubernetes Service) setups, as well as CSPM (Cloud Security Posture Management) tasks. The purpose of these scripts is to streamline the generation process and make it more efficient.

## Prerequisites

Before running the scripts, ensure that you have set the following environment variables:

### EC (Elastic Cloud) Instance with ELK Configuration

- `ES_USER`: The username for the Elasticsearch instance.
- `ES_PASSWORD`: The password for the Elasticsearch instance.
- `KIBANA_URL`: The URL of the Kibana instance.

### AWS Configuration

- `AWS_ACCESS_KEY_ID`: The access key ID for your AWS account.
- `AWS_SECRET_ACCESS_KEY`: The secret access key for your AWS account.

Make sure to set these variables with the appropriate values based on your specific setup.

## Installation and Execution

Follow these steps to install the dependencies and execute the different scripts:

1. Open your terminal and navigate to the directory `deploy/test-environments/fleet_api` using the following command:

    ```bash
    cd deploy/test-environments/fleet_api
    ```

2. Install the dependencies using Poetry by running the command:

    ``` bash
    poetry install
    ```

3. To execute the KSPM unmanaged integration, use the following command:

    ``` bash
    poetry run python src/install_kspm_unmanaged_integration.py
    ```

4. To execute the KSPM EKS integration, use the following command:

    ``` bash
    poetry run python src/install_kspm_eks_integration.py
    ```

5. To execute the CSPM integration, use the following command:

    ``` bash
    poetry run python src/install_cspm_integration.py
    ```

6. To purge integrations, use the following command:

    ``` bash
    poetry run python src/purge_integrations.py
    ```