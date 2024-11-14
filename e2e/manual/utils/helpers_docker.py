import time

import docker
from docker.models.containers import Container


def wait_for_container(container_name_prefix: str) -> Container:
    print(f"Waiting for container {container_name_prefix}...")

    cli = docker.DockerClient()
    while True:
        containers = cli.containers.list()
        filtered = [c for c in containers if
                    c.name.startswith(container_name_prefix)]
        if len(filtered) > 0:
            return filtered[0]
        else:
            time.sleep(1)
