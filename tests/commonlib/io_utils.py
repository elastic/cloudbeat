"""
This module provides input / output manipulations on streams / files
"""

from datetime import datetime
import os
import io
import json
import shutil
from pathlib import Path
import yaml
from munch import Munch, munchify


def get_events_from_index(elastic_client, index_name: str, rule_tag: str, time_after: datetime) -> list[Munch]:
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
            "must": [
                {
                    "match": {
                        "rule.tags": rule_tag
                    }
                }
            ],
            "filter": [
                {
                    "range": {
                        "@timestamp": {
                            "gte": time_after.strftime('%Y-%m-%dT%H:%M:%S.%f')
                        }
                    }
                }
            ]
        }
    }
    sort = [{
        "@timestamp": {
            "order": "desc"
        }
    }]
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
            except Exception as ex:
                print(ex)
                continue

    return result


def get_k8s_yaml_objects(file_path: Path) -> list[str: dict]:
    """
    This function loads yaml file, and returns the following list:
    [ {<k8s_kind> : {<k8s_metadata}}]
    :param file_path: YAML path
    :return: [ {<k8s_kind> : {<k8s_metadata}}]
    """
    if not file_path:
        raise Exception(f'{file_path} is required')
    with file_path.open() as yaml_file:
        return [resource for resource in yaml.safe_load_all(yaml_file)]


class FsClient:
    """
    This class provides functionality for working with
    file system operations
    """

    @staticmethod
    def exec_command(container_name: str, command: str, param_value: str, resource: str):
        """
        This function executes os command
        @param container_name: Container node
        @param command: Linux command to be executed
        @param param_value: Value to be used in exec command
        @param resource: File / Resource path
        @return: None
        """

        if command == 'touch':
            if os.path.exists(param_value):
                return
            open(param_value, "a+").close()
            return

        if container_name == '':
            raise Exception("Unknown container name is sent")

        current_resource = Path(resource)
        if not (current_resource.is_file() or current_resource.is_dir()):
            raise Exception(
                f"File {resource} does not exist or mount missing.")

        if command == 'chmod':
            os.chmod(path=resource, mode=int(param_value, base=8))
        elif command == 'chown':
            uid_gid = param_value.split(':')
            if len(uid_gid) != 2:
                raise Exception(
                    "User and group parameter shall be separated by ':' ")
            shutil.chown(path=resource, user=uid_gid[0], group=uid_gid[1])
        elif command == 'unlink':
            if not Path(param_value).is_dir():
                Path(param_value).unlink()
        else:
            raise Exception(
                f"Command '{command}' still not implemented in test framework")

    @staticmethod
    def edit_process_file(container_name: str, dictionary, resource: str):
        """
        This function edits a process file
        @param container_name: Container node
        @param dictionary: Process parameters to set/unset
        @param resource: File / Resource path
        @return: None
        """
        if container_name == '':
            raise Exception("Unknown container name is sent")

        current_resource = Path(resource)
        if not current_resource.is_file():
            raise Exception(
                f"File {resource} does not exist or mount missing.")

        # Open and load the YAML into variable
        with current_resource.open() as f:
            r_file = yaml.safe_load(f)

        # Get process configuration arguments
        command = r_file["spec"]["containers"][0]["command"]

        # Collect set/unset keys and values from the dictionary
        set_dict = dictionary.get("set", {})
        unset_list = dictionary.get("unset", [])

        # Cycle across set items from the dictionary
        for skey, s_value in set_dict.items():
            # Find if set key exists already in the configuration arguments
            if any(skey == x.split("=")[0] for x in command):
                # Replace the value of the key with the new value from the set items
                command = list(map(lambda x: x.replace(
                    x, skey + "=" + s_value) if skey == x.split("=")[0] else x, command))
            else:
                # In case of non-existing key in the configuration arguments,
                # append the key/value from set items
                command.append(skey + "=" + s_value)

        # Cycle across unset items from the dictionary
        for us_key in unset_list:
            # Filter out the unset keys from the configuration arguments
            command = [x for x in command if us_key != x.split("=")[0]]

        # Override the configuration arguments with the newly built configuration arguments
        r_file["spec"]["containers"][0]["command"] = command

        # Write the newly built configuration arguments
        with current_resource.open(mode="w") as f:
            yaml.dump(r_file, f)

    @staticmethod
    def edit_config_file(container_name: str, dictionary, resource: str):
        """
        This function edits a config file
        @param container_name: Container node
        @param dictionary: Config parameters to set/unset
        @param resource: Config path
        @return: None
        """
        if container_name == '':
            raise Exception("Unknown container name is sent")

        current_resource = Path(resource)
        if not current_resource.is_file():
            raise Exception(
                f"File {resource} does not exist or mount missing.")

        # Open and load the YAML into variable
        with current_resource.open() as f:
            r_file = yaml.safe_load(f)

        # Collect set/unset keys and values from the dictionary
        set_dict = dictionary.get("set", {})
        unset_list = dictionary.get("unset", [])

        # Merge two dictionaries with priority for the set items
        r_file = {**r_file, **set_dict}

        # Cycle across unset items from the dictionary
        for us_key in unset_list:
            # Parsed dot separated key values
            keys = us_key.split('.')
            key_to_del = keys.pop()
            p = r_file

            # Advance inside the dictionary for nested keys
            for key in keys:
                p = p.get(key, None)
                if p is None:
                    # Non-existing nested key
                    break
            # Remove nested keys when all path exists
            if p:
                del p[key_to_del]
        # Write the newly built config
        with current_resource.open(mode="w") as f:
            yaml.dump(r_file, f)

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
        beat_list = response['Applications']
        for beat in beat_list:
            if beat['Name'] == beat_name:
                return beat['Message']
        return ''
