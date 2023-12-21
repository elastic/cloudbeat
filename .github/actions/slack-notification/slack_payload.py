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

    Multiline Messages:
        GITHUB_OUTPUT is utilized for output messages for Slack notification.
        The challenge arises when the message is a multiline string.
        When attempting to write to GITHUB_OUTPUT, GitHub Actions notifies about formatting problems with the string.
        There are several solutions to work around this issue.
        One approach is to use delimiters between lines, as suggested by Github Workflows.
        https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#multiline-strings
        THe GitHub solution warns that the delimiter chosen should not occur on a line of its own within the value,
        and this can be problematic for arbitrary multiline messages.
        To avoid this, another solution involves encoding the string.
        This function supports handling of the string encoding.

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
    message = os.environ.get("MESSAGE")
    if message:
        message = replace_user_mentions(message)
        if os.environ["URL_ENCODED"] == "true":
            # URL decode if necessary (used as a handler for multiline messages stored in GITHUB_OUTPUT)
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


def replace_user_mentions(message):
    """
    Replace github user mentions in the message with the corresponding Slack user IDs.

    Parameters:
    - message (str): The message to be processed.

    Returns:
    str: The message with user mentions replaced with Slack user IDs.
    """
    github_user = os.environ.get("GITHUB_ACTOR")
    if not github_user:
        print("::warning::GitHub user information is missing. Falling back to the original message")
        return message
    slack_user_id = os.environ.get("SLACK_USER")
    if not slack_user_id:
        print("::warning::Slack user ID is missing. Falling back to the original message")
        return message
    return message.replace(f"{github_user}", f"<@{slack_user_id}>")


def main():
    """
    Main function to process environment variables and generate a Slack-compatible message.
    """
    try:
        if os.environ.get("MESSAGE"):
            json_data = process_message_env()
        elif os.environ.get("PAYLOAD"):
            payload = replace_user_mentions(os.environ["PAYLOAD"])
            json_data = json.loads(payload)
        else:
            raise ValueError("Either message or payload must be set.")

        set_output("payload", json.dumps(json_data))

        if os.environ.get("MASK") == "true":
            print(f"::add-mask::{json.dumps(json_data)}")
    except TypeError as err:
        print(f"::set-failed::Failed to serialize to JSON: {str(err)}")
    except json.JSONDecodeError as err:
        print(f"::set-failed::Failed to deserialize: {str(err)}")
    except ValueError as err:
        print(f"::set-failed::{str(err)}")


if __name__ == "__main__":
    main()
