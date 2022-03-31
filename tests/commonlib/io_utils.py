"""
This module provides input / output manipulations on streams / files
"""

import io
import json
import yaml
from pathlib import Path
from munch import Munch


def get_logs_from_stream(stream: str) -> list[Munch]:
    """
    This function converts logs stream to list of Munch objects (dictionaries)
    @param stream: StringIO stream
    @param pattern: text to be search in log stream
    @return: List of Munch objects
    """
    logs = io.StringIO(stream)
    result = []
    for log in logs:
        if log and "bundles" in log:
            result.append(Munch(json.loads(log)))
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
    metadata_list = ['name', 'namespace']
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
