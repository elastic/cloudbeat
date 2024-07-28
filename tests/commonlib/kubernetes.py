"""
This module provides kubernetes functionality based on original kubernetes python library.
"""

from pathlib import Path
from subprocess import CalledProcessError
from typing import Union

from commonlib.io_utils import get_k8s_yaml_objects
from kubernetes import client, config, utils
from kubernetes.client import ApiException
from kubernetes.stream import stream
from kubernetes.watch import watch
from loguru import logger

RESOURCE_POD = "Pod"
RESOURCE_SERVICE_ACCOUNT = "ServiceAccount"
LEASE_NAME = "cloudbeat-cluster-leader"


class KubernetesHelper:
    """
    This class is Kubernetes wrapper
    """

    def __init__(self, is_in_cluster_config: bool = False):
        if is_in_cluster_config:
            self.config = config.load_incluster_config()
        else:
            self.config = config.load_kube_config()

        # self.policy_c1_api = client.PolicyV1beta1Api()
        self.core_v1_client = client.CoreV1Api()
        self.app_api = client.AppsV1Api()
        self.rbac_api = client.RbacAuthorizationV1Api()
        self.coordination_v1_api = client.CoordinationV1Api()
        self.api_client = client.api_client.ApiClient(configuration=self.config)

        self.dispatch_list = {
            RESOURCE_POD: self.core_v1_client.list_namespaced_pod,
            "ConfigMap": self.core_v1_client.list_namespaced_config_map,
            RESOURCE_SERVICE_ACCOUNT: self.core_v1_client.list_namespaced_service_account,
            "DaemonSet": self.app_api.list_namespaced_daemon_set,
            "Role": self.rbac_api.list_namespaced_role,
            "RoleBinding": self.rbac_api.list_namespaced_role_binding,
            "ClusterRoleBinding": self.rbac_api.list_cluster_role_binding,
            "ClusterRole": self.rbac_api.list_cluster_role,
            # "PodSecurityPolicy": self.policy_c1_api.list_pod_security_policy,
            "Lease": self.coordination_v1_api.list_namespaced_lease,
        }

        self.dispatch_delete = {
            RESOURCE_POD: self.core_v1_client.delete_namespaced_pod,
            "ConfigMap": self.core_v1_client.delete_namespaced_config_map,
            RESOURCE_SERVICE_ACCOUNT: self.core_v1_client.delete_namespaced_service_account,
            "DaemonSet": self.app_api.delete_namespaced_daemon_set,
            "Role": self.rbac_api.delete_namespaced_role,
            "RoleBinding": self.rbac_api.delete_namespaced_role_binding,
            "ClusterRoleBinding": self.rbac_api.delete_cluster_role_binding,
            # "PodSecurityPolicy": self.policy_c1_api.delete_pod_security_policy,
            "ClusterRole": self.rbac_api.delete_cluster_role,
            "Lease": self.coordination_v1_api.delete_namespaced_lease,
        }

        self.dispatch_patch = {
            "Pod": self.core_v1_client.patch_namespaced_pod,
            "ConfigMap": self.core_v1_client.patch_namespaced_config_map,
            "ServiceAccount": self.core_v1_client.patch_namespaced_service_account,
            "DaemonSet": self.app_api.patch_namespaced_daemon_set,
            "Role": self.rbac_api.patch_namespaced_role,
            "RoleBinding": self.rbac_api.patch_namespaced_role_binding,
            "ClusterRoleBinding": self.rbac_api.patch_cluster_role_binding,
            # "PodSecurityPolicy": self.policy_c1_api.patch_pod_security_policy,
            "ClusterRole": self.rbac_api.patch_cluster_role,
            "Lease": self.coordination_v1_api.patch_namespaced_lease,
        }

        self.dispatch_create = {
            "Pod": self.core_v1_client.create_namespaced_pod,
            "ConfigMap": self.core_v1_client.create_namespaced_config_map,
            "ServiceAccount": self.core_v1_client.create_namespaced_service_account,
            "DaemonSet": self.app_api.create_namespaced_daemon_set,
            "Role": self.rbac_api.create_namespaced_role,
            "RoleBinding": self.rbac_api.create_namespaced_role_binding,
            "ClusterRoleBinding": self.rbac_api.create_cluster_role_binding,
            # "PodSecurityPolicy": self.policy_c1_api.create_pod_security_policy,
            "ClusterRole": self.rbac_api.create_cluster_role,
            "Lease": self.coordination_v1_api.create_namespaced_lease,
        }

        self.dispatch_get = {
            "Pod": self.core_v1_client.read_namespaced_pod,
            "ConfigMap": self.core_v1_client.read_namespaced_config_map,
            "ServiceAccount": self.core_v1_client.read_namespaced_service_account,
            "DaemonSet": self.app_api.read_namespaced_daemon_set,
            "Role": self.rbac_api.read_namespaced_role,
            "RoleBinding": self.rbac_api.read_namespaced_role_binding,
            "ClusterRoleBinding": self.rbac_api.read_cluster_role_binding,
            # "PodSecurityPolicy": self.policy_c1_api.read_pod_security_policy,
            "ClusterRole": self.rbac_api.read_cluster_role,
            "Lease": self.coordination_v1_api.read_namespaced_lease,
        }

    def get_agent_pod_instances(self, agent_name: str, namespace: str):
        """
        This function retrieves all Pod instances that start with agent_name in the given namespace.
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

    def get_pod_image_version(self, pod_name: str, namespace: str) -> dict:
        """
        This function retrieves image version for specified pod
        @param pod_name: Pod instance name
        @param namespace: Namespace location
        @return:
        """

        result_dict = {}
        try:
            current_pod_list = self.core_v1_client.list_namespaced_pod(
                namespace=namespace,
            )
            for pod in current_pod_list.items:
                if pod_name in pod.metadata.name:
                    result_dict[pod.metadata.name] = pod.spec.containers[0].image
            return result_dict
        except ValueError:
            logger.warning("Cannot retrieve pod image version")

        return result_dict

    def get_nodes_versions(self) -> dict:
        """
        This function retrieves cluster nod versions
        @return:
        """
        nodes_data = {}
        try:
            nodes = self.get_cluster_nodes()
            for node in nodes:
                nodes_data[node.metadata.name] = node.status.node_info.kubelet_version
        except ValueError:
            logger.warning("Cannot retrieve nodes data")

        return nodes_data

    def get_service_accounts(self, namespace: str):
        """
        This function retrieves all ServiceAccount instances in the given namespace.
        """
        service_accounts = self.core_v1_client.list_namespaced_service_account(
            namespace=namespace,
        )
        return service_accounts

    def get_cluster_nodes(self):
        """
        This function retrieves all Nodes.
        """
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

        return self.core_v1_client.read_namespaced_pod_log(
            name=pod_name,
            namespace=namespace,
            **kwargs,
        )

    def start_agent(self, yaml_file: str, namespace: str):
        """
        This function deploys cloudbeat agent from yaml file
        :return:
        """
        return self.create_from_yaml(yaml_file=yaml_file, namespace=namespace)

    def create_from_yaml(self, yaml_file: str, namespace: str, verbose: bool = True):
        """
        Create the K8S resources described in the given YALM manifest file.
        """
        return utils.create_from_yaml(
            k8s_client=self.api_client,
            yaml_file=yaml_file,
            namespace=namespace,
            verbose=verbose,
        )

    def create_from_dict(self, data: dict, namespace: str, verbose: bool = True, **_):
        """
        Create the K8S resources described in the given dict.
        """
        return utils.create_from_dict(
            k8s_client=self.api_client,
            data=data,
            namespace=namespace,
            verbose=verbose,
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
            if yaml_object is None:
                continue
            metadata = yaml_object["metadata"]
            relevant_metadata = {k: metadata[k] for k in ("name", "namespace") if k in metadata}
            try:
                self.get_resource(resource_type=yaml_object["kind"], **relevant_metadata)
                result_list.append(
                    self.delete_resources(resource_type=yaml_object["kind"], **relevant_metadata),
                )
            except ApiException as not_found:
                logger.warning(f"{relevant_metadata['name']} not found {not_found.status}")

        return result_list

    def delete_resources(self, resource_type: str, **kwargs):
        """
        Delete the given K8S resources.
        """
        return self.dispatch_delete[resource_type](**kwargs)

    def patch_resources(self, resource_type: str, **kwargs):
        """
        Update the given K8S resources using the given patch.

        Only a limited set of resources/patches can be applied directly.
        For the rest, the current resource is deleted and a new one created
        with the given patch.
        """
        if resource_type != RESOURCE_POD:
            return self.dispatch_patch[resource_type](**kwargs)

        patch_body = kwargs.pop("body")

        self.get_resource(resource_type, **kwargs)
        self.delete_resources(resource_type=resource_type, **kwargs)
        deleted = self.wait_for_resource(
            resource_type=resource_type,
            status_list=["DELETED"],
            **kwargs,
        )

        if not deleted:
            raise ValueError(f"could not delete {resource_type}: {kwargs}")

        return self.create_patched_resource(resource_type, patch_body)

    def create_patched_resource(self, patch_resource_type, patch_body):
        """
        Delete and recreate the given resource with the given patch.
        """
        file_path = Path(__file__).parent / "../test_environments/mock-pod.yml"
        k8s_resources = get_k8s_yaml_objects(file_path=file_path)

        patch_metadata = patch_body["metadata"]
        patch_relevant_metadata = {k: patch_metadata[k] for k in ("name", "namespace") if k in patch_metadata}

        patched_resource = None

        for yml_resource in k8s_resources:
            resource_type, metadata = yml_resource["kind"], yml_resource["metadata"]
            relevant_metadata = {k: metadata[k] for k in ("name", "namespace") if k in metadata}

            if resource_type != patch_resource_type or relevant_metadata != patch_relevant_metadata:
                continue

            patched_body = self.patch_resource_body(yml_resource, patch_body)
            created_resource = self.create_from_dict(patched_body, **relevant_metadata)

            done = self.wait_for_resource(
                resource_type=resource_type,
                status_list=["RUNNING", "ADDED"],
                **relevant_metadata,
            )
            if done:
                patched_resource = created_resource

            break

        return patched_resource

    def patch_resource_body(
        self,
        body: Union[list, dict],
        patch: Union[list, dict],
    ) -> Union[list, dict]:
        """
        Update the given resource body with the given patch.
        """
        if not isinstance(body, type(patch)):
            raise ValueError(
                f"Cannot compare {type(body)}: {body} with {type(patch)}: {patch}",
            )

        if isinstance(body, dict):
            for key, val in patch.items():
                if key not in body:
                    body[key] = val
                else:
                    if isinstance(val, (dict, list)):
                        body[key] = self.patch_resource_body(body[key], val)
                    else:
                        body[key] = val

        elif isinstance(body, list):
            for i, val in enumerate(patch):
                if i >= len(body):
                    break

                if isinstance(val, (dict, list)):
                    body[i] = self.patch_resource_body(body[i], val)
                else:
                    body[i] = val

            if len(patch) > len(body):
                body += patch[len(body) :]

        else:
            raise ValueError(f"Invalid body {body} of type {type(body)}")

        return body

    def list_resources(self, resource_type: str, **kwargs):
        """
        List resources of the given resource_type.
        """
        return self.dispatch_list[resource_type](**kwargs)

    def create_resources(self, resource_type: str, **kwargs):
        """
        Create resources of the given resource_type.
        """
        return self.dispatch_create[resource_type](**kwargs)

    def get_resource(self, resource_type: str, name: str, **kwargs):
        """
        Fetch details of the given resource.
        """
        try:
            return self.dispatch_get[resource_type](name, **kwargs)
        except ApiException as exc:
            logger.warning(f"Resource not found: {exc.reason}")
            raise exc

    def wait_for_resource(
        self,
        resource_type: str,
        name: str,
        status_list: list,
        timeout: int = 120,
        **kwargs,
    ) -> bool:
        """
        watches a resources for a status change
        @param resource_type: the resource type
        @param name: resource name
        @param status_list: accepted statuses e.g., RUNNING, DELETED, MODIFIED, ADDED
        @param timeout: until wait
        @return: True if status reached
        """
        # When pods are being created, MODIFIED events are also of interest to check if
        # they successfully transition from ContainerCreating to Running state.
        if (resource_type == RESOURCE_POD) and ("ADDED" in status_list) and ("MODIFIED" not in status_list):
            status_list.append("MODIFIED")

        kube_watch = watch.Watch()
        for event in kube_watch.stream(func=self.dispatch_list[resource_type], timeout_seconds=timeout, **kwargs):
            if name in event["object"].metadata.name and event["type"] in status_list:
                if (
                    (resource_type == RESOURCE_POD)
                    and ("ADDED" in status_list)
                    and (event["object"].status.phase == "Pending")
                ):
                    continue
                kube_watch.stop()
                return True

        return False

    def pod_exec(self, name: str, namespace: str, command: list) -> str:
        """
        This function connects to pod and executes command
        @param name: Pod name
        @param namespace: Pod namespace
        @param command: Command to be executed
        @return: Executed command response
        """

        resp = stream(
            self.core_v1_client.connect_get_namespaced_pod_exec,
            name,
            namespace,
            command=command,
            stderr=True,
            stdin=False,
            stdout=True,
            tty=False,
            _preload_content=False,
        )
        response = ""
        while resp.is_open():
            resp.update(timeout=10)
            if resp.peek_stdout():
                response = resp.read_stdout()
            if resp.peek_stderr():
                response = resp.read_stderr()

        resp.close()
        if resp.returncode != 0:
            raise CalledProcessError(returncode=resp.returncode, cmd=command, output=response)
        return response

    def get_cluster_leader(self, namespace: str, pods: list) -> str:
        """
        retrieve the node name of the leading cloudbeat
        @param namespace: namespace
        @param pods: a list of the cluster pods
        @return: Leader's node name
        """

        lease_info = self.get_resource(
            resource_type="Lease",
            name=LEASE_NAME,
            namespace=namespace,
        )
        lease_holder_identity = lease_info.spec.holder_identity
        holder_id = lease_holder_identity.split("_")[-1]

        for pod in pods:
            if holder_id in pod.metadata.name:
                return pod.spec.node_name

        return ""
