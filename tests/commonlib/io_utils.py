"""
This module provides input / output manipulations on streams / files
"""


import io
import json
import os
import shutil
import subprocess

from datetime import datetime
from pathlib import Path

import yaml
from munch import Munch, munchify


def get_events_from_index(
    elastic_client,
    index_name: str,
    rule_tag: str,
    time_after: datetime,
) -> list[Munch]:
    """
    This function returns events from the given index,
    filtering for the given rule_tag, and after the given time.
    @param elastic_client: Client to connect to Elasticsearch.
    @param index_name: Index to get events from.
    @param rule_tag: Rule tag to filter.
    @param time_after: Filter events having timestamp > time_after
    @return: List of Munch objects
    """
    query = {
        "bool": {
            "must": [{"match": {"rule.tags": rule_tag}}],
            "filter": [
                {
                    "range": {
                        "@timestamp": {
                            "gte": time_after.strftime("%Y-%m-%dT%H:%M:%S.%f"),
                        },
                    },
                },
            ],
        },
    }
    sort = [{"@timestamp": {"order": "desc"}}]
    result = elastic_client.get_index_data(
        index_name=index_name,
        query=query,
        sort=sort,
        size=1000,
    )

    events = []
    for event in munchify(dict(result)).hits.hits:
        events.append(event._source)

    return events


def get_logs_from_stream(stream: str) -> list[Munch]:
    """
    This function converts logs stream to list of Munch objects (dictionaries)
    @param stream: StringIO stream
    @return: List of Munch objects
    """
    logs = io.StringIO(stream)
    result = []
    for log in logs:
        if log and "bundles" in log:
            try:
                result.append(munchify(json.loads(log)))
            except json.decoder.JSONDecodeError:
                result.append(munchify(json.loads(log.replace("'", '"'))))
            except AttributeError as exc:
                print(exc)
                continue

    return result


def get_k8s_yaml_objects(file_path: Path) -> list[str:dict]:
    """
    This function loads yaml file, and returns the following list:
    [ {<k8s_kind> : {<k8s_metadata}}]
    :param file_path: YAML path
    :return: [ {<k8s_kind> : {<k8s_metadata}}]
    """
    if not file_path:
        raise Exception(f"{file_path} is required")
    with file_path.open(encoding="utf-8") as yaml_file:
        return list(yaml.safe_load_all(yaml_file))


class FsClient:
    """
    This class provides functionality for working with
    file system operations
    """

    @staticmethod
    def exec_command(
        container_name: str,
        command: str,
        param_value: str,
        resource: str,
    ):
        """
        This function executes os command
        @param container_name: Container node
        @param command: Linux command to be executed
        @param param_value: Value to be used in exec command
        @param resource: File / Resource path
        @return: None
        """

        if command == "touch":
            if os.path.exists(param_value):
                return
            with open(param_value, "a+", encoding="utf-8"):
                pass
            return

        if command == "cat":
            with open(resource, "w", encoding="utf-8") as file:
                file.write(param_value)
            return

        if container_name == "":
            raise Exception("Unknown container name is sent")

        current_resource = Path(resource)
        if not (current_resource.is_file() or current_resource.is_dir()):
            raise Exception(f"File {resource} does not exist or mount missing.")

        if command == "chmod":
            os.chmod(path=resource, mode=int(param_value, base=8))
        elif command == "chown":
            try:
                uid, gid = param_value.split(":")
            except ValueError as exc:
                raise Exception(
                    "User and group parameter shall be separated by ':' ",
                ) from exc

            FsClient.add_users_to_node([uid, gid], in_place=True)
            shutil.chown(path=resource, user=uid, group=gid)
        elif command == "unlink":
            if not Path(param_value).is_dir():
                Path(param_value).unlink()
        else:
            raise Exception(
                f"Command '{command}' still not implemented in test framework",
            )

    @staticmethod
    def add_users_to_node(users: list, in_place: bool):
        """
        This function creates the given users along with groups with the
        same name, on the local container as well the host node.
        @param users: List of users to create.
        @param in_place: Whether host node configuration files should be modified in-place or overwritten.
        @return: None
        """
        if in_place:
            host_users_file = Path("/hostfs/etc/passwd")
            host_groups_file = Path("/hostfs/etc/group")

            temp_etc = Path("/tmp/etc")
            temp_etc.mkdir(parents=True, exist_ok=True)

            temp_users_file = temp_etc / "passwd"
            temp_groups_file = temp_etc / "group"

            shutil.copyfile(host_users_file, temp_users_file)
            shutil.copyfile(host_groups_file, temp_groups_file)

            for user in users:
                # These commands fail silently for users/groups that exist.
                subprocess.run(
                    ["groupadd", user, "-P", "/tmp"],
                    capture_output=True,
                    check=False,
                )
                subprocess.run(
                    ["useradd", user, "-g", user, "-P", "/tmp"],
                    capture_output=True,
                    check=False,
                )
                subprocess.run(
                    ["useradd", user],
                    capture_output=True,
                    check=False,
                )  # For container to get around chmod check.

            FsClient.in_place_copy(temp_users_file, host_users_file)
            FsClient.in_place_copy(temp_groups_file, host_groups_file)

        else:
            # TODO(yashtewari): Implement this section which simulates a "normal" user flow
            # where useradd command overwrites passwd and group files,
            # as part of tests for: https://github.com/elastic/cloudbeat/issues/235
            pass

    @staticmethod
    def in_place_copy(source, destination):
        """
        Copy the contents of source into destination without overwriting the destination file.
        """
        with open(source, "r", encoding="utf-8") as sfile, open(
            destination,
            "w",
            encoding="utf-8",
        ) as dfile:
            for line in sfile:
                dfile.write(line)

    @staticmethod
    def edit_process_file(container_name: str, dictionary, resource: str):
        """
        This function edits a process file
        @param container_name: Container node
        @param dictionary: Process parameters to set/unset
        @param resource: File / Resource path
        @return: None
        """
        if container_name == "":
            raise Exception("Unknown container name is sent")

        current_resource = Path(resource)
        if not current_resource.is_file():
            raise Exception(f"File {resource} does not exist or mount missing.")

        # Open and load the YAML into variable
        with current_resource.open(encoding="utf-8") as file:
            r_file = yaml.safe_load(file)

        # Get process configuration arguments
        command = r_file["spec"]["containers"][0]["command"]

        # Collect set/unset keys and values from the dictionary
        set_dict = dictionary.get("set", {})
        unset_list = dictionary.get("unset", [])

        # Cycle across set items from the dictionary
        for s_key, s_value in set_dict.items():
            # Find if set key exists already in the configuration arguments
            if any(s_key == x.split("=")[0] for x in command):
                # Replace the value of the key with the new value from the set items
                command = list(
                    map(
                        lambda x, s_key, s_value: x.replace(x, s_key + "=" + s_value)
                        if s_key == x.split("=")[0]
                        else x,
                        command,
                    ),
                )
            else:
                # In case of non-existing key in the configuration arguments,
                # append the key/value from set items
                command.append(s_key + "=" + s_value)

        # Cycle across unset items from the dictionary
        for us_key in unset_list:
            # Filter out the unset keys from the configuration arguments
            command = [x for x in command if us_key != x.split("=")[0]]

        # Override the configuration arguments with the newly built configuration arguments
        r_file["spec"]["containers"][0]["command"] = command

        # Write the newly built configuration arguments
        with current_resource.open(mode="w", encoding="utf-8") as file:
            yaml.dump(r_file, file)

    @staticmethod
    def edit_config_file(container_name: str, dictionary, resource: str):
        """
        This function edits a config file
        @param container_name: Container node
        @param dictionary: Config parameters to set/unset
        @param resource: Config path
        @return: None
        """
        if container_name == "":
            raise Exception("Unknown container name is sent")

        current_resource = Path(resource)
        if not current_resource.is_file():
            raise Exception(f"File {resource} does not exist or mount missing.")

        # Open and load the YAML into variable
        with current_resource.open(encoding="utf-8") as file:
            r_file = yaml.safe_load(file)

        # Collect set/unset keys and values from the dictionary
        set_dict = dictionary.get("set", {})
        unset_list = dictionary.get("unset", [])

        # Merge two dictionaries with priority for the set items
        r_file = {**r_file, **set_dict}

        # Cycle across unset items from the dictionary
        for us_key in unset_list:
            # Parsed dot separated key values
            keys = us_key.split(".")
            key_to_del = keys.pop()
            r_dict = r_file

            # Advance inside the dictionary for nested keys
            for key in keys:
                r_dict = r_dict.get(key, None)
                if r_dict is None:
                    # Non-existing nested key
                    break
            # Remove nested keys when all path exists
            if r_dict:
                del r_dict[key_to_del]
        # Write the newly built config
        with current_resource.open(mode="w", encoding="utf-8") as file:
            yaml.dump(r_file, file)

    @staticmethod
    def get_beat_status_from_json(response: str, beat_name: str) -> str:
        """
        This function parses status response json retrieved as json and
        returns information from Application.Message field.
        @param response: Elastic-agent status string (param --output json)
        @param beat_name: The name of beat the status should be retrieved
        @return: status message string
        """
        response = json.loads(response)
        beat_list = response["Applications"]
        for beat in beat_list:
            if beat["Name"] == beat_name:
                return beat["Message"]
        return ""
