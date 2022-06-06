"""
This module provides kubernetes functionality based on original kubernetes python library.
"""

from kubernetes import client, config, utils
from kubernetes.client import ApiException
from kubernetes.watch import watch


class KubernetesHelper:

    def __init__(self, is_in_cluster_config: bool = False):
        if is_in_cluster_config:
            self.config = config.load_incluster_config()
        else:
            self.config = config.load_kube_config()

        self.policy_c1_api = client.PolicyV1beta1Api()
        self.core_v1_client = client.CoreV1Api()
        self.app_api = client.AppsV1Api()
        self.rbac_api = client.RbacAuthorizationV1Api()
        self.coordination_v1_api = client.CoordinationV1Api()
        self.api_client = client.api_client.ApiClient(configuration=self.config)

        self.dispatch_list = {
            'Pod': self.core_v1_client.list_namespaced_pod,
            'ConfigMap': self.core_v1_client.list_namespaced_config_map,
            'ServiceAccount': self.core_v1_client.list_namespaced_service_account,
            'DaemonSet': self.app_api.list_namespaced_daemon_set,
            'Role': self.rbac_api.list_namespaced_role,
            'RoleBinding': self.rbac_api.list_namespaced_role_binding,
            'ClusterRoleBinding': self.rbac_api.list_cluster_role_binding,
            'PodSecurityPolicy': self.policy_c1_api.list_pod_security_policy
            'ClusterRole': self.rbac_api.list_cluster_role,
            'Lease': self.coordination_v1_api.list_namespaced_lease,
        }

        self.dispatch_delete = {
            'Pod': self.core_v1_client.delete_namespaced_pod,
            'ConfigMap': self.core_v1_client.delete_namespaced_config_map,
            'ServiceAccount': self.core_v1_client.delete_namespaced_service_account,
            'DaemonSet': self.app_api.delete_namespaced_daemon_set,
            'Role': self.rbac_api.delete_namespaced_role,
            'RoleBinding': self.rbac_api.delete_namespaced_role_binding,
            'ClusterRoleBinding': self.rbac_api.delete_cluster_role_binding,
            'PodSecurityPolicy': self.policy_c1_api.delete_pod_security_policy
            'ClusterRole': self.rbac_api.delete_cluster_role,
            'Lease': self.coordination_v1_api.delete_namespaced_lease
        }

        self.dispatch_patch = {
            'Pod': self.core_v1_client.patch_namespaced_pod,
            'ConfigMap': self.core_v1_client.patch_namespaced_config_map,
            'ServiceAccount': self.core_v1_client.patch_namespaced_service_account,
            'DaemonSet': self.app_api.patch_namespaced_daemon_set,
            'Role': self.rbac_api.patch_namespaced_role,
            'RoleBinding': self.rbac_api.patch_namespaced_role_binding,
            'ClusterRoleBinding': self.rbac_api.patch_cluster_role_binding,
            'PodSecurityPolicy': self.policy_c1_api.patch_pod_security_policy
            'ClusterRole': self.rbac_api.patch_cluster_role,
            'Lease': self.coordination_v1_api.patch_namespaced_lease
        }

        self.dispatch_create = {
            'Pod': self.core_v1_client.create_namespaced_pod,
            'ConfigMap': self.core_v1_client.create_namespaced_config_map,
            'ServiceAccount': self.core_v1_client.create_namespaced_service_account,
            'DaemonSet': self.app_api.create_namespaced_daemon_set,
            'Role': self.rbac_api.create_namespaced_role,
            'RoleBinding': self.rbac_api.create_namespaced_role_binding,
            'ClusterRoleBinding': self.rbac_api.create_cluster_role_binding,
            'PodSecurityPolicy': self.policy_c1_api.create_pod_security_policy
            'ClusterRole': self.rbac_api.create_cluster_role,
            'Lease': self.coordination_v1_api.create_namespaced_lease
        }

        self.dispatch_get = {
            'Pod': self.core_v1_client.read_namespaced_pod,
            'ConfigMap': self.core_v1_client.read_namespaced_config_map,
            'ServiceAccount': self.core_v1_client.read_namespaced_service_account,
            'DaemonSet': self.app_api.read_namespaced_daemon_set,
            'Role': self.rbac_api.read_namespaced_role,
            'RoleBinding': self.rbac_api.read_namespaced_role_binding,
            'ClusterRoleBinding': self.rbac_api.read_cluster_role_binding,
            'PodSecurityPolicy': self.policy_c1_api.read_pod_security_policy
            'ClusterRole': self.rbac_api.read_cluster_role,
            'Lease': self.coordination_v1_api.read_namespaced_lease
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

    def get_service_accounts(self, namespace: str):
        service_accounts = self.core_v1_client.list_namespaced_service_account(namespace=namespace)
        return service_accounts

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
        return self.create_from_yaml(yaml_file=yaml_file, namespace=namespace)

    def create_from_yaml(self, yaml_file: str, namespace: str, verbose: bool = True):
        return utils.create_from_yaml(
            k8s_client=self.api_client,
            yaml_file=yaml_file,
            namespace=namespace,
            verbose=verbose
        )

    def create_from_dict(self, data: dict, namespace: str, verbose: bool = True, **_):
        return utils.create_from_dict(
            k8s_client=self.api_client,
            data=data,
            namespace=namespace,
            verbose=verbose
        )

    def delete_from_yaml(self, yaml_objects_list: list):
        """
        This function will delete all cloudbeat kubernetes resources.
        Currently, there is no ability to remove through utils due to the following:
        https://github.com/kubernetes-client/python/pull/1392
        So below is cloud-security-posture own implementation.
        :return: V1Object - result
        """
        result_list = []
        for yaml_object in yaml_objects_list:
            metadata = yaml_object['metadata']
            relevant_metadata = {k: metadata[k] for k in ('name', 'namespace') if k in metadata}
            try:
                self.get_resource(
                    resource_type=yaml_object['kind'],
                    **relevant_metadata
                )
                result_list.append(self.delete_resources(
                    resource_type=yaml_object['kind'],
                    **relevant_metadata
                ))
            except ApiException as notFound:
                print(f"{relevant_metadata['name']} not found {notFound.status}")

        return result_list

    def delete_resources(self, resource_type: str, **kwargs):
        """
        """
        return self.dispatch_delete[resource_type](**kwargs)

    def patch_resources(self, resource_type: str, **kwargs):
        """
        """
        return self.dispatch_patch[resource_type](**kwargs)

    def list_resources(self, resource_type: str, **kwargs):
        """
        """
        return self.dispatch_list[resource_type](**kwargs)

    def create_resources(self, resource_type: str, **kwargs):
        """
        """
        return self.dispatch_create[resource_type](**kwargs)

    def get_resource(self, resource_type: str, name: str, **kwargs):
        """
        """
        try:
            return self.dispatch_get[resource_type](name, **kwargs)
        except ApiException as e:
            print(f"Resource not found: {e.reason}")
            raise e

    def wait_for_resource(self, resource_type: str, name: str, status_list: list,
                          timeout: int = 120, **kwargs) -> bool:
        """
        watches a resources for a status change
        @param resource_type: the resource type
        @param name: resource name
        @param status_list: excepted statuses e.g., RUNNING, DELETED, MODIFIED, ADDED
        @param timeout: until wait
        @return: True if status reached
        """
        w = watch.Watch()
        for event in w.stream(func=self.dispatch_list[resource_type],
                              timeout_seconds=timeout,
                              **kwargs):
            if event["object"].metadata.name == name and event["type"] in status_list:
                w.stop()
                return True
        return False
