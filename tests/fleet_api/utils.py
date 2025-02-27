"""
This module provides utility functions for working with JSON files and performing file operations.

Functions:
- read_json(json_path: Path) -> dict: Read JSON data from a file.
- save_state(file_path: Path, data: list) -> None: Save data to a JSON file.
- delete_file(file_path: Path): Delete a file.

Module Dependencies:
- json: Standard library module for working with JSON data.
- pathlib.Path: Class representing file system paths.
- loguru.logger: Logger object for logging messages.

Usage:
Import this module to utilize the provided functions for JSON file operations and file deletion.
"""

import json
import sys
from pathlib import Path
from typing import Union

import ruamel.yaml
from jinja2 import Template
from loguru import logger


def read_json(json_path: Path) -> dict:
    """
    Read JSON data from a file. Exits if the file is not found or an error occurs while reading the file.

    Args:
        json_path (Path): Path to the JSON file.

    Returns:
        dict: Dictionary containing the JSON data.
    """

    try:
        with json_path.open("r") as json_file:
            return json.load(json_file)
    except FileNotFoundError:
        logger.error(f"{json_path.name} file not found.")
        sys.exit(1)
    except json.JSONDecodeError as ex:
        logger.error(f"Error reading file {json_path}: {ex}")
        sys.exit(1)


def delete_file(file_path: Path):
    """
    Delete a file.

    Args:
        file_path (Path): Path to the file to be deleted.

    Raises:
        FileNotFoundError: If the specified file does not exist.
        Exception: If an error occurs while deleting the file.
    """
    try:
        file_path.unlink()
        logger.info(f"File '{file_path}' deleted successfully.")
    except FileNotFoundError:
        logger.warning(f"File '{file_path}' does not exist.")
    except OSError as ex:
        logger.error(f"Error occurred while deleting file '{file_path}': {ex}")
        raise ex


def update_key(data: Union[dict, list], search_key: str, value_to_apply: str):
    """Update the value of a specific key in the given data.

    If the data is a dictionary, it searches for the specified key and updates its value.
    If the data is a list, it recursively calls the function for each item in the list.

    Args:
        data (Union[dict,list]): The data to be updated, which can be either a dictionary or a list.
        search_key (str): The key to search for in the data.
        value_to_apply (str): The value to be applied to the matching key.

    Returns:
        None
    """
    if isinstance(data, list):
        for item in data:
            update_key(item, search_key, value_to_apply)
    elif isinstance(data, dict):
        for key, value in data.items():
            if key == search_key:
                data[key]["value"] = value_to_apply
            elif isinstance(value, (dict, list)):
                update_key(value, search_key, value_to_apply)


def update_key_value(data: Union[dict, list], search_key: str, value_to_apply: str):
    """Update the value of a specific key in the given data.

    If the data is a dictionary, it searches for the specified key and updates its value.
    If the data is a list, it recursively calls the function for each item in the list.

    Args:
        data (Union[dict,list]): The data to be updated, which can be either a dictionary or a list.
        search_key (str): The key to search for in the data.
        value_to_apply (str): The value to be applied to the matching key.

    Returns:
        None
    """
    if isinstance(data, list):
        for item in data:
            update_key_value(item, search_key, value_to_apply)
    elif isinstance(data, dict):
        for key, value in data.items():
            if key == search_key:
                data[key] = value_to_apply
            elif isinstance(value, (dict, list)):
                update_key_value(value, search_key, value_to_apply)


def delete_key(data: Union[dict, list], search_key: str, key_to_delete: str):
    """Delete a specific key from the data if it matches the search key.

    If the data is a dictionary,
    it searches for the specified key and deletes the corresponding key_to_delete.

    If the data is a list, it recursively calls the function for each item in the list.

    Args:
        data (Union[dict, list]): The data from which the key should be deleted,
                                    which can be either a dictionary or a list.
        search_key (str): The key to search for in the data.
        key_to_delete (str): The key to delete if it matches the search key.

    Returns:
        None
    """
    if isinstance(data, list):
        for item in data:
            delete_key(item, search_key, key_to_delete)
    elif isinstance(data, dict):
        for key, value in data.items():
            if key == search_key and isinstance(value, dict) and "value" in value:
                del value[key_to_delete]
            elif isinstance(value, (dict, list)):
                delete_key(value, search_key, key_to_delete)


def render_template(template_path, replacements):
    """
    Render a template file with the provided replacements.

    Args:
        template_path (str): The path to the template file.
        replacements (dict): A dictionary containing the replacements to be applied to the template.

    Returns:
        str: The rendered content of the template file with replacements applied.

    Raises:
        FileNotFoundError: If the template file specified by `template_path` does not exist.
        IOError: If there is an error reading the template file.

    """
    with open(template_path, "r", encoding="utf-8") as t_file:
        template_content = t_file.read()

    template = Template(template_content)
    rendered_content = template.render(replacements)

    return rendered_content


def replace_image_recursive(data, new_image: str):
    """
    Recursively searches for the 'image' field in the YAML data and replaces its value.

    Args:
        data (Union[CommentedMap, list]): The YAML data to be processed.
        new_image (str): The new image value to replace the existing one.

    Returns:
        None
    """
    if isinstance(data, ruamel.yaml.comments.CommentedMap):
        for key in data:
            if key == "image":
                data[key] = new_image
            else:
                replace_image_recursive(data[key], new_image)
    elif isinstance(data, list):
        for item in data:
            replace_image_recursive(item, new_image)


def replace_image_field(yaml_string: str, new_image: str) -> str:
    """
    Replaces the value of the 'image' field in the provided YAML string with a new image value.

    Args:
        yaml_string (str): The YAML string to be processed.
        new_image (str): The new image value to replace the existing one.

    Returns:
        str: The modified YAML string with the updated 'image' field.
    """
    yaml = ruamel.yaml.YAML()
    yaml.preserve_quotes = True
    yaml.indent(mapping=2, sequence=4, offset=2)
    yaml.explicit_start = True

    output = []
    for doc in yaml.load_all(yaml_string):
        replace_image_recursive(doc, new_image)
        if doc:
            output.append(doc)

    # Create an output stream
    output_stream = ruamel.yaml.compat.StringIO()

    # Dump the modified YAML data to the output stream
    yaml.dump_all(output, output_stream)

    # Get the YAML string from the output stream
    yaml_string = output_stream.getvalue()

    return yaml_string


def add_capabilities(yaml_content: str) -> str:
    """
    Adds capabilities to the 'securityContext' if not already present.

    Args:
        yaml_content (str): The YAML content (collection of documents) to be processed.

    Returns:
        str: The modified YAML content with added capabilities.
    """
    yaml = ruamel.yaml.YAML()
    yaml.preserve_quotes = True
    yaml.indent(mapping=2, sequence=4, offset=2)
    yaml.explicit_start = True

    output = []
    # Process each document individually and add capabilities
    documents = list(yaml.load_all(yaml_content))
    for doc in documents:
        if isinstance(doc, dict) and doc.get("kind") == "DaemonSet":
            containers = doc["spec"]["template"]["spec"].get("containers", [])
            for container in containers:
                security_context = container.setdefault("securityContext", {})
                capabilities = security_context.setdefault("capabilities", {})
                add_list = capabilities.setdefault("add", [])
                capabilities_to_add = ["BPF", "PERFMON", "SYS_RESOURCE"]
                for cap in capabilities_to_add:
                    if cap not in add_list:
                        add_list.append(cap)
        if doc:
            output.append(doc)

    # Create an output stream
    output_stream = ruamel.yaml.compat.StringIO()

    # Dump the modified YAML data to the output stream
    yaml.dump_all(output, output_stream)

    # Get the YAML string from the output stream
    modified_content = output_stream.getvalue()

    return modified_content


def rename_file_by_suffix(file_path: Path, suffix: str) -> None:
    """
    Rename a file by adding a specified suffix to its filename.

    Args:
        file_path (Path): The path to the file to be renamed.
        suffix (str): The suffix to be added to the filename.

    Returns:
        None
    """
    if not file_path.exists():
        logger.warning(f"File {file_path.name} not found")
        return

    try:
        new_name = f"{file_path.stem}{suffix}{file_path.suffix}"
        new_file_path = file_path.parent / new_name
        Path(file_path).rename(new_file_path)
    except FileNotFoundError:
        logger.warning(f"File {file_path.name} not found")
    except FileExistsError:
        logger.warning(f"File {new_file_path} already exists")


def add_tags(tags: str, yaml_content: str):
    """
    Add custom tags to a YAML content while preserving formatting.

    Args:
        tags (str): Custom tags in the format "key1=value1 key2=value2 ...".
        yaml_content (str): YAML content to which custom tags will be added.

    Returns:
        str: The modified YAML content with custom tags.
    """
    # Create a ruamel.yaml instance with the ability to preserve formatting
    yaml = ruamel.yaml.YAML()
    yaml.preserve_quotes = True
    yaml.explicit_start = True
    yaml.indent(mapping=2, sequence=4, offset=2)

    cnvm_template = yaml.load(yaml_content)

    # Get custom tags from the input argument
    custom_tags = tags.split()
    tag_dicts = []

    for tag in custom_tags:
        key_values = tag.split(",")
        tag_dict = {}

        for key_value in key_values:
            key, value = key_value.split("=")
            tag_dict[key] = value
        tag_dicts.append(tag_dict)

    for resource in cnvm_template["Resources"].values():
        if resource["Type"] == "AWS::EC2::Instance":
            if "Properties" not in resource:
                resource["Properties"] = {}
            if "Tags" not in resource["Properties"]:
                resource["Properties"]["Tags"] = []
            resource["Properties"]["Tags"] += tag_dicts

    # Create an output stream
    output_stream = ruamel.yaml.compat.StringIO()

    # Dump the modified YAML data to the output stream
    yaml.dump(cnvm_template, output_stream)

    # Get the YAML string from the output stream
    modified_content = output_stream.getvalue()

    return modified_content


def get_install_servers_option(stack_version: str):
    """
    Returns "--install-servers" if stack_version starts with "9.", otherwise returns None.
    """
    if stack_version.startswith("9."):
        return "--install-servers"
    return None
