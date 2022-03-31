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

    def exec_command(self, container_name: str, command: str):
        """
        This function retrieves container by name / id and executes (docker exec) command to container
        @param container_name: Container id or name
        @param command: String command to be executed (for docker exec)
        @return: Command output, if exists
        """
        container = self.client.containers.get(container_id=container_name)
        exit_code, output = container.exec_run(cmd=command)
        if exit_code > 0:
            return ''
        return output.decode().strip()
