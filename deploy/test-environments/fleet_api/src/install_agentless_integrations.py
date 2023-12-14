#!/usr/bin/env python
"""
This script installs CSPM integrations on the 'Agentless' agent policy.

The following steps are performed:
1. Create a CSPM AWS integration.
2. Create a CSPM Azure integration.
3. Create a CSPM GCP integration.
"""

import os
import configuration_fleet as cnfg
from api.package_policy_api import create_cspm_integration
from package_policy import (
    generate_package_policy,
    generate_policy_template,
)
from loguru import logger

AGENT_POLICY_ID = "agentless"


def generate_aws_integration_data():
    """
    Generate data for creating CSPM AWS integration
    """
    access_key_id = os.getenv("AWS_ACCESS_KEY_ID", "")
    secret_access_key = os.getenv("AWS_SECRET_ACCESS_KEY", "")
    return {
        "name": "cspm_aws",
        "input_name": "cis_aws",
        "posture": "cspm",
        "deployment": "aws",
        "vars": {
            "aws.account_type": "single-account",
            "aws.credentials.type": "direct_access_keys",
            "access_key_id": access_key_id,
            "secret_access_key": secret_access_key,
        },
    }


def generate_azure_integration_data():
    """
    Generate data for creating CSPM Azure integration
    """
    return {
        "name": "cspm_azure",
        "input_name": "cis_azure",
        "posture": "cspm",
        "deployment": "azure",
        "vars": {
            "azure.account_type": "single-account",
            "azure.credentials.type": "manual",
        },
    }


def generate_gcp_integration_data():
    """
    Generate data for creating CSPM GCP integration
    """
    application_credentials = os.getenv("GOOGLE_APPLICATION_CREDENTIALS", "")
    with open(application_credentials, "r") as creadentials_json_file:
        creadentials_json = creadentials_json_file.read()
    return {
        "name": "cspm_gcp",
        "input_name": "cis_gcp",
        "posture": "cspm",
        "deployment": "gcp",
        "vars": {
            "setup_access": "manual",
            "gcp.account_type": "single-account",
            "gcp.credentials.type": "credentials-json",
            "gcp.credentials.json": creadentials_json,
        },
    }


if __name__ == "__main__":
    integrations = [
        generate_aws_integration_data(),
    ]
    cspm_template = generate_policy_template(cfg=cnfg.elk_config)
    for integration_data in integrations:
        NAME = integration_data["name"]
        logger.info(f"Create {NAME} integration for policy {AGENT_POLICY_ID}")
        package_policy = generate_package_policy(cspm_template, integration_data)

        logger.info(f"Created {package_policy}")

        create_cspm_integration(
            cfg=cnfg.elk_config,
            pkg_policy=package_policy,
            agent_policy_id=AGENT_POLICY_ID,
            cspm_data={},
        )
        logger.info(f"Installation of {NAME} integration is done")
