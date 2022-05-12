"""
This module provides input / output manipulations on streams / files
"""

import os
import io
import json
import yaml
import shutil
from pathlib import Path
from munch import Munch
# import grp, pwd


def get_logs_from_stream(stream: str) -> list[Munch]:
    """
    This function converts logs stream to list of Munch objects (dictionaries)
    @param stream: StringIO stream
    @return: List of Munch objects
    """

    logs = io.StringIO(stream)
    result = []
    # with open("pod_logs.log", 'a') as pod_log_file:
    #     pod_log_file.writelines(logs)
    for log in logs:
        # current_log = log.split(sep="Z ")[1]
        if log and "bundles" in log:
            try:
                result.append(Munch(json.loads(log)))
            except json.decoder.JSONDecodeError:
                result.append(Munch(json.loads(log.replace("'", '"'))))
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
    result_list = []
    with file_path.open() as yaml_file:
        yaml_objects = yaml.safe_load_all(yaml_file)
        for yml_doc in yaml_objects:
            if yml_doc:
                doc = Munch(yml_doc)
                result_list.append({
                    doc.get('kind'): {key: value for key, value in doc.get('metadata').items()
                                      if key in ['name', 'namespace']}
                })
    return result_list


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
            raise Exception(f"Unknown {container_name} is sent")

        current_resource = Path(resource)
        if not current_resource.is_file():
            raise Exception(f"File {resource} does not exist or mount missing.")

        if command == 'chmod':
            os.chmod(path=resource, mode=int(param_value))
        elif command == 'chown':
            uid_gid = param_value.split(':')
            if len(uid_gid) != 2:
                raise Exception("User and group parameter shall be separated by ':' ")
            shutil.chown(path=resource, user=uid_gid[0], group=uid_gid[1])
        else:
            raise Exception(f"Command '{command}' still not implemented in test framework")
