"""
This module provides kubernetes functionality based on original kuberentes python library.
"""

from kubernetes import client, config


class KubernetesHelper:

    def __init__(self, is_in_cluster_config: bool = False):

        if is_in_cluster_config:
            config.load_incluster_config()
        else:
            config.load_kube_config()

        self.core_v1_client = client.CoreV1Api()

    def get_agent_pod_instances(self, agent_name: str, namespace: str):
        """
        This function retrieves all pod instances starts with agent_name, and located in defined namespace
        :param agent_name: pod instance name
        :param namespace: namespace
        :return: Pods list, otherwise []
        """
        pods = []
        if not namespace or not agent_name:
            return pods
        current_pod_list = self.core_v1_client.list_namespaced_pod(namespace=namespace)
        for pod in current_pod_list.items:
            if pod.metadata.name.startswith(agent_name) and "test" not in pod.metadata.name:
                pods.append(pod)
        return pods

    def get_cluster_nodes(self):
        nodes = self.core_v1_client.list_node()
        return nodes.items

