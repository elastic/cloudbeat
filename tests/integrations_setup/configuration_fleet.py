"""
This module provides configuration settings and paths
for the ELK (Elasticsearch, Logstash, Kibana) integration.

Module contents:
    - elk_config: Munch object containing the ELK configuration settings.
    - state_data_file: Path object representing the path to the state data file.

Dependencies:
    - os: Module for accessing environment variables.
    - pathlib: Module for working with file paths.
    - munch: Module for creating convenient data containers.

Note: This module assumes that environment variables for
the ELK configuration (ES_USER, ES_PASSWORD, KIBANA_URL)
have been set in the system environment.
"""

import os

from munch import Munch

# CNVM_TAGS format: "Key=<key1>,Value=<value1> Key=<key2>,Value=<value2> ..."
# Note: Each key-value pair is separated by a space. This space is required and used in the add_tags function.
CNVM_TAGS = (
    "Key=division,Value=engineering "
    "Key=org,Value=security "
    "Key=team,Value=cloud-security-posture "
    "Key=project,Value=test-environments"
)

DOCKER_ARTIFACTORY_AGENT = "docker.elastic.co/elastic-agent/elastic-agent"

elk_config = Munch()
elk_config.user = os.getenv("ES_USER", "NA")
elk_config.password = os.getenv("ES_PASSWORD", "NA")
elk_config.kibana_url = os.getenv("KIBANA_URL", "")
elk_config.stack_version = os.getenv("STACK_VERSION", "NA")
elk_config.agent_version = os.getenv("AGENT_VERSION", "")
elk_config.auth = (elk_config.user, elk_config.password)

kspm_config = Munch()
kspm_config.docker_image_override = os.getenv("DOCKER_IMAGE_OVERRIDE", "")
if not kspm_config.docker_image_override and elk_config.stack_version.endswith("SNAPSHOT"):
    kspm_config.docker_image_override = f"{DOCKER_ARTIFACTORY_AGENT}:{elk_config.stack_version}"
    if elk_config.agent_version:
        kspm_config.docker_image_override = f"{DOCKER_ARTIFACTORY_AGENT}:{elk_config.agent_version}"

aws_config = Munch()
aws_config.access_key_id = os.getenv("AWS_ACCESS_KEY_ID", "NA")
aws_config.secret_access_key = os.getenv("AWS_SECRET_ACCESS_KEY", "NA")
aws_config.cnvm_tags = os.getenv("AWS_CNVM_TAGS", CNVM_TAGS)
aws_config.cnvm_stack_name = os.getenv("CNVM_STACK_NAME", "NA")
aws_config.cloudtrail_s3 = os.getenv("CLOUDTRAIL_S3", "NA")

gcp_dm_config = Munch()
gcp_dm_config.deployment_name = os.getenv("DEPLOYMENT_NAME", "")
gcp_dm_config.zone = os.getenv("ZONE", "us-central1-a")
gcp_dm_config.allow_ssh = os.getenv("ALLOW_SSH", "false") == "true"
gcp_dm_config.credentials_file = os.getenv("GOOGLE_APPLICATION_CREDENTIALS", "")
gcp_dm_config.service_account_json_path = os.getenv("SERVICE_ACCOUNT_JSON_PATH", "")
gcp_dm_config.project_id = os.getenv("GOOGLE_CLOUD_PROJECT", "")

gcp_audit_config = Munch()
gcp_audit_config.credentials_file = os.getenv("GOOGLE_APPLICATION_CREDENTIALS", "")
gcp_audit_config.project_id = os.getenv("GOOGLE_CLOUD_PROJECT", "")
gcp_audit_config.topic_name = os.getenv("GCP_TOPIC_NAME", "")
gcp_audit_config.subscription_name = os.getenv("GCP_SUBSCRIPTION_NAME", "")

# Used for Azure deployment on stack 8.11.* (1.6.* package version)
azure_arm_parameters = Munch()
azure_arm_parameters.deployment_name = os.getenv("DEPLOYMENT_NAME", "")
azure_arm_parameters.location = os.getenv("LOCATION", "CentralUS")
azure_arm_parameters.credentials = os.getenv("AZURE_CREDENTIALS", "")

# Used for Azure Logs integration
azure_config = Munch()
azure_config.eventhub = os.getenv("EVENTHUB", "")
azure_config.consumer_group = os.getenv("CONSUMER_GROUP", "$Default")
azure_config.connection_string = os.getenv("CONNECTION_STRING", "")
azure_config.storage_account = os.getenv("STORAGE_ACCOUNT", "")
azure_config.storage_account_key = os.getenv("STORAGE_ACCOUNT_KEY", "")

# Used for WIZ integration
wiz_config = Munch()
wiz_config.client_id = os.getenv("WIZ_CLIENT_ID", "")
wiz_config.client_secret = os.getenv("WIZ_CLIENT_SECRET", "")
wiz_config.url = os.getenv("WIZ_URL", "")
wiz_config.token_url = os.getenv("WIZ_TOKEN_URL", "")
