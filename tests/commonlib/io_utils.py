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
from loguru import logger
from munch import Munch, munchify


def get_events_from_index(
    elastic_client,
    rule_tag: str,
    time_after: datetime,
) -> list[Munch]:
    """
    This function returns events from the given index,
    filtering for the given rule_tag, and after the given time.
    @param elastic_client: Client to connect to Elasticsearch.
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
        query=query,
        sort=sort,
        size=1000,
    )

    events = []
    for event in munchify(dict(result)).hits.hits:
        events.append(event._source)

    return events

def get_assets_from_index(
    elastic_client,
    category: str, sub_category: str, type_: str, sub_type: str,
    time_after: datetime,
) -> list[Munch]:
    """
    TODO(kuba) Update the docstring
    """
    query = {
        "bool": {
            "must": [
                {"match": {"asset.category": category}},
                {"match": {"asset.sub_category": sub_category}},
                {"match": {"asset.type": type_}},
                {"match": {"asset.sub_type": sub_type}},
            ],
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
        query=query,
        sort=sort,
        size=1000,
    )

    assets = []
    for asset in munchify(dict(result)).hits.hits:
        assets.append(asset._source)

    return assets


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
                logger.warning(exc)
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
        raise ValueError(f"{file_path} is required")
    with file_path.open(encoding="utf-8") as yaml_file:
        return list(yaml.safe_load_all(yaml_file))


class FsClient:
    """
    This class provides functionality for working with
    file system operations
    """

    @staticmethod
    def exec_command(  # noqa: C901
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

        if command == "mkdir":
            os.makedirs(param_value, exist_ok=True)
            return

        if command == "cat":
            with open(resource, "w", encoding="utf-8") as file:
                file.write(param_value)
            return

        if container_name == "":
            raise ValueError("Unknown container name is sent")

        current_resource = Path(resource)
        if not (current_resource.is_file() or current_resource.is_dir()):
            raise ValueError(f"File {resource} does not exist or mount missing.")

        if command == "chmod":
            os.chmod(path=resource, mode=int(param_value, base=8))
        elif command == "chown":
            try:
                uid, gid = param_value.split(":")
            except ValueError as exc:
                logger.error("User and group parameter shall be separated by ':' ")
                raise exc

            FsClient.add_users_to_node([uid, gid], in_place=True)
            shutil.chown(path=resource, user=uid, group=gid)
        elif command == "unlink":
            if not Path(param_value).is_dir():
                Path(param_value).unlink()
        else:
            raise ValueError(
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
    def get_beat_status_from_json(response: str, beat_name: str) -> str:
        """
        This function parses status response json retrieved as json and
        returns information from Application.Message field.
        @param response: Elastic-agent status string (param --output json)
        @param beat_name: The name of beat the status should be retrieved
        @return: status message string
        """
        response = json.loads(response)
        beat_list = response["components"]
        for beat in beat_list:
            if beat_name in beat["name"]:
                return beat["message"]
        return ""
