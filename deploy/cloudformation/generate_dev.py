"""
This module modifies the content of the production CloudFormation template into a development template.
"""

import os

PROD_TEMPLATE_PATH = "elastic-agent-ec2.yml"
DEV_TEMPLATE_PATH = "elastic-agent-ec2-dev.yml"


def edit_artifact_url(content):
    """
    Replace the production artifact URL with the snapshot artifact URL.
    """
    prod_url = "https://artifacts.elastic.co/downloads/beats/elastic-agent/"

    # TODO: Dynamically get the latest snapshot URL
    dev_url = "https://snapshots.elastic.co/8.8.0-4c45f51b/downloads/beats/elastic-agent/"
    return content.replace(prod_url, dev_url)


def main():
    """
    Read the production template, modify it, and write it to the development template.
    """
    script_path = os.path.abspath(__file__)
    current_dir = os.path.dirname(script_path)

    input_path = os.path.join(current_dir, PROD_TEMPLATE_PATH)
    output_path = os.path.join(current_dir, DEV_TEMPLATE_PATH)

    with open(input_path, "r") as file:
        file_contents = file.read()

    modified_contents = edit_artifact_url(file_contents)

    with open(output_path, "w") as file:
        file.write(modified_contents)

    print(f"Created {output_path}")


main()
