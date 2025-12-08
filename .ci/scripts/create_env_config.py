#!/usr/bin/env python
"""
Script: create_env_config.py

Description:
This script generates an environment configuration JSON file (`env_config.json`)
based on provided deployment name, expiration days, ess-region, and other deployment
configuration retrieved from environment variables.

Usage:
Ensure `DEPLOYMENT_NAME` and `EXPIRATION_DAYS` environment variables are set before
running the script. Other variables are optional and have defaults.
It calculates the expiration date by adding `EXPIRATION_DAYS` to the current date
and saves the configuration to `env_config.json` in the current directory.

Example:
$ export DEPLOYMENT_NAME="my_deploy"
$ export EXPIRATION_DAYS="14"
$ export ESS_REGION="production-cft"  # Optional, defaults to "production-cft"
$ export ESS_REGION_MAPPED="gcp-us-west2"  # Optional
$ export EC_URL="https://cloud.elastic.co"  # Optional
$ export SERVERLESS_MODE="false"  # Optional
$ python create_env_config.py

Output:
Creates a JSON file `env_config.json` with content like:
{
    "deployment_name": "my_deploy",
    "expiration": "yyyy-mm-dd",
    "ess_region": "production-cft",
    "ess_region_mapped": "gcp-us-west2",
    "ec_url": "https://cloud.elastic.co",
    "serverless_mode": "false"
}

"""
import json
import os
import sys
from datetime import datetime, timedelta

# Define the output file path
ENV_CONFIG_FILE = "env_config.json"


def create_env_config(
    deploy_name,
    expire_days,
    ess_region_input=None,
    ess_region_mapped_val=None,
    ec_url_val=None,
    serverless_mode_val=None,
):
    """
    Create environment configuration dictionary.

    Args:
        deploy_name (str): The name of the deployment.
        expire_days (str): The number of days until expiration.
        ess_region_input (str, optional): The ess-region input format
            (e.g., "production-cft"). Defaults to "production-cft".
        ess_region_mapped_val (str, optional): The mapped ESS region
            (e.g., "gcp-us-west2").
        ec_url_val (str, optional): The Elastic Cloud URL.
        serverless_mode_val (str, optional): Whether deployment is serverless ("true" or "false").

    Returns:
        dict: The environment configuration dictionary.
    """
    # Calculate the expiration date
    current_date = datetime.now()
    expiration_date = current_date + timedelta(days=int(expire_days))

    # Format expiration date as yyyy-mm-dd
    expiration_date_str = expiration_date.strftime("%Y-%m-%d")

    # Default to production-cft if not provided
    ess_region = ess_region_input or "production-cft"

    config = {
        "deployment_name": deploy_name,
        "expiration": expiration_date_str,
        "ess_region": ess_region,
    }

    # Add optional fields if provided
    if ess_region_mapped_val:
        config["ess_region_mapped"] = ess_region_mapped_val
    if ec_url_val:
        config["ec_url"] = ec_url_val
    if serverless_mode_val is not None:
        config["serverless_mode"] = serverless_mode_val

    return config


if __name__ == "__main__":
    # Retrieve deployment_name and expiration_days from environment variables
    deployment_name = os.getenv("DEPLOYMENT_NAME")
    expiration_days = os.getenv("EXPIRATION_DAYS")
    ess_region_env = os.getenv("ESS_REGION", "production-cft")  # Default to production-cft
    ess_region_mapped = os.getenv("ESS_REGION_MAPPED")  # Optional
    ec_url = os.getenv("EC_URL")  # Optional
    serverless_mode = os.getenv("SERVERLESS_MODE")  # Optional

    if not deployment_name or not expiration_days:
        print("Error: DEPLOYMENT_NAME or EXPIRATION_DAYS environment variables not set.")
        sys.exit(1)

    # Create environment configuration
    env_config = create_env_config(
        deployment_name,
        expiration_days,
        ess_region_env,
        ess_region_mapped,
        ec_url,
        serverless_mode,
    )

    # Save to JSON file
    with open(ENV_CONFIG_FILE, "w", encoding="utf-8") as f:
        json.dump(env_config, f, indent=4)

    print(f"Saved environment configuration to {ENV_CONFIG_FILE}")
