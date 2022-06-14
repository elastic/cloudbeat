"""
This module provides input / output manipulations on streams / files
"""

import os
import io
import json
import yaml
import shutil
from pathlib import Path
from munch import Munch, munchify


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
            except Exception as e:
                print(e)
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
            else:
                open(param_value, "a+")
                return

        # if command == 'getent' and param_value == 'group':
        #     try:
        #         grp.getgrnam(param_value)
        #         return ['etcd']
        #     except KeyError:
        #         return []
        #
        # if command == 'getent' and param_value == 'passwd':
        #     try:
        #         pwd.getpwnam(param_value)
        #         return ['etcd']
        #     except KeyError:
        #         return []
        #
        # if command == 'groupadd' and param_value == 'etcd':
        #     try:
        #         grp.getgrnam(param_value)
        #         return ['etcd']
        #     except KeyError:
        #         return []

        if container_name == '':
            raise Exception("Unknown container name is sent")

        current_resource = Path(resource)
        if not current_resource.is_file():
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
            raise Exception(f"Unknown container name is sent")

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
        for skey, svalue in set_dict.items():
            # Find if set key exists already in the configuration arguments 
            if any(skey == x.split("=")[0] for x in command):
                # Replace the value of the key with the new value from the set items
                command = list(map(lambda x: x.replace(
                    x, skey + "=" + svalue) if skey == x.split("=")[0] else x, command))
            else:
                # In case of non existing key in the configuration arguments, append the key/value from set items
                command.append(skey + "=" + svalue)

        # Cycle across unset items from the dictionary
        for uskey in unset_list:
            # Filter out the unset keys from the configuration arguments 
            command = [x for x in command if uskey != x.split("=")[0]]

        # Override the the configuration arguments with the newly built configuration arguments
        r_file["spec"]["containers"][0]["command"] = command

        # Write the newly build configuration arguments
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
        r_file = { **r_file, **set_dict }

        # Cycle across unset items from the dictionary
        for uskey in unset_list:
            # Parsed dot separated key values
            keys = uskey.split('.')
            key_to_del = keys.pop()
            p = r_file

            # Advance inside the dictionary for nested keys
            for key in keys:
                p = p.get(key, None)
                if p is None:
                    # Non existing nested key
                    break
            # Remove nested keys when all path exists
            if p:
                del p[key_to_del]
        
        # Write the newly build config
        with current_resource.open(mode="w") as f:
            yaml.dump(r_file, f)
