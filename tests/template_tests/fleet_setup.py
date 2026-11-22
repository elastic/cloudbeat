"""
Fleet setup for the template tests (CloudFormation CNVM and GCP Deployment Manager CSPM).

Run from the ``tests`` directory:

    poetry run python -m template_tests.fleet_setup cnvm
    poetry run python -m template_tests.fleet_setup cspm-gcp
    poetry run python -m template_tests.fleet_setup wait-agents

``cnvm`` and ``cspm-gcp`` create an agent policy and a package policy in Kibana, then write the config file
consumed by the template deployment (``deploy/cloudformation/config.json`` or
``deploy/deployment-manager/config.json``). ``cnvm`` also downloads the published CloudFormation template.
``wait-agents`` waits until the agent of the last created policy is enrolled.

Environment: ES_USER, ES_PASSWORD, KIBANA_URL, STACK_VERSION, AGENT_VERSION. GCP also reads DEPLOYMENT_NAME, ZONE,
ALLOW_SSH, GOOGLE_CLOUD_PROJECT and GOOGLE_APPLICATION_CREDENTIALS (or SERVICE_ACCOUNT_JSON_PATH, relative paths
are resolved against the current directory).

The enrollment token is a secret: it is masked (``::add-mask::``) before anything is written, and it must never be
logged. Do not add log statements that print the Fleet or config data.
"""

import argparse
import json
import os
import re
import sys
import time
from pathlib import Path
from urllib.parse import parse_qs, urlparse

import requests
import ruamel.yaml
from commonlib.kibana_api import get_agents, request_json
from loguru import logger
from munch import Munch

PACKAGE_NAME = "cloud_security_posture"
STREAM_NAME = f"{PACKAGE_NAME}.findings"
# Inputs whose stream needs an (empty) vars object.
INPUTS_WITH_STREAM_VARS = ["cloudbeat/cis_aws", "cloudbeat/cis_eks"]

REPO_ROOT = Path(__file__).resolve().parents[2]
CLOUDFORMATION_DIR = REPO_ROOT / "deploy" / "cloudformation"
CNVM_CONFIG = CLOUDFORMATION_DIR / "config.json"
CNVM_TEMPLATE = CLOUDFORMATION_DIR / "elastic-agent-ec2-cnvm.yml"
GCP_DM_CONFIG = REPO_ROOT / "deploy" / "deployment-manager" / "config.json"
STATE_FILE = Path(__file__).resolve().parent / ".fleet_state.json"

CNVM_AGENT_TAGS = ["cft_version:*", "cft_arn:arn:aws:cloudformation:.*"]
# "Key=<key>,Value=<value>" pairs separated by spaces.
CNVM_TAGS = (
    "Key=division,Value=engineering "
    "Key=org,Value=security "
    "Key=team,Value=contextual-security "
    "Key=project,Value=test-environments"
)

SNAPSHOT_ARTIFACTORY_URL = "https://snapshots.elastic.co/"
STAGING_ARTIFACTORY_URL = "https://staging.elastic.co/"
AGENT_ARTIFACT_SUFFIX = "/downloads/beats/elastic-agent"

HTTP_TIMEOUT_SEC = 60
WAIT_TIMEOUT_SEC = 600
WAIT_INTERVAL_SEC = 10


def load_config() -> Munch:
    """Read the Kibana connection settings from the environment."""
    cfg = Munch()
    cfg.kibana_url = os.getenv("KIBANA_URL", "")
    cfg.auth = (os.getenv("ES_USER", "NA"), os.getenv("ES_PASSWORD", "NA"))
    cfg.stack_version = os.getenv("STACK_VERSION", "")
    cfg.agent_version = os.getenv("AGENT_VERSION", "")
    return cfg


def mask(value: str):
    """Ask GitHub Actions to mask a secret in all later log output."""
    if value and os.getenv("GITHUB_ACTIONS") == "true":
        print(f"::add-mask::{value}", flush=True)


def write_json(path: Path, data: dict):
    """Write a JSON file readable only by the current user."""
    descriptor = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    with os.fdopen(descriptor, "w", encoding="utf-8") as file:
        json.dump(data, file)
    os.chmod(path, 0o600)


# ---------- Kibana (Fleet) ----------


def get_package(cfg: Munch) -> dict:
    """Latest (including prerelease) cloud_security_posture package. 9.x wraps it in ``item``, 8.x in ``response``."""
    body = request_json(
        "GET",
        cfg,
        f"/api/fleet/epm/packages/{PACKAGE_NAME}",
        params={"full": "true", "prerelease": "true"},
    )
    return body.get("item") or body.get("response") or {}


def build_package_policy(package: dict, name: str, input_name: str, posture: str, deployment: str, stream_vars=None):
    """
    Build the package policy payload: every input of the package disabled, except the one whose id contains
    ``input_name``, which is enabled together with its stream.
    """
    inputs = {}
    for template in package.get("policy_templates", []):
        data_stream = template["data_streams"][0]
        stream_id = f"{PACKAGE_NAME}.{data_stream}"
        for template_input in template.get("inputs", []):
            input_type = template_input["type"]
            input_id = f"{template['name']}-{input_type}"
            enabled = input_name in input_id
            stream = {"enabled": enabled}
            if input_type in INPUTS_WITH_STREAM_VARS:
                stream["vars"] = (stream_vars or {}) if enabled else {}
            elif enabled and stream_vars:
                stream["vars"] = stream_vars
            inputs[input_id] = {"enabled": enabled, "streams": {stream_id: stream}}

    package_vars = {var["name"]: "" for var in package.get("vars", [])}
    package_vars.update({"posture": posture, "deployment": deployment})
    return {
        "policy_id": "",
        "package": {"name": package.get("name", ""), "version": package.get("version", "")},
        "name": name,
        "description": "",
        "namespace": "default",
        "inputs": inputs,
        "vars": package_vars,
    }


def create_agent_policy(cfg: Munch, name: str) -> str:
    """Create an agent policy and return its id."""
    body = request_json(
        "POST",
        cfg,
        "/api/fleet/agent_policies",
        json={"name": name, "namespace": "default", "monitoring_enabled": ["logs", "metrics"]},
    )
    return body["item"]["id"]


def create_package_policy(cfg: Munch, package_policy: dict, agent_policy_id: str) -> str:
    """Create the package policy on an agent policy and return its id."""
    package_policy["policy_id"] = agent_policy_id
    body = request_json("POST", cfg, "/api/fleet/package_policies", json=package_policy)
    return (body.get("item") or body.get("response", {}).get("item") or {}).get("id", "")


def install_package(cfg: Munch, version: str):
    """Install the package version (needed for the GCP integration)."""
    request_json(
        "POST",
        cfg,
        f"/api/fleet/epm/packages/{PACKAGE_NAME}/{version}",
        json={"force": True, "ignore_constraints": True},
    )


def get_enrollment_token(cfg: Munch, policy_id: str) -> str:
    """Enrollment token of an agent policy. Returns an empty string when there is none."""
    body = request_json("GET", cfg, "/api/fleet/enrollment_api_keys")
    for key in body.get("items") or body.get("list") or []:
        if key.get("policy_id") == policy_id:
            return key.get("api_key", "")
    return ""


def get_fleet_server_host(cfg: Munch) -> str:
    """URL of the default Fleet Server."""
    body = request_json("GET", cfg, "/api/fleet/fleet_server_hosts")
    for server in body.get("items", []):
        if server.get("is_default"):
            return server["host_urls"][0]
    raise KeyError("No default fleet server found")


def get_artifact_server(version: str) -> str:
    """Artifact server URL of the build of an Elastic version (SNAPSHOT builds are on snapshots.elastic.co)."""
    base = SNAPSHOT_ARTIFACTORY_URL if "SNAPSHOT" in version else STAGING_ARTIFACTORY_URL
    response = requests.get(f"{base}latest/{version}.json", timeout=HTTP_TIMEOUT_SEC)
    response.raise_for_status()
    return f"{base}{response.json()['build_id']}{AGENT_ARTIFACT_SUFFIX}"


# ---------- CNVM template ----------


def extract_template_url(url_string: str) -> str:
    """Value of the ``templateURL`` query parameter, or an empty string."""
    values = parse_qs(urlparse(url_string, allow_fragments=False).query).get("templateURL")
    return values[0] if values else ""


def get_package_template_url(package: dict, template_name: str, input_type: str) -> str:
    """The ``templateURL`` of the default value of the first var of a package input."""
    for template in package.get("policy_templates", []):
        if template.get("name") != template_name:
            continue
        for template_input in template.get("inputs", []):
            if template_input.get("type") == input_type and template_input.get("vars"):
                return extract_template_url(template_input["vars"][0].get("default", ""))
    return ""


def add_tags(tags: str, yaml_content: str) -> str:
    """
    Add custom tags to every EC2 instance of a CloudFormation template, preserving the formatting.

    Args:
        tags: ``Key=<key>,Value=<value>`` pairs separated by spaces.
        yaml_content: The template.
    """
    yaml = ruamel.yaml.YAML()
    yaml.preserve_quotes = True
    yaml.explicit_start = True
    yaml.indent(mapping=2, sequence=4, offset=2)
    template = yaml.load(yaml_content)

    tag_dicts = [dict(key_value.split("=", 1) for key_value in tag.split(",")) for tag in tags.split()]
    for resource in template["Resources"].values():
        if resource["Type"] == "AWS::EC2::Instance":
            properties = resource.setdefault("Properties", {})
            properties.setdefault("Tags", [])
            properties["Tags"] += tag_dicts

    output = ruamel.yaml.compat.StringIO()
    yaml.dump(template, output)
    return output.getvalue()


def download_template(url: str) -> str:
    """Download the published CloudFormation template."""
    response = requests.get(url, timeout=HTTP_TIMEOUT_SEC)
    response.raise_for_status()
    return response.text


def install_published_template(url: str, path: Path, tags: str):
    """Replace the template at ``path`` with the published one (tagged), keeping the original as ``*-orig``."""
    content = add_tags(tags=tags, yaml_content=download_template(url))
    if path.exists():
        path.rename(path.with_name(f"{path.stem}-orig{path.suffix}"))
    path.write_text(content, encoding="utf-8")
    logger.info(f"CNVM template is available at: '{path}'")


# ---------- State, commands ----------


def save_state(policy_id: str, expected_agents: int, expected_tags: list):
    """Remember what ``wait-agents`` must wait for."""
    write_json(
        STATE_FILE,
        {"policy_id": policy_id, "expected_agents": expected_agents, "expected_tags": expected_tags},
    )


def random_name(prefix: str) -> str:
    """Policy name unique per run."""
    return f"{prefix}-{os.urandom(4).hex()}"


def setup_cnvm(cfg: Munch):
    """Create the CNVM AWS integration and the CloudFormation config."""
    package = get_package(cfg)
    logger.info(f"Package version: {package.get('version')}")
    agent_policy_id = create_agent_policy(cfg, random_name("cnvm-aws"))
    package_policy = build_package_policy(
        package,
        name=random_name("pkg-cnvm-aws"),
        input_name="vuln_mgmt_aws",
        posture="vuln_mgmt",
        deployment="aws",
    )
    package_policy_id = create_package_policy(cfg, package_policy, agent_policy_id)
    logger.info(f"Created CNVM AWS package policy '{package_policy_id}' on agent policy '{agent_policy_id}'")
    save_state(agent_policy_id, 1, CNVM_AGENT_TAGS)

    token = get_enrollment_token(cfg, agent_policy_id)
    mask(token)
    write_json(
        CNVM_CONFIG,
        {
            "ENROLLMENT_TOKEN": token,
            "FLEET_URL": get_fleet_server_host(cfg),
            "ELASTIC_AGENT_VERSION": cfg.agent_version,
            "ELASTIC_ARTIFACT_SERVER": get_artifact_server(cfg.agent_version),
        },
    )

    template_url = get_package_template_url(package, "vuln_mgmt", "cloudbeat/vuln_mgmt_aws")
    if template_url:
        logger.info(f"Using {template_url} for stack creation")
        install_published_template(template_url, CNVM_TEMPLATE, os.getenv("AWS_CNVM_TAGS", CNVM_TAGS))
    else:
        logger.warning("templateURL not found in the package, using the template from the repository")


def gcp_credentials() -> dict:
    """Service account (or ADC credentials) JSON used by the integration."""
    path = os.getenv("SERVICE_ACCOUNT_JSON_PATH") or os.getenv("GOOGLE_APPLICATION_CREDENTIALS", "")
    if not path:
        raise ValueError("Set SERVICE_ACCOUNT_JSON_PATH or GOOGLE_APPLICATION_CREDENTIALS")
    return json.loads(Path(path).read_text(encoding="utf-8"))


def setup_cspm_gcp(cfg: Munch):
    """Create the CSPM GCP integration and the Deployment Manager config."""
    package = get_package(cfg)
    version = package.get("version", "")
    logger.info(f"Package version: {version}")
    install_package(cfg, version)

    stream_vars = {
        "gcp.account_type": "single-account",
        "gcp.credentials.json": json.dumps(gcp_credentials()),
        "gcp.project_id": os.getenv("GOOGLE_CLOUD_PROJECT", ""),
        "gcp.credentials.type": "credentials-json",
    }
    agent_policy_id = create_agent_policy(cfg, random_name("cspm-gcp"))
    package_policy = build_package_policy(
        package,
        name=random_name("pkg-cspm-gcp"),
        input_name="cis_gcp",
        posture="cspm",
        deployment="gcp",
        stream_vars=stream_vars,
    )
    package_policy_id = create_package_policy(cfg, package_policy, agent_policy_id)
    logger.info(f"Created CSPM GCP package policy '{package_policy_id}' on agent policy '{agent_policy_id}'")
    save_state(agent_policy_id, 1, [])

    token = get_enrollment_token(cfg, agent_policy_id)
    mask(token)
    write_json(
        GCP_DM_CONFIG,
        {
            "ENROLLMENT_TOKEN": token,
            "FLEET_URL": get_fleet_server_host(cfg),
            "ELASTIC_ARTIFACT_SERVER": get_artifact_server(cfg.agent_version),
            "DEPLOYMENT_NAME": os.getenv("DEPLOYMENT_NAME", ""),
            "ZONE": os.getenv("ZONE", "us-central1-a"),
            "ALLOW_SSH": os.getenv("ALLOW_SSH", "false") == "true",
            "STACK_VERSION": cfg.agent_version,
        },
    )


def agents_enrolled(agents: list, policy_id: str, expected_agents: int, expected_tags: list) -> bool:
    """True when exactly ``expected_agents`` agents of the policy have every expected tag (regex, from the start)."""
    count = 0
    for agent in agents:
        if agent.get("policy_id") != policy_id:
            continue
        tags = agent.get("tags") or []
        if all(any(re.match(pattern, tag) for tag in tags) for pattern in expected_tags):
            count += 1
    logger.info(f"Policy {policy_id}: {count}/{expected_agents} agents enrolled")
    return count == expected_agents


def wait_agents(cfg: Munch, timeout: int = WAIT_TIMEOUT_SEC, interval: int = WAIT_INTERVAL_SEC) -> bool:
    """Wait until the agents of the last created policy are enrolled."""
    state = json.loads(STATE_FILE.read_text(encoding="utf-8"))
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        if agents_enrolled(
            get_agents(cfg),
            state["policy_id"],
            state["expected_agents"],
            state["expected_tags"],
        ):
            return True
        time.sleep(interval)
    return False


def main(argv=None) -> int:
    """Entry point."""
    parser = argparse.ArgumentParser(description=__doc__.split("\n\n", maxsplit=1)[0])
    parser.add_argument("command", choices=["cnvm", "cspm-gcp", "wait-agents"])
    parser.add_argument("--timeout", type=int, default=WAIT_TIMEOUT_SEC, help="wait-agents timeout in seconds")
    args = parser.parse_args(argv)

    cfg = load_config()
    if args.command == "cnvm":
        setup_cnvm(cfg)
    elif args.command == "cspm-gcp":
        setup_cspm_gcp(cfg)
    elif not wait_agents(cfg, timeout=args.timeout):
        logger.error("Not all agents were enrolled")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
