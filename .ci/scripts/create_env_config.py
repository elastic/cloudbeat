#!/usr/bin/env python
"""
Script: create_env_config.py

Description:
This script generates an environment configuration JSON file (`env_config.json`)
based on provided deployment name, expiration days, and ess-region retrieved from
environment variables (`DEPLOYMENT_NAME`, `EXPIRATION_DAYS`, and `ESS_REGION`).

Usage:
Ensure `DEPLOYMENT_NAME` and `EXPIRATION_DAYS` environment variables are set before
running the script. `ESS_REGION` is optional and defaults to "production-cft" if not provided.
It calculates the expiration date by adding `EXPIRATION_DAYS` to the current date
and saves the configuration to `env_config.json` in the current directory.

Example:
$ export DEPLOYMENT_NAME="my_deploy"
$ export EXPIRATION_DAYS="14"
$ export ESS_REGION="production-cft"  # Optional, defaults to "production-cft"
$ python create_env_config.py

Output:
Creates a JSON file `env_config.json` with content like:
{
    "deployment_name": "my_deploy",
    "expiration": "yyyy-mm-dd",
    "ess_region": "production-cft"
}

"""
import json
import os
import sys
from datetime import datetime, timedelta

# Define the output file path
ENV_CONFIG_FILE = "env_config.json"


def create_env_config(deploy_name, expire_days, ess_region_input=None):
    """
    Create environment configuration dictionary.

    Args:
        deploy_name (str): The name of the deployment.
        expire_days (str): The number of days until expiration.
        ess_region_input (str, optional): The ess-region used for deployment. Defaults to "production-cft".

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

    return {
        "deployment_name": deploy_name,
        "expiration": expiration_date_str,
        "ess_region": ess_region,
    }


if __name__ == "__main__":
    # Retrieve deployment_name and expiration_days from environment variables
    deployment_name = os.getenv("DEPLOYMENT_NAME")
    expiration_days = os.getenv("EXPIRATION_DAYS")
    ess_region_env = os.getenv("ESS_REGION", "production-cft")  # Default to production-cft

    if not deployment_name or not expiration_days:
        print("Error: DEPLOYMENT_NAME or EXPIRATION_DAYS environment variables not set.")
        sys.exit(1)

    # Create environment configuration
    env_config = create_env_config(deployment_name, expiration_days, ess_region_env)

    # Save to JSON file
    with open(ENV_CONFIG_FILE, "w", encoding="utf-8") as f:
        json.dump(env_config, f, indent=4)

    print(f"Saved environment configuration to {ENV_CONFIG_FILE}")
