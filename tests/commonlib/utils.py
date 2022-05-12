from commonlib.io_utils import get_logs_from_stream
import time


def get_evaluation(k8s, timeout, pod_name, namespace, rule_tag, resource_identifier=lambda r: True) -> str:
    """
    This function retrieves pod logs and verifies if evaluation result is equal to expected result.
    @param resource_identifier: function to filter a specific resource
    @param k8s: Kubernetes wrapper instance
    @param timeout: Exit timeout
    @param pod_name: Name of pod the logs shall be retrieved from
    @param namespace: Kubernetes namespace
    @param rule_tag: Log rule tag
    """
    start_time = time.time()
    while time.time() - start_time < timeout:
        try:
            logs = get_logs_from_stream(k8s.get_pod_logs(pod_name=pod_name, namespace=namespace, since_seconds=1))
        except Exception as e:
            print(e)
            continue

        for log in logs:
            for finding in log.result.findings:
                # print(f"Resource Kind = {log.result.resource.get('kind', log.result.resource.get('subtype'))}")
                if rule_tag in finding.rule.tags:
                    resource = log.result.resource
                    if resource_identifier(resource):
                        return finding.result.evaluation
    return "Unknown"


def compare_dicts(small, big):
    if isinstance(small, dict):
        if not set(small.keys()) <= set(big.keys()):
            return False
        for key in small.keys():
            if not compare_dicts(small.get(key), big.get(key)):
                return False
        return True

    return small == big


def get_resource_identifier(body):
    def resource_identifier(resource):
        if getattr(resource, "to_dict", None):
            return compare_dicts(body, resource.to_dict())
        if getattr(resource, "__dict__", None):
            return compare_dicts(body, dict(resource))

    return resource_identifier
