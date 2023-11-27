#!/usr/bin/env python
"""
This script updates AWS CNVM agent.

The following steps are performed:
1. Download the latest CNVM template.
2. Get all the required parameters.
3. Execute a CloudFormation stack update.

Note: This script requires the configuration and dependencies provided by the 'cnfg' and 'utils' modules.

For execution, you can create a configuration file 'cnvm_config.json' in the same directory.

Example 'cnvm_config.json':
{
  "ENROLLMENT_TOKEN": "YourEnrollmentToken"
}

Ensure that AWS credentials are properly configured for Boto3.

You can also modify the 'stack_tags' variable to set custom tags for the CloudFormation stack.

"""
from pathlib import Path
import boto3
from munch import Munch
from loguru import logger
from utils import read_json
import configuration_fleet as cnfg
from api.common_api import (
    get_artifact_server,
    get_fleet_server_host,
)
from package_policy import (
    get_package_default_url,
    extract_template_url,
)


CNVM_JSON_PATH = Path(__file__).parent / "cnvm_config.json"


def update_cloudformation_stack(cfg: Munch):
    """
    Update an AWS CloudFormation stack with the provided configuration.

    Args:
        cfg (Munch): A configuration object containing the following attributes:
            - stack_name (str): The name of the CloudFormation stack to update.
            - template (str): The URL or S3 path to the CloudFormation template.
            - elastic_agent_version (str): The Elastic Agent version to set as a parameter.
            - elastic_artifact_server (str): The Elastic Artifact Server URL to set as a parameter.
            - enrollment_token (str): The Enrollment Token to set as a parameter.
            - fleet_url (str): The Fleet URL to set as a parameter.
            - stack_tags (list of dict): Tags to apply to the CloudFormation stack.

    Returns:
        None

    The function performs a CloudFormation stack update using the provided configuration.
    It initiates the stack update, waits for the update to complete, and logs the status.
    """
    # Create a Boto3 CloudFormation client
    cf_client = boto3.client("cloudformation")

    # Parameters in the format ParameterKey=Key,ParameterValue=Value
    parameters = [
        {"ParameterKey": "ElasticAgentVersion", "ParameterValue": cfg.elastic_agent_version},
        {"ParameterKey": "ElasticArtifactServer", "ParameterValue": cfg.elastic_artifact_server},
        {"ParameterKey": "EnrollmentToken", "ParameterValue": cfg.enrollment_token},
        {"ParameterKey": "FleetUrl", "ParameterValue": cfg.fleet_url},
    ]

    # Capabilities
    capabilities = ["CAPABILITY_NAMED_IAM"]

    # Perform the stack update with the YAML template body
    response = cf_client.update_stack(
        StackName=cfg.stack_name,
        TemplateURL=cfg.template,
        Parameters=parameters,
        Capabilities=capabilities,
        Tags=cfg.stack_tags,
    )
    logger.info(f"Stack {response.get('StackId', 'NA')} update initiated. Waiting for update to complete...")

    # Wait until the stack update is complete
    cf_client.get_waiter("stack_update_complete").wait(StackName=cfg.stack_name)

    logger.info(f"Stack {cfg.stack_name} update is complete.")


if __name__ == "__main__":
    config = Munch()
    config.stack_name = cnfg.aws_config.cnvm_stack_name
    # Get template
    logger.info("Get AWS CNVM template")
    default_url = get_package_default_url(
        cfg=cnfg.elk_config,
        policy_name="vuln_mgmt",
        policy_type="cloudbeat/vuln_mgmt_aws",
    )
    template_url = extract_template_url(url_string=default_url)

    config.template = template_url
    config.elastic_agent_version = cnfg.elk_config.stack_version
    config.elastic_artifact_server = get_artifact_server(cnfg.elk_config.stack_version)

    # Tags for the CloudFormation stack
    stack_tags = [
        {"Key": "division", "Value": "engineering"},
        {"Key": "org", "Value": "security"},
        {"Key": "team", "Value": "cloud-security-posture"},
        {"Key": "project", "Value": "test-environments"},
    ]
    config.stack_tags = stack_tags

    # Get enrollment token
    cnvm_json = read_json(CNVM_JSON_PATH)
    config.enrollment_token = cnvm_json.get("ENROLLMENT_TOKEN", "")
    config.fleet_url = get_fleet_server_host(cfg=cnfg.elk_config)
    update_cloudformation_stack(cfg=config)
