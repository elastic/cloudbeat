# pylint: skip-file
import datetime
import json
import time
from typing import Union

import allure
from commonlib.io_utils import get_logs_from_stream, get_events_from_index
from loguru import logger


def get_ES_evaluation(
    elastic_client,
    timeout,
    rule_tag,
    exec_timestamp,
    resource_identifier=lambda r: True,
) -> Union[str, None]:
    """
    This function retrieves ES events and verifies if evaluation resource matches the resource being tested.
    It returns None if no events for the given rule_tag and the given resource_identifier can be found before timeout.
    @param elastic_client: a client to interact with ES
    @param resource_identifier: function to match the evaluation resource
    @param timeout: time before the function will stop trying to look for matching events
    @param rule_tag: event rule_tag to match
    @param exec_timestamp: the timestamp after which to look for events
    """
    start_time = time.time()
    latest_timestamp = exec_timestamp

    while time.time() - start_time < timeout:
        try:
            # timeout used for reducing requests frequency to ElasticSearch
            time.sleep(2)
            events = get_events_from_index(
                elastic_client,
                elastic_client.index,
                rule_tag,
                latest_timestamp,
            )
        except Exception as e:
            logger.debug(e)
            continue

        for event in events:
            findings_timestamp = datetime.datetime.strptime(
                getattr(event, "@timestamp"),
                "%Y-%m-%dT%H:%M:%S.%fZ",
            )
            if findings_timestamp > latest_timestamp:
                latest_timestamp = findings_timestamp

            try:
                evaluation = event.result.evaluation
            except AttributeError:
                logger.warning("got finding with missing fields:", event)
                continue

            if resource_identifier(event):
                allure.attach(
                    json.dumps(
                        event,
                        indent=4,
                        sort_keys=True,
                    ),
                    rule_tag,
                    attachment_type=allure.attachment_type.JSON,
                )
                return evaluation

    return None


def get_logs_evaluation(
    k8s,
    timeout,
    pod_name,
    namespace,
    rule_tag,
    exec_timestamp,
    resource_identifier=lambda r: True,
) -> Union[str, None]:
    """
    This is a legacy function for debugging purposes.
    This function retrieves pod logs and verifies if evaluation result is equal to expected result.
    It returns None if no pod logs for evaluation for the given rule_tag can be found.
    @param resource_identifier: function to filter a specific resource
    @param k8s: Kubernetes wrapper instance
    @param timeout: Exit timeout
    @param pod_name: Name of pod the logs shall be retrieved from
    @param namespace: Kubernetes namespace
    @param rule_tag: Log rule tag
    @param exec_timestamp: the timestamp the command executed
    """
    start_time = time.time()
    while time.time() - start_time < timeout:
        try:
            logs = get_logs_from_stream(
                k8s.get_pod_logs(
                    pod_name=pod_name,
                    namespace=namespace,
                    since_seconds=2,
                ),
            )
        except Exception as e:
            logger.warning(e)
            continue

        for log in logs:
            findings_timestamp = datetime.datetime.strptime(
                log.time,
                "%Y-%m-%dT%H:%M:%Sz",
            )
            if (findings_timestamp - exec_timestamp).total_seconds() < 0:
                continue

            try:
                findings = log.result.findings
                resource = log.result.resource
            except AttributeError:
                continue

            for finding in findings:
                if rule_tag in finding.rule.tags:
                    if resource_identifier(resource):
                        return finding.result.evaluation
    return None


def dict_contains(small, big):
    """
    Checks if the small dict like object is contained inside the big object
    @param small: dict like object
    @param big: dict like object
    @return: true iff the small dict like object is contained inside the big object
    """
    if isinstance(small, dict):
        if not set(small.keys()) <= set(big.keys()):
            return False
        for key in small.keys():
            if not dict_contains(small.get(key), big.get(key)):
                return False
        return True

    return small == big


def get_resource_identifier(body):
    def resource_identifier(resource):
        if getattr(resource, "to_dict", None):
            return dict_contains(body, resource.to_dict())
        if getattr(resource, "__dict__", None):
            return dict_contains(body, dict(resource))

    return resource_identifier


def wait_for_cycle_completion(elastic_client, nodes: list) -> bool:
    """
    Wait for all agents to finish sending findings to ES.
    Done by waiting for all agents to send at least a single finding in the second cycle,
    by that we verify that the first cycle is completed.
    @param elastic_client: ES client
    @param nodes: nodes list
    @return: true if all agents finished sending findings in the configured timeout
    """
    required_cycles = 2
    start_time = time.time()
    prev_sequence = ""
    curr_sequence = ""
    agents_cycles_count = 0
    num_cycles = 0
    cycle_timeout = 30  # change this timeout if cloudbeat fetching period changed
    all_cycles_timeout = cycle_timeout * required_cycles * len(nodes)
    node_cycle_timeout = cycle_timeout * required_cycles

    # This 'while' is used for all cycles, all nodes
    while num_cycles < required_cycles and not is_timeout(start_time, all_cycles_timeout):
        for node in nodes:
            start_time_per_agent = time.time()
            query, sort = elastic_client.build_es_query(
                term={"agent.name": node.metadata.name},
            )
            # this 'while' used for single node all cycles timeout
            while not is_timeout(start_time_per_agent, node_cycle_timeout):
                # keep query ES until the sequence has changed
                result = elastic_client.get_index_data(
                    index_name=elastic_client.index,
                    query=query,
                    sort=sort,
                )
                doc_src = elastic_client.get_doc_source(data=result)
                if len(doc_src) == 0:
                    continue
                curr_sequence = doc_src["event"]["sequence"]

                if elastic_client.get_total_value(data=result) != 0 and curr_sequence != prev_sequence:
                    # New cycle findings for this node
                    agents_cycles_count += 1
                    break
                time.sleep(1)

        if prev_sequence != curr_sequence:
            prev_sequence = curr_sequence
            num_cycles += 1

    return agents_cycles_count >= (len(nodes) * required_cycles)


def is_timeout(start_time: time, timeout: int) -> bool:
    return time.time() - start_time > timeout


def command_contains_arguments(command, arguments_dict):
    args = command.split()[1:]
    args_dict = {}
    for arg in args:
        key, val = arg.split("=", 1)
        args_dict[key] = val

    set_dict = arguments_dict.get("set", {})
    unset_list = arguments_dict.get("unset", [])

    for key, val in set_dict.items():
        arg_val = args_dict.get(key)
        if val != arg_val:
            return False

    for key in unset_list:
        if key in args_dict:
            return False

    return True


def config_contains_arguments(config, arguments_dict):
    set_dict = arguments_dict.get("set", {})
    unset_list = arguments_dict.get("unset", [])

    if not dict_contains(set_dict, config):
        return False

    for arg in unset_list:
        current = config
        arg_set = True

        for arg_part in arg.split("."):
            if (not isinstance(current, dict)) or (arg_part not in current):
                arg_set = False
                break

            current = current[arg_part]

        if arg_set:
            return False

    return True


def get_findings(elastic_client, config_timeout, match_type):
    query, sort = elastic_client.build_es_query(term={"resource.type": match_type})
    start_time = time.time()
    result = {}
    while time.time() - start_time < config_timeout:
        current_result = elastic_client.get_index_data(
            index_name=elastic_client.index,
            query=query,
            sort=sort,
        )
        if elastic_client.get_total_value(data=current_result) != 0:
            allure.attach(
                json.dumps(
                    elastic_client.get_doc_source(data=current_result),
                    indent=4,
                    sort_keys=True,
                ),
                match_type,
                attachment_type=allure.attachment_type.JSON,
            )
            result = current_result
            break
        time.sleep(1)

    return result


def identifier_by_name(case_identifier, eval_resource) -> bool:
    """
    This function compares unique field value retrieved from elastic to the value defined in the test case data.
    Each test case has own unique field. This function uses resource.name as identifier.
    @param case_identifier: Test case data identifier to be used for comparison
    @param eval_resource: Resource event data retrieved from elastic
    @return: True / False
    """
    try:
        return eval_resource.resource.name == case_identifier
    except AttributeError:
        return False
