#!/usr/bin/env python
"""
This script is designed to be used in a GitHub Actions workflow to send Slack notifications.
It reads environment variables set by the GitHub Actions runtime, validates them, and
constructs a Slack payload based on the workflow status.
"""
import json
import os

color_by_job_status = {
    "success": "#36a64f",
    "failure": "#D40E0D",
}


def check_env_var(env_var: str) -> str:
    """
    Retrieve the value of the specified environment variable.

    Parameters:
        env_var (str): The name of the environment variable to retrieve.

    Returns:
        str: The value of the specified environment variable.
    """
    value = os.environ.get(env_var)
    if not value:
        if env_var == "DOCKER_IMAGE_OVERRIDE":
            return "N/A"  # If docker image is not set, slack notification will display N/A
        raise ValueError(f"The env var '{env_var}' isn't defined.")
    if env_var == "ESS_TYPE":
        return "Project" if value == "true" else "Deployment"
    return value


env_vars = {
    "workflow": check_env_var("WORKFLOW"),
    "github_actor": check_env_var("GITHUB_ACTOR"),
    "github_run_url": check_env_var("RUN_URL"),
    "job_status": check_env_var("JOB_STATUS"),
    "kibana_url": check_env_var("KIBANA_URL"),
    "s3_bucket": check_env_var("S3_BUCKET"),
    "deployment_name": check_env_var("DEPLOYMENT_NAME"),
    "stack_version": check_env_var("STACK_VERSION"),
    "docker_image": check_env_var("DOCKER_IMAGE_OVERRIDE"),
    "ess_type": check_env_var("ESS_TYPE"),
}


def set_output(name: str, value: str):
    """
    Set an output variable for the GitHub Actions workflow.

    Parameters:
        name (str): The name of the output variable.
        value (str): The value to set for the output variable.
    """
    with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as fh:
        print(f"{name}={value}", file=fh)


def set_failed(message: str):
    """
    Set the GitHub Actions workflow status to 'failed' with a specified message.

    Parameters:
        message (str): The message to be associated with the failure.
    """
    print(f"::set-failed::{message}")


def generate_slack_payload() -> dict:
    """
    Generate a Slack payload based on the provided environment variables.

    Returns:
        dict: A dictionary representing the Slack payload.
    """
    color = color_by_job_status.get(env_vars.get("job_status"), "#439FE0")
    ess_type_msg = f"*ESS Type:* `{env_vars.get('ess_type')}`"
    stack_version_msg = f"*Stack Version: *`{env_vars.get('stack_version')}`"
    docker_image_msg = f"*Docker Override:* `{env_vars.get('docker_image')}`"
    message = f"{ess_type_msg}\n{stack_version_msg}\n{docker_image_msg}"
    workflow = env_vars.get("workflow")
    deployment = env_vars.get("deployment_name")
    actor = env_vars.get("github_actor")
    title_text = f"{workflow} job `{deployment}` triggered by `{actor}`"
    docs_url = "https://github.com/elastic/cloudbeat/blob/main/dev-docs/Cloud-Env-Testing.md"
    slack_payload = {
        "text": title_text,
        "blocks": [
            {
                "type": "divider",
            },
        ],
        "attachments": [
            {
                "color": color,
                "blocks": [
                    {
                        "type": "section",
                        "text": {
                            "type": "mrkdwn",
                            "text": title_text,
                        },
                    },
                    {
                        "type": "divider",
                    },
                    {
                        "type": "section",
                        "text": {
                            "type": "mrkdwn",
                            "text": message,
                        },
                    },
                    {
                        "type": "divider",
                    },
                    {
                        "type": "actions",
                        "elements": [
                            {
                                "type": "button",
                                "text": {
                                    "type": "plain_text",
                                    "text": "kibana link",
                                },
                                "style": "primary",
                                "url": f"{env_vars.get('kibana_url')}",
                                "action_id": "kibana-instance-button",
                            },
                            {
                                "type": "button",
                                "text": {
                                    "type": "plain_text",
                                    "text": "state bucket",
                                },
                                "style": "primary",
                                "url": f"{env_vars.get('s3_bucket')}",
                                "action_id": "s3-bucket-button",
                            },
                            {
                                "type": "button",
                                "text": {
                                    "type": "plain_text",
                                    "text": "action run",
                                },
                                "style": "primary",
                                "url": f"{env_vars.get('github_run_url')}",
                                "action_id": "action-run-button",
                            },
                            {
                                "type": "button",
                                "text": {
                                    "type": "plain_text",
                                    "text": "docs",
                                },
                                "style": "primary",
                                "url": docs_url,
                                "action_id": "docs-button",
                            },
                        ],
                    },
                ],
            },
        ],
    }
    return slack_payload


def run():
    """
    Main function to run the Slack notification workflow.

    This function is responsible for validating environment variables, generating a Slack payload,
    setting GitHub Action outputs, and handling exceptions related to building Slack notifications.
    """
    try:
        slack_payload = generate_slack_payload()
        set_output("payload", json.dumps(slack_payload))
    except ValueError as err:
        set_failed(str(err))
    except TypeError as err:
        set_failed(f"Failed to serialize to JSON: {str(err)}")
    except (KeyError, FileNotFoundError, PermissionError) as err:
        set_failed(f"Failed to store GITHUB_OUTPUT: {str(err)}")


if __name__ == "__main__":
    run()
