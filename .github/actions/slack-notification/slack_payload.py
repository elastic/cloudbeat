"""
Slack Notification Script

This script processes environment variables to generate a Slack-compatible message.
It supports two modes:
1. If "MESSAGE" is set, it processes the message using the 'process_message_env' function.
2. If "PAYLOAD" is set, it directly loads the payload as JSON.

Environment Variables:
- "MESSAGE": The text content for the Slack message.
- "PAYLOAD": A JSON-formatted payload for the Slack message.
- "URL_ENCODED": If set to 'true', it URL-decodes the "MESSAGE".
- "GITHUB_OUTPUT": The file path for GitHub Actions workflow output.
- "MASK": If set to 'true', masks sensitive information in the output.

The resulting JSON data is either written to the specified output file for GitHub Actions workflow
or printed with masking if "MASK" is set to 'true'.
"""

import json
import os
from urllib.parse import unquote


class SlackException(Exception):
    """
    Custom exception class for errors related to Slack notifications.
    """


def create_message_block(message):
    """
    Creates a message block for Slack with the given text.

    Parameters:
    - message (str): The text content for the message block.

    Returns:
    dict: A dictionary representing the Slack message block in the following format:
          {
              "type": "section",
              "text": {
                  "type": "mrkdwn",
                  "text": message
              }
          }
    """
    return {
        "type": "section",
        "text": {
            "type": "mrkdwn",
            "text": message,
        },
    }


def process_message_env():
    """
    Process the 'MESSAGE' environment variable for creating a Slack-compatible message.

    If 'URL_ENCODED' is set to 'true', the message is URL-decoded.
    The message is then formatted as a Slack message containing a single text block.

    Returns:
    dict: A dictionary representing the Slack message with the processed 'MESSAGE'.
          The dictionary has the following format:
          {
              "text": processed_message,
              "blocks": [
                  {
                      "type": "section",
                      "text": {
                          "type": "mrkdwn",
                          "text": processed_message
                      }
                  }
              ]
          }
    """
    message = os.environ.get("MESSAGE", "No message")
    if os.environ["URL_ENCODED"] == "true":
        message = unquote(message)
    message = "\n".join(line.strip() for line in message.splitlines())
    return {
        "text": message,
        "blocks": [create_message_block(message)],
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


def main():
    """
    Main function to process environment variables and generate a Slack-compatible message.
    """
    if os.environ.get("MESSAGE"):
        json_data = process_message_env()
    elif os.environ.get("PAYLOAD"):
        json_data = json.loads(os.environ["PAYLOAD"])
    else:
        raise SlackException("Either message or payload must be set.")

    set_output("payload", json.dumps(json_data))

    if os.environ.get("MASK") == "true":
        print(f"::add-mask::{json.dumps(json_data)}")


if __name__ == "__main__":
    main()
