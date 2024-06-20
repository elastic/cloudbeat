#!/usr/bin/env python
"""
Script: create_env_config.py

Description:
This script generates an environment configuration JSON file (`env_config.json`)
based on provided deployment name and expiration days retrieved from environment
variables (`DEPLOYMENT_NAME` and `EXPIRATION_DAYS`).

Usage:
Ensure `DEPLOYMENT_NAME` and `EXPIRATION_DAYS` environment variables are set before
running the script. It calculates the expiration date by adding `EXPIRATION_DAYS`
to the current date and saves the configuration to `env_config.json` in the current
directory.

Example:
$ export DEPLOYMENT_NAME="my_deploy"
$ export EXPIRATION_DAYS="14"
$ python create_env_config.py

Output:
Creates a JSON file `env_config.json` with content like:
{
    "deployment_name": "my_deploy",
    "expiration": "yyyy-mm-dd"
}

"""
import json
import os
import sys
from datetime import datetime, timedelta

# Define the output file path
ENV_CONFIG_FILE = "env_config.json"


def create_env_config(deploy_name, expire_days):
    """
    Create environment configuration dictionary.

    Args:
        deploy_name (str): The name of the deployment.
        expire_days (str): The number of days until expiration.

    Returns:
        dict: The environment configuration dictionary.
    """
    # Calculate the expiration date
    current_date = datetime.now()
    expiration_date = current_date + timedelta(days=int(expire_days))

    # Format expiration date as yyyy-mm-dd
    expiration_date_str = expiration_date.strftime("%Y-%m-%d")

    return {
        "deployment_name": deploy_name,
        "expiration": expiration_date_str,
    }


if __name__ == "__main__":
    # Retrieve deployment_name and expiration_days from environment variables
    deployment_name = os.getenv("DEPLOYMENT_NAME")
    expiration_days = os.getenv("EXPIRATION_DAYS")

    if not deployment_name or not expiration_days:
        print("Error: DEPLOYMENT_NAME or EXPIRATION_DAYS environment variables not set.")
        sys.exit(1)

    # Create environment configuration
    env_config = create_env_config(deployment_name, expiration_days)

    # Save to JSON file
    with open(ENV_CONFIG_FILE, "w", encoding="utf-8") as f:
        json.dump(env_config, f, indent=4)

    print(f"Saved environment configuration to {ENV_CONFIG_FILE}")
