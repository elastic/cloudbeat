#!/usr/bin/env python
"""
This script installs CSPM integrations on the 'Agentless' agent policy.
The following steps are performed:
1. Create a CSPM AWS integration.
2. Create a CSPM Azure integration.
3. Create a CSPM GCP integration.
"""

import json
import configuration_fleet as cnfg
from loguru import logger
from fleet_api.package_policy_api import create_cspm_integration
from package_policy import (
    generate_package_policy,
    generate_policy_template,
    generate_random_name,
)

AGENT_POLICY_ID = "agentless"


def generate_aws_integration_data():
    """
    Generate data for creating CSPM AWS integration
    """
    return {
        "name": generate_random_name("agentless-pkg-cspm-aws"),
        "input_name": "cis_aws",
        "posture": "cspm",
        "deployment": "aws",
        "vars": {
            "aws.account_type": "single-account",
            "aws.credentials.type": "direct_access_keys",
            "access_key_id": cnfg.aws_config.access_key_id,
            "secret_access_key": cnfg.aws_config.secret_access_key,
        },
    }


def generate_azure_integration_data():
    """
    Generate data for creating CSPM Azure integration
    """
    creds = json.loads(cnfg.azure_arm_parameters.credentials)
    return {
        "name": generate_random_name("agentless-pkg-cspm-azure"),
        "input_name": "cis_azure",
        "posture": "cspm",
        "deployment": "azure",
        "vars": {
            "azure.account_type": "single-account",
            "azure.credentials.type": "service_principal_with_client_secret",
            "azure.credentials.client_id": creds["clientId"],
            "azure.credentials.tenant_id": creds["tenantId"],
            "azure.credentials.client_secret": creds["clientSecret"],
        },
    }


def generate_gcp_integration_data():
    """
    Generate data for creating CSPM GCP integration
    """
    with open(cnfg.gcp_dm_config.credentials_file, "r") as credentials_json_file:
        credentials_json = credentials_json_file.read()
    return {
        "name": generate_random_name("agentless-pkg-cspm-gcp"),
        "input_name": "cis_gcp",
        "posture": "cspm",
        "deployment": "gcp",
        "vars": {
            "setup_access": "manual",
            "gcp.account_type": "single-account",
            "gcp.credentials.type": "credentials-json",
            "gcp.credentials.json": credentials_json,
        },
    }


if __name__ == "__main__":
    integrations = [
        generate_aws_integration_data(),
        generate_azure_integration_data(),
    ]
    cspm_template = generate_policy_template(cfg=cnfg.elk_config)
    for integration_data in integrations:
        NAME = integration_data["name"]
        logger.info(f"Creating {NAME} integration for policy {AGENT_POLICY_ID}")
        package_policy = generate_package_policy(cspm_template, integration_data)
        package_policy["force"] = True

        logger.info(f"Creating {package_policy}")

        create_cspm_integration(
            cfg=cnfg.elk_config,
            pkg_policy=package_policy,
            agent_policy_id=AGENT_POLICY_ID,
            cspm_data={},
        )
        logger.info(f"Installation of {NAME} integration is done")
