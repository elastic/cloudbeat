#!/usr/bin/env python
"""
This script installs CSPM integrations for Agentless agents.
The following steps are performed:
1. Create a CSPM AWS integration.
2. Create a CSPM Azure integration.
3. Create a CSPM GCP integration.
"""

import json

import configuration_fleet as cnfg
from fleet_api.agent_policy_api import create_agent_policy
from fleet_api.package_policy_api import create_cspm_integration
from loguru import logger
from package_policy import generate_policy_template, generate_random_name, load_data


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
            "gcp.account_type": "single-account",
            "gcp.credentials.type": "credentials-json",
            "gcp.credentials.json": credentials_json,
        },
    }


if __name__ == "__main__":
    integrations = [
        generate_aws_integration_data(),
        generate_azure_integration_data(),
        generate_gcp_integration_data(),
    ]
    cspm_template = generate_policy_template(
        cfg=cnfg.elk_config,
        stream_prefix="cloud_security_posture",
    )
    for integration_data in integrations:
        INTEGRATION_NAME = integration_data["name"]
        AGENTLESS_INPUT = {
            "name": f"Agentless policy for {INTEGRATION_NAME}",
            "supports_agentless": True,
        }

        logger.info(f"Starting installation of agentless-agent {INTEGRATION_NAME} integration.")
        agent_data, package_data = load_data(
            cfg=cnfg.elk_config,
            agent_input=AGENTLESS_INPUT,
            package_input=integration_data,
            stream_name="cloud_security_posture.findings",
        )

        logger.info("Create agentless-agent policy")
        agent_policy_id = create_agent_policy(cfg=cnfg.elk_config, json_policy=agent_data)

        logger.info(f"Create agentless-agent {INTEGRATION_NAME} integration")
        package_policy_id = create_cspm_integration(
            cfg=cnfg.elk_config,
            pkg_policy=package_data,
            agent_policy_id=agent_policy_id,
            cspm_data={},
        )

        logger.info(f"Installation of {INTEGRATION_NAME} integration is done")