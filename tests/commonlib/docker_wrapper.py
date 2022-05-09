"""
This module provides kubernetes functionality based on original docker SDK python library.
"""

import docker


class DockerWrapper:

    def __init__(self, config=None):
        if config.base_url != "":
            self.client = docker.DockerClient(base_url=config.base_url)
        else:
            self.client = docker.from_env()

    def exec_command(self, container_name: str, command: str, param_value: str, resource: str):
        """
        This function retrieves container by name / id and executes (docker exec) command to container
        @param container_name: Container id or name
        @param command: String command to be executed (for docker exec)
        @param param_value: Command function parameter value to be updated
        @param resource: Path to resource file
        @return: Command output, if exists
        """
        container = self.client.containers.get(container_id=container_name)
        command_f = f"{command} {param_value} {resource}"
        exit_code, output = container.exec_run(cmd=command_f)
        if exit_code > 0:
            return 'error'
        return output.decode().strip()
