#!/usr/bin/env python
"""
This script is designed to be used in a GitHub Actions workflow to send Slack notifications.
It reads environment variables set by the GitHub Actions runtime, validates them, and
constructs a Slack payload based on the workflow status.
"""
import os
import json

github_to_slack = {
    "gurevichdmitry": "U030XM1N3BP",
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
    if value is None or value == "":
        raise ValueError(f"The env var '{env_var}' isn't defined.")
    return value


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


def color_by_job_status(status: str) -> str:
    """
    Determine the Slack color based on the GitHub Actions job status.

    Parameters:
        status (str): The GitHub Actions job status, e.g., "success", "failure", etc.

    Returns:
        str: The Slack color corresponding to the GitHub Actions job status.
             Possible values: "good" for success, "danger" for failure, or an empty string.
    """
    if status == "success":
        return "good"
    if status == "failure":
        return "danger"
    return ""


class BuildSlackException(Exception):
    """
    Custom exception class for errors related to building Slack notifications.
    """


def run():
    """
    Main function to run the Slack notification workflow.

    This function is responsible for validating environment variables, generating a Slack payload,
    setting GitHub Action outputs, and handling exceptions related to building Slack notifications.
    """
    try:
        # Validate env vars
        workflow = check_env_var("WORKFLOW")
        github_actor = check_env_var("GITHUB_ACTOR")
        github_run_url = check_env_var("RUN_URL")
        job_status = check_env_var("JOB_STATUS")
        kibana_url = check_env_var("KIBANA_URL")
        s3_bucket = check_env_var("S3_BUCKET")
        deployment_name = check_env_var("DEPLOYMENT_NAME")
        stack_version = check_env_var("STACK_VERSION")
        es_password = check_env_var("ES_PASSWORD")
        ess_region = check_env_var("ESS_REGION")
        docker_image = os.getenv("DOCKER_IMAGE", "N/A")
        if docker_image == "":
            docker_image = "N/A"

        is_project = bool(os.getenv("ESS_TYPE", "true") == "true")
        ess_type = "Deployment"
        if is_project:
            ess_type = "Project"

        color = color_by_job_status(job_status)
        # Set output
        set_output("color", color)
        set_output("run_url", github_run_url)
        set_output("kibana_url", github_run_url)
        slack_name = github_to_slack.get(github_actor, github_actor)

        slack_payload = {
            "text": "Create / Destroy github workflow",
            "blocks": [
                {
                    "type": "section",
                    "text": {
                        "type": "mrkdwn",
                        "text": f"{workflow} job <{github_run_url}|{deployment_name}> by <@{slack_name}>",
                    },
                },
                {
                    "type": "divider",
                },
                {
                    "type": "section",
                    "fields": [
                        {
                            "type": "mrkdwn",
                            "text": f"*ESS Type:*\n`{ess_type}`",
                        },
                        {
                            "type": "mrkdwn",
                            "text": f"*Stack Version:*\n`{stack_version}`",
                        },
                    ],
                },
                {
                    "type": "section",
                    "fields": [
                        {
                            "type": "mrkdwn",
                            "text": f"*Cloud Region:*\n`{ess_region}`",
                        },
                        {
                            "type": "mrkdwn",
                            "text": f"*Docker Override:*\n`{docker_image}`",
                        },
                    ],
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
                                "text": "kibana",
                            },
                            "url": f"{kibana_url}",
                            "action_id": "kibana-instance-button",
                        },
                        {
                            "type": "button",
                            "text": {
                                "type": "plain_text",
                                "text": "bucket",
                            },
                            "url": f"{s3_bucket}",
                            "action_id": "s3-bucket-button",
                        },
                    ],
                },
                {
                    "type": "divider",
                },
                {
                    "type": "section",
                    "text": {
                        "type": "mrkdwn",
                        "text": f"*Kibana password:* `{es_password}`",
                    },
                },
            ],
        }
        set_output("payload", json.dumps(slack_payload))

    except BuildSlackException as err:
        set_failed(str(err))


if __name__ == "__main__":
    run()
