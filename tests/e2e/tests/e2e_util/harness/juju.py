#
# Copyright 2023 Canonical, Ltd.
#
import logging
import shlex
import subprocess
from pathlib import Path

from e2e_util import config
from e2e_util.harness import Harness, HarnessError
from e2e_util.util import run

LOG = logging.getLogger(__name__)


class JujuHarness(Harness):
    """A Harness that creates an Juju machine for each instance."""

    def __init__(self):
        super(JujuHarness, self).__init__()

        self.model = config.JUJU_MODEL
        if not self.model:
            raise HarnessError("Set JUJU_MODEL to the Juju model to use")

        if config.JUJU_CONTROLLER:
            self.model = f"{config.JUJU_CONTROLLER}:{self.model}"

        self.constraints = config.JUJU_CONSTRAINTS
        self.base = config.JUJU_BASE
        self.existing_machines = {}
        self.instances = set()

        if config.JUJU_MACHINES:
            self.existing_machines = {
                instance_id.strip(): False
                for instance_id in config.JUJU_MACHINES.split()
            }
            LOG.debug(
                "Configured Juju substrate (model %s, machines %s)",
                self.model,
                config.JUJU_MACHINES,
            )

        else:
            LOG.debug(
                "Configured Juju substrate (model %s, base %s, constraints %s)",
                self.model,
                self.base,
                self.constraints,
            )

    def new_instance(self) -> str:
        for instance_id in self.existing_machines:
            if not self.existing_machines[instance_id]:
                LOG.debug("Reusing existing machine %s", instance_id)
                self.existing_machines[instance_id] = True
                self.instances.add(instance_id)
                return instance_id

        LOG.debug("Creating instance with constraints %s", self.constraints)
        try:
            p = run(
                [
                    "juju",
                    "add-machine",
                    "-m",
                    self.model,
                    "--constraints",
                    self.constraints,
                    "--base",
                    self.base,
                ],
                capture_output=True,
            )

            output = p.stderr.decode().strip()
            if not output.startswith("created machine "):
                raise HarnessError(f"failed to parse output from juju add-machine {p=}")

            instance_id = output.split(" ")[2]
        except subprocess.CalledProcessError as e:
            raise HarnessError("Failed to create Juju machine") from e

        self.instances.add(instance_id)

        self.exec(instance_id, ["snap", "wait", "system", "seed.loaded"])
        return instance_id

    def send_file(self, instance_id: str, source: str, destination: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        if not Path(destination).is_absolute():
            raise HarnessError(f"path {destination} must be absolute")

        LOG.debug(
            "Copying file %s to instance %s at %s", source, instance_id, destination
        )
        try:
            self.exec(
                instance_id,
                ["mkdir", "-m=0777", "-p", Path(destination).parent.as_posix()],
            )
            run(["juju", "scp", source, f"{instance_id}:{destination}"])
        except subprocess.CalledProcessError as e:
            raise HarnessError("juju scp command failed") from e

    def pull_file(self, instance_id: str, source: str, destination: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        if not Path(source).is_absolute():
            raise HarnessError(f"path {source} must be absolute")

        LOG.debug(
            "Copying file %s from instance %s to %s", source, instance_id, destination
        )
        try:
            run(["juju", "scp", f"{instance_id}:{source}", destination])
        except subprocess.CalledProcessError as e:
            raise HarnessError("juju scp command failed") from e

    def exec(self, instance_id: str, command: list, **kwargs):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        LOG.debug("Execute command %s in instance %s", command, instance_id)
        return run(
            [
                "juju",
                "exec",
                "-m",
                self.model,
                "--machine",
                instance_id,
                "--",
                "sudo",
                "-E",
                "bash",
                "-c",
                shlex.join(command),
            ],
            **kwargs,
        )

    def delete_instance(self, instance_id: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        if self.existing_machines.get(instance_id):
            # For existing machines, simply mark it as free
            LOG.debug("No longer using machine %s", instance_id)
            self.existing_machines[instance_id] = False
        else:
            # Remove the machine
            LOG.debug("Removing machine %s", instance_id)
            try:
                run(["juju", "remove-machine", instance_id, "--force"])
            except subprocess.CalledProcessError as e:
                raise HarnessError(f"failed to delete instance {instance_id}") from e

        self.instances.discard(instance_id)

    def cleanup(self):
        for instance_id in self.instances.copy():
            self.delete_instance(instance_id)