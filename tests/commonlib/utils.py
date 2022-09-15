import datetime
import time

from typing import Union

from commonlib.io_utils import get_logs_from_stream, get_events_from_index


def get_ES_evaluation(elastic_client, timeout, rule_tag, exec_timestamp,
                      resource_identifier=lambda r: True) -> Union[str, None]:
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
            events = get_events_from_index(elastic_client, elastic_client.index, rule_tag, latest_timestamp)
        except Exception as e:
            print(e)
            continue

        print('MATCHING EVENTS:', len(events))

        for event in events:
            findings_timestamp = datetime.datetime.strptime(getattr(event, '@timestamp'), '%Y-%m-%dT%H:%M:%S.%fZ')
            if findings_timestamp > latest_timestamp:
                latest_timestamp = findings_timestamp

            try:
                resource = event.resource.raw
                evaluation = event.result.evaluation
            except AttributeError:
                continue

            if resource_identifier(resource):
                print('FINDING MATCH:', event)
                return evaluation

    return None


def get_logs_evaluation(k8s, timeout, pod_name, namespace, rule_tag, exec_timestamp,
                        resource_identifier=lambda r: True) -> Union[str, None]:
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
            logs = get_logs_from_stream(k8s.get_pod_logs(pod_name=pod_name, namespace=namespace, since_seconds=2))
        except Exception as e:
            print(e)
            continue

        for log in logs:
            findings_timestamp = datetime.datetime.strptime(log.time, '%Y-%m-%dT%H:%M:%Sz')
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
    active_agents = 0
    num_cycles = 0

    while num_cycles < required_cycles and not is_timeout(start_time, 30):
        for node in nodes:
            start_time_per_agent = time.time()
            query, sort = elastic_client.build_es_query(term={"agent.name": node.metadata.name})
            while not is_timeout(start_time_per_agent, 10):
                # keep query ES until the sequence has changed
                result = elastic_client.get_index_data(index_name=elastic_client.index,
                                                       query=query,
                                                       sort=sort)
                doc_src = elastic_client.get_doc_source(data=result)
                curr_sequence = doc_src['event']['sequence']

                if elastic_client.get_total_value(data=result) != 0 and curr_sequence != prev_sequence:
                    # New cycle findings for this node
                    active_agents += 1
                    break
                time.sleep(1)

        if prev_sequence != curr_sequence:
            prev_sequence = curr_sequence
            num_cycles += 1

    return active_agents == (len(nodes) * required_cycles)


def is_timeout(start_time: time, timeout: int) -> bool:
    return time.time() - start_time > timeout
