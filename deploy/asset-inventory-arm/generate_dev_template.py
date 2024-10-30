#!/usr/bin/env python
# pylint: disable=duplicate-code
"""
Generate and deploy development templates for Azure deployment.

Enables SSH access to the VMs and installs the elastic-agent with the given version and enrollment token.
"""
import argparse
import json
import os
import pathlib
import shlex
import subprocess
import sys
import time


def main():
    """
    Parse arguments and run the script.
    """
    args = parse_args(load_file_args() + sys.argv[1:])

    with open(args.template_file) as f:
        template = json.load(f)

    modify_template(template)
    with open(args.output_file, "w") as f:
        print(json.dumps(template, indent=4), file=f)  # Pretty-print the template in a JSON file.

    if args.deploy:
        if args.template_type == "organization-account":
            deploy_to_management_group(args)
        else:
            deploy_to_subscription(args)


def load_file_args():
    """
    Load extra command-line arguments from a file.
    """
    config_file = pathlib.Path(__file__).parent / "dev-flags.conf"
    if not config_file.exists():
        return []
    with open(config_file) as f:
        return shlex.split(f.read().strip())


def parse_args(argv):
    """
    Parse command-line arguments.
    :param argv: The arguments
    :return: Parsed argparse namespace
    """
    will_call_az_cli = "--deploy" in argv

    parser = argparse.ArgumentParser(description="Deploy Azure resources for a single account")
    parser.add_argument(
        "--template-type",
        help="The type of template to use",
        default="single-account",
        choices=["single-account", "organization-account"],
    )
    parser.add_argument(
        "--output-file",
        help="The output file to write the modified template to",
        default=None,  # Replace later
    )
    parser.add_argument("--deploy", help="Perform deployment", action="store_true")
    parser.add_argument(
        "--resource-group",
        help="The resource group to deploy to",
        default=f"{os.environ.get('USER', 'unknown')}-cloudbeat-dev-{int(time.time())}",
    )
    parser.add_argument("--location", help="The location to deploy to", default=os.environ.get("LOCATION", "centralus"))
    parser.add_argument("--subscription-id", help="The subscription ID to deploy to (defaults to current)")
    parser.add_argument("--management-group-id", help="The management group ID to deploy to")

    parser.add_argument("--public-ssh-key", help="SSH public key to use for the VMs", required=will_call_az_cli)
    parser.add_argument("--artifact-server", help="The URL of the artifact server", required=will_call_az_cli)
    parser.add_argument(
        "--elastic-agent-version",
        help="The version of elastic-agent to install",
        default=os.environ.get("ELK_VERSION", ""),
    )
    parser.add_argument("--fleet-url", help="The fleet URL of elastic-agent", required=will_call_az_cli)
    parser.add_argument("--enrollment-token", help="The enrollment token of elastic-agent", required=will_call_az_cli)
    args = parser.parse_args(argv)

    if args.deploy != will_call_az_cli:
        parser.error("Assertion failed: --deploy detected but parser returned different result")

    args.template_file = pathlib.Path(__file__).parent / f"ARM-for-{args.template_type}.json"
    if args.output_file is None:
        args.output_file = str(args.template_file).replace(".json", ".dev.json")
    if args.template_type == "single-account" and args.management_group_id is not None:
        parser.error("Cannot specify management group for single-account template")
    elif args.deploy and args.template_type == "organization-account" and args.management_group_id is None:
        parser.error("Must specify management group for organization-account template")

    return args


def modify_template(template):
    """
    Modify the template in-place.
    :param template: Parsed dictionary of the template
    """
    template["parameters"]["PublicKeyDevOnly"] = {
        "type": "string",
        "metadata": {"description": "The public key of the SSH key pair"},
    }

    # Shallow copy of all resources and resources of deployments
    all_resources = template["resources"][:]
    for resource in template["resources"]:
        if resource["type"] == "Microsoft.Resources/deployments":
            all_resources += resource["properties"]["template"]["resources"]
    for resource in all_resources:
        modify_resource(resource)


def modify_resource(resource):
    """
    Modify a single resource in-place.
    :param resource: Parsed dictionary of the resource
    """
    # Delete generated key pair from all dependencies
    depends_on = [d for d in resource.get("dependsOn", []) if not d.startswith("cloudbeatGenerateKeypair")]

    if resource["name"] == "cloudbeatVM":
        # Use user-provided public key
        resource["properties"]["osProfile"]["linuxConfiguration"]["ssh"]["publicKeys"] = [
            {
                "path": "/home/cloudbeat/.ssh/authorized_keys",
                "keyData": "[parameters('PublicKeyDevOnly')]",
            },
        ]
    elif resource["name"] == "cloudbeatVNet":
        # Add network security group to virtual network
        nsg_resource_id = "[resourceId('Microsoft.Network/networkSecurityGroups', 'cloudbeatNSGDevOnly')]"
        resource["properties"]["subnets"][0]["properties"]["networkSecurityGroup"] = {"id": nsg_resource_id}
        depends_on += [nsg_resource_id]
    elif resource["name"] == "cloudbeatNic":
        # Add public IP to network interface
        public_ip_resource_id = "[resourceId('Microsoft.Network/publicIPAddresses', 'cloudbeatPublicIPDevOnly')]"
        resource["properties"]["ipConfigurations"][0]["properties"]["publicIpAddress"] = {"id": public_ip_resource_id}
        depends_on += [public_ip_resource_id]
    elif resource["name"] == "cloudbeatVM/customScriptExtension":
        # Modify agent installation to *not* disable SSH
        resource["properties"]["settings"] = {
            "fileUris": ["https://raw.githubusercontent.com/elastic/cloudbeat/main/deploy/azure/install-agent-dev.sh"],
            "commandToExecute": (
                "[concat('"
                "bash install-agent-dev.sh ', "
                "parameters('ElasticAgentVersion'), ' ', "
                "parameters('ElasticArtifactServer'), ' ', "
                "parameters('FleetUrl'), ' ', "
                "parameters('EnrollmentToken'))]"
            ),
        }
    elif resource["name"] == "cloudbeat-vm-deployment":
        resource["properties"]["parameters"] = {"PublicKeyDevOnly": {"value": "[parameters('PublicKeyDevOnly')]"}}
        resource["properties"]["template"]["parameters"] = {"PublicKeyDevOnly": {"type": "string"}}
        modify_vm_deployment_template_resources_array(resource["properties"]["template"])

    if depends_on:
        resource["dependsOn"] = depends_on


def modify_vm_deployment_template_resources_array(template):
    """
    Modify the resources array of the cloudbeat VM deployment template in-place.
    :param template: Parsed dictionary of the template
    """
    template["resources"] = [
        resource
        for resource in template["resources"]
        # Delete generated key pair since we provide our own
        if resource["name"] != "cloudbeatGenerateKeypair"
    ] + [
        {
            "type": "Microsoft.Network/publicIPAddresses",
            "name": "cloudbeatPublicIpDevOnly",
            "apiVersion": "2020-05-01",
            "location": "[resourceGroup().location]",
            "properties": {"publicIPAllocationMethod": "Dynamic"},
        },
        {
            "type": "Microsoft.Network/networkSecurityGroups",
            "name": "cloudbeatNSGDevOnly",
            "apiVersion": "2021-04-01",
            "location": "[resourceGroup().location]",
            "properties": {
                "securityRules": [
                    {
                        "name": "AllowSshAll",
                        "properties": {
                            "access": "Allow",
                            "destinationAddressPrefix": "*",
                            "destinationPortRange": "22",
                            "direction": "Inbound",
                            "priority": 100,
                            "protocol": "Tcp",
                            "sourceAddressPrefix": "*",
                            "sourcePortRange": "*",
                        },
                    },
                ],
            },
        },
    ]


def deploy_to_subscription(args):
    """
    Deploy the template to a subscription.
    :param args: The parsed arguments
    """
    parameters = parameters_from_args(args)
    subscription_args = ["--subscription", args.subscription_id] if args.subscription_id else []
    subprocess.check_call(
        [
            "az",
            "group",
            "create",
            "--name",
            args.resource_group,
            "--location",
            args.location,
        ]
        + subscription_args,
    )
    subprocess.check_call(
        [
            "az",
            "deployment",
            "group",
            "create",
            "--resource-group",
            args.resource_group,
            "--template-file",
            args.output_file,
            "--parameters",
            json.dumps(parameters),
        ]
        + subscription_args,
    )


def deploy_to_management_group(args):
    """
    Deploy the template to a management group.
    :param args: The parsed arguments
    """
    parameters = parameters_from_args(args)
    parameters["parameters"]["ResourceGroupName"] = {"value": args.resource_group}
    if args.subscription_id is None:
        args.subscription_id = (
            subprocess.check_output(["az", "account", "show", "--query", "id", "-o", "tsv"])
            .decode(
                "utf-8",
            )
            .strip()
        )
    parameters["parameters"]["SubscriptionId"] = {"value": args.subscription_id}
    subprocess.check_call(
        [
            "az",
            "deployment",
            "mg",
            "create",
            "--location",
            args.location,
            "--template-file",
            args.output_file,
            "--parameters",
            json.dumps(parameters),
            "--management-group-id",
            args.management_group_id,
        ],
    )


def parameters_from_args(args):
    """
    Generate the deployment parameters file from the parsed arguments.
    :param args: The parsed arguments
    :return:
    """
    return {
        "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentParameters.json#",
        "contentVersion": "1.0.0.0",
        "parameters": {
            "ElasticArtifactServer": {"value": args.artifact_server},
            "ElasticAgentVersion": {"value": args.elastic_agent_version},
            "FleetUrl": {"value": args.fleet_url},
            "EnrollmentToken": {"value": args.enrollment_token},
            "PublicKeyDevOnly": {"value": args.public_ssh_key},
        },
    }


if __name__ == "__main__":
    main()
