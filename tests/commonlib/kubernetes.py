"""
This module provides kubernetes functionality based on original kuberentes python library.
"""

from kubernetes import client, config, utils


class KubernetesHelper:

    def __init__(self, is_in_cluster_config: bool = False):

        if is_in_cluster_config:
            self.config = config.load_incluster_config()
        else:
            self.config = config.load_kube_config()

        self.core_v1_client = client.CoreV1Api()
        self.app_api = client.AppsV1Api()
        self.rbac_api = client.RbacAuthorizationV1Api()
        self.api_client = client.api_client.ApiClient(configuration=self.config)
        self.dispatch_delete = {
            'ConfigMap': self.core_v1_client.delete_namespaced_config_map,
            'ServiceAccount': self.core_v1_client.delete_namespaced_service_account,
            'DaemonSet': self.app_api.delete_namespaced_daemon_set,
            'Role': self.rbac_api.delete_namespaced_role,
            'RoleBinding': self.rbac_api.delete_namespaced_role_binding,
            'ClusterRoleBinding': self.rbac_api.delete_cluster_role_binding,
            'ClusterRole': self.rbac_api.delete_cluster_role
        }

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

    def get_pod_logs(self, pod_name: str, namespace: str, **kwargs):
        """
        This function returns pod logs
        @param pod_name: Name of pod
        @param namespace: Pod namespace
        @param kwargs:
        @return: Pod logs stream
        """

        return self.core_v1_client.read_namespaced_pod_log(name=pod_name, namespace=namespace, **kwargs)

    def start_agent(self, yaml_file: str, namespace: str):
        """
        This function deploys cloudbeat agent from yaml file
        :return:
        """

        return utils.create_from_yaml(k8s_client=self.api_client,
                                      yaml_file=yaml_file,
                                      namespace=namespace,
                                      verbose=True)

    def stop_agent(self, yaml_objects_list: list):
        """
        This function will delete all cloudbeat kubernetes resources.
        Currently, there is no ability to remove throug utils due to the following:
        https://github.com/kubernetes-client/python/pull/1392
        So below is cloud-security-posture own implementation.
        :return: V1Object - result
        """
        result_list = []
        for yaml_object in yaml_objects_list:
            for dict_key in yaml_object:
                result_list.append(self._delete_resources(resource_type=dict_key, **yaml_object[dict_key]))
        return result_list

    def _delete_resources(self, resource_type: str, **kwargs):
        """
        This is internal method for executing delete method depends on resource type.
        Binding is done using dispatch_delete dictionary.
        :param resource_type: Kubernetes resource to be deleted
        :param kwargs: Depends on resource type, it may be a name / name and namespace.
        :return:
        """
        return self.dispatch_delete[resource_type](**kwargs)

