"""
Offline unit tests for fleet_setup.py (no Kibana, AWS or GCP needed).

Run from the ``tests`` directory:

    poetry run python -m pytest --noconftest -o addopts="" template_tests/test_fleet_setup.py
"""

# pylint: disable=missing-function-docstring

import copy
import json
import stat
from pathlib import Path
from unittest import mock

import pytest
import ruamel.yaml
from munch import Munch, munchify

from template_tests import fleet_setup

FIXTURES = Path(__file__).parent / "fixtures"
REPO_CNVM_TEMPLATE = fleet_setup.REPO_ROOT / "deploy" / "cloudformation" / "elastic-agent-ec2-cnvm.yml"
TOKEN = "SECRET-ENROLLMENT-TOKEN"
TEMPLATE_URL = "https://example.com/cloudformation-cnvm.yml"
QUICKCREATE_URL = "https://console.aws.amazon.com/cloudformation/home#/stacks/quickcreate"
# The values used to generate golden_gcp.json with the previous generator.
GCP_STREAM_VARS = {
    "gcp.account_type": "single-account",
    "gcp.credentials.json": '{"type": "service_account"}',
    "gcp.project_id": "my-project",
    "gcp.credentials.type": "credentials-json",
}


def load_fixture(name: str) -> dict:
    return json.loads((FIXTURES / name).read_text(encoding="utf-8"))


def package_with_template_url() -> dict:
    package = load_fixture("package_cloud_security_posture.json")
    for template in package["policy_templates"]:
        if template["name"] == "vuln_mgmt":
            template["inputs"][0]["vars"] = [
                {"default": f"{QUICKCREATE_URL}?templateURL={TEMPLATE_URL}"},
            ]
    return package


class FakeKibana:
    """Answers the Fleet calls of fleet_setup and records them."""

    def __init__(self, package: dict, wrap_old_shapes: bool = False):
        self.package = package
        self.old = wrap_old_shapes
        self.calls = []

    def __call__(self, method, cfg, path, **kwargs):  # pylint: disable=unused-argument
        self.calls.append((method, path, kwargs))
        if method == "GET" and path == "/api/fleet/epm/packages/cloud_security_posture":
            return {"response": self.package} if self.old else {"item": self.package}
        if path == "/api/fleet/agent_policies":
            return {"item": {"id": "agent-policy-1"}}
        if path == "/api/fleet/package_policies":
            item = {"id": "package-policy-1"}
            return {"response": {"item": item}} if self.old else {"item": item}
        if method == "POST" and path.startswith("/api/fleet/epm/packages/cloud_security_posture/"):
            return {}
        if path == "/api/fleet/enrollment_api_keys":
            keys = [{"policy_id": "other", "api_key": "wrong"}, {"policy_id": "agent-policy-1", "api_key": TOKEN}]
            return {"list": keys} if self.old else {"items": keys}
        if path == "/api/fleet/fleet_server_hosts":
            return {
                "items": [
                    {"is_default": False, "host_urls": ["https://not-default:443"]},
                    {"is_default": True, "host_urls": ["https://fleet.example:443"]},
                ],
            }
        raise AssertionError(f"unexpected call {method} {path}")

    def sent(self, path: str) -> dict:
        return next(
            kwargs["json"] for method, call_path, kwargs in self.calls if call_path == path and method == "POST"
        )


def fake_http_get(template_text: str, fail_template: bool = False):
    def get(url, timeout=None):  # pylint: disable=unused-argument
        response = mock.Mock()
        if "/latest/" in url:
            response.json.return_value = {"build_id": "9.9.9-abc123"}
        elif fail_template:
            response.raise_for_status.side_effect = RuntimeError("download failed")
        else:
            response.text = template_text
        return response

    return get


@pytest.fixture(name="workdir")
def fixture_workdir(tmp_path, monkeypatch):
    """Redirect every file fleet_setup writes into a temp dir and set the environment."""
    monkeypatch.setattr(fleet_setup, "CNVM_CONFIG", tmp_path / "cnvm-config.json")
    monkeypatch.setattr(fleet_setup, "CNVM_TEMPLATE", tmp_path / "elastic-agent-ec2-cnvm.yml")
    monkeypatch.setattr(fleet_setup, "GCP_DM_CONFIG", tmp_path / "gcp-config.json")
    monkeypatch.setattr(fleet_setup, "STATE_FILE", tmp_path / ".fleet_state.json")
    (tmp_path / "elastic-agent-ec2-cnvm.yml").write_text(
        REPO_CNVM_TEMPLATE.read_text(encoding="utf-8"),
        encoding="utf-8",
    )
    credentials = tmp_path / "KEY_FILE.json"
    credentials.write_text('{"type": "service_account"}', encoding="utf-8")
    monkeypatch.setenv("KIBANA_URL", "https://kibana.example")
    monkeypatch.setenv("ES_USER", "elastic")
    monkeypatch.setenv("ES_PASSWORD", "pw")
    monkeypatch.setenv("AGENT_VERSION", "9.6.0-SNAPSHOT")
    monkeypatch.setenv("SERVICE_ACCOUNT_JSON_PATH", str(credentials))
    monkeypatch.setenv("GOOGLE_CLOUD_PROJECT", "my-project")
    monkeypatch.setenv("DEPLOYMENT_NAME", "dm-test")
    monkeypatch.delenv("ALLOW_SSH", raising=False)
    monkeypatch.delenv("ZONE", raising=False)
    monkeypatch.delenv("AWS_CNVM_TAGS", raising=False)
    monkeypatch.delenv("GITHUB_ACTIONS", raising=False)
    return tmp_path


def instance_tags(template_text: str) -> list:
    template = ruamel.yaml.YAML().load(template_text)
    for resource in template["Resources"].values():
        if resource["Type"] == "AWS::EC2::Instance":
            return [
                (tag["Key"], tag["Value"]) for tag in resource["Properties"]["Tags"] if isinstance(tag["Value"], str)
            ]
    return []


# ---------- U1: payload equivalence with the previous generator ----------


@pytest.mark.parametrize(
    "golden,kwargs",
    [
        (
            "golden_gcp.json",
            {"input_name": "cis_gcp", "posture": "cspm", "deployment": "gcp", "stream_vars": GCP_STREAM_VARS},
        ),
        ("golden_cnvm.json", {"input_name": "vuln_mgmt_aws", "posture": "vuln_mgmt", "deployment": "aws"}),
        ("golden_aws.json", {"input_name": "cis_aws", "posture": "cspm", "deployment": "aws"}),
    ],
)
def test_package_policy_matches_previous_generator(golden, kwargs):
    package = load_fixture("package_cloud_security_posture.json")
    policy = fleet_setup.build_package_policy(package, name="pkg-test", **kwargs)
    assert policy == load_fixture(golden)


# ---------- U2: response shapes of 9.x and 8.x ----------


@pytest.mark.parametrize("old", [False, True])
def test_response_shapes(old):
    package = load_fixture("package_cloud_security_posture.json")
    kibana = FakeKibana(package, wrap_old_shapes=old)
    cfg = fleet_setup.load_config()
    with mock.patch.object(fleet_setup, "request_json", kibana):
        assert fleet_setup.get_package(cfg) == package
        assert fleet_setup.create_package_policy(cfg, {}, "agent-policy-1") == "package-policy-1"
        assert fleet_setup.get_enrollment_token(cfg, "agent-policy-1") == TOKEN
        assert fleet_setup.get_enrollment_token(cfg, "missing") == ""
        assert fleet_setup.get_fleet_server_host(cfg) == "https://fleet.example:443"


def test_artifact_server():
    with mock.patch.object(fleet_setup.requests, "get", fake_http_get("")):
        assert (
            fleet_setup.get_artifact_server("9.6.0-SNAPSHOT")
            == "https://snapshots.elastic.co/9.9.9-abc123/downloads/beats/elastic-agent"
        )
        assert fleet_setup.get_artifact_server("9.6.0").startswith("https://staging.elastic.co/9.9.9-abc123/")


# ---------- U3: config files ----------


def test_cnvm_config_and_state(workdir):
    kibana = FakeKibana(package_with_template_url())
    with (
        mock.patch.object(fleet_setup, "request_json", kibana),
        mock.patch.object(fleet_setup.requests, "get", fake_http_get(REPO_CNVM_TEMPLATE.read_text(encoding="utf-8"))),
    ):
        fleet_setup.setup_cnvm(fleet_setup.load_config())

    config = json.loads((workdir / "cnvm-config.json").read_text())
    assert config == {
        "ENROLLMENT_TOKEN": TOKEN,
        "FLEET_URL": "https://fleet.example:443",
        "ELASTIC_AGENT_VERSION": "9.6.0-SNAPSHOT",
        "ELASTIC_ARTIFACT_SERVER": "https://snapshots.elastic.co/9.9.9-abc123/downloads/beats/elastic-agent",
    }
    assert json.loads((workdir / ".fleet_state.json").read_text()) == {
        "policy_id": "agent-policy-1",
        "expected_agents": 1,
        "expected_tags": fleet_setup.CNVM_AGENT_TAGS,
    }
    policy = kibana.sent("/api/fleet/package_policies")
    assert policy["policy_id"] == "agent-policy-1"
    assert policy["inputs"]["vuln_mgmt-cloudbeat/vuln_mgmt_aws"]["enabled"] is True
    assert policy["inputs"]["cspm-cloudbeat/cis_gcp"]["enabled"] is False
    assert ("POST", "/api/fleet/epm/packages/cloud_security_posture/3.5.1") not in [(m, p) for m, p, _ in kibana.calls]


def test_gcp_config_and_state(workdir):
    kibana = FakeKibana(load_fixture("package_cloud_security_posture.json"))
    with (
        mock.patch.object(fleet_setup, "request_json", kibana),
        mock.patch.object(fleet_setup.requests, "get", fake_http_get("")),
    ):
        fleet_setup.setup_cspm_gcp(fleet_setup.load_config())

    config = json.loads((workdir / "gcp-config.json").read_text())
    assert config == {
        "ENROLLMENT_TOKEN": TOKEN,
        "FLEET_URL": "https://fleet.example:443",
        "ELASTIC_ARTIFACT_SERVER": "https://snapshots.elastic.co/9.9.9-abc123/downloads/beats/elastic-agent",
        "DEPLOYMENT_NAME": "dm-test",
        "ZONE": "us-central1-a",
        "ALLOW_SSH": False,
        "STACK_VERSION": "9.6.0-SNAPSHOT",
    }
    assert json.loads((workdir / ".fleet_state.json").read_text())["expected_tags"] == []
    stream = kibana.sent("/api/fleet/package_policies")["inputs"]["cspm-cloudbeat/cis_gcp"]["streams"][
        fleet_setup.STREAM_NAME
    ]
    assert stream["vars"] == GCP_STREAM_VARS
    assert kibana.sent("/api/fleet/epm/packages/cloud_security_posture/3.5.1") == {
        "force": True,
        "ignore_constraints": True,
    }


# ---------- U4: published template ----------


def test_extract_template_url():
    url = (
        f"https://console.aws.amazon.com/cloudformation/home#/stacks/quickcreate?templateURL={TEMPLATE_URL}&stackName=x"
    )
    assert fleet_setup.extract_template_url(url) == TEMPLATE_URL
    assert fleet_setup.extract_template_url("https://example.com/no-template") == ""


def test_add_tags_keeps_existing_tags():
    tagged = fleet_setup.add_tags(fleet_setup.CNVM_TAGS, REPO_CNVM_TEMPLATE.read_text(encoding="utf-8"))
    tags = dict(instance_tags(tagged))
    assert tags["Task"] == "Vulnerability Management Scanner"
    assert {key: tags[key] for key in ("division", "org", "team", "project")} == {
        "division": "engineering",
        "org": "security",
        "team": "contextual-security",
        "project": "test-environments",
    }
    assert "!Join" in tagged  # CloudFormation intrinsic functions survive the round trip


def test_published_template_replaces_repo_template(workdir):
    original = (workdir / "elastic-agent-ec2-cnvm.yml").read_text()
    with mock.patch.object(fleet_setup.requests, "get", fake_http_get(original)):
        fleet_setup.install_published_template(
            TEMPLATE_URL,
            workdir / "elastic-agent-ec2-cnvm.yml",
            fleet_setup.CNVM_TAGS,
        )
    assert (workdir / "elastic-agent-ec2-cnvm-orig.yml").read_text() == original
    assert ("project", "test-environments") in instance_tags((workdir / "elastic-agent-ec2-cnvm.yml").read_text())


# ---------- U5: secrets ----------


@pytest.mark.usefixtures("workdir")
def test_token_is_masked_before_it_is_written_and_never_logged(monkeypatch, capfd):
    monkeypatch.setenv("GITHUB_ACTIONS", "true")
    seen_before_write = []
    real_write_json = fleet_setup.write_json

    def checking_write_json(path, data):
        if "ENROLLMENT_TOKEN" in data:
            seen_before_write.append(capfd.readouterr().out)
        real_write_json(path, data)

    kibana = FakeKibana(package_with_template_url())
    with (
        mock.patch.object(fleet_setup, "request_json", kibana),
        mock.patch.object(fleet_setup.requests, "get", fake_http_get(REPO_CNVM_TEMPLATE.read_text(encoding="utf-8"))),
        mock.patch.object(fleet_setup, "write_json", checking_write_json),
    ):
        fleet_setup.setup_cnvm(fleet_setup.load_config())

    assert f"::add-mask::{TOKEN}" in seen_before_write[0]
    rest = capfd.readouterr()
    assert TOKEN not in rest.err
    assert TOKEN not in (rest.out.replace(f"::add-mask::{TOKEN}", ""))
    assert TOKEN not in "".join(seen_before_write).replace(f"::add-mask::{TOKEN}", "")


def test_written_files_are_private(workdir):
    kibana = FakeKibana(package_with_template_url())
    with (
        mock.patch.object(fleet_setup, "request_json", kibana),
        mock.patch.object(fleet_setup.requests, "get", fake_http_get(REPO_CNVM_TEMPLATE.read_text(encoding="utf-8"))),
    ):
        fleet_setup.setup_cnvm(fleet_setup.load_config())
    for name in ("cnvm-config.json", ".fleet_state.json"):
        assert stat.S_IMODE((workdir / name).stat().st_mode) == 0o600


# ---------- U6: error paths ----------


def test_kibana_failure_is_not_swallowed(workdir):
    with mock.patch.object(
        fleet_setup,
        "request_json",
        side_effect=RuntimeError("Kibana API GET failed: status code 401"),
    ):
        with pytest.raises(RuntimeError):
            fleet_setup.main(["cnvm"])
    assert not (workdir / "cnvm-config.json").exists()


def test_missing_template_url_keeps_repo_template(workdir):
    original = (workdir / "elastic-agent-ec2-cnvm.yml").read_text()
    kibana = FakeKibana(load_fixture("package_cloud_security_posture.json"))  # no templateURL
    with (
        mock.patch.object(fleet_setup, "request_json", kibana),
        mock.patch.object(fleet_setup.requests, "get", fake_http_get("")),
    ):
        fleet_setup.setup_cnvm(fleet_setup.load_config())
    assert (workdir / "elastic-agent-ec2-cnvm.yml").read_text() == original
    assert not (workdir / "elastic-agent-ec2-cnvm-orig.yml").exists()


def test_failed_template_download_keeps_repo_template(workdir):
    original = (workdir / "elastic-agent-ec2-cnvm.yml").read_text()
    with mock.patch.object(fleet_setup.requests, "get", fake_http_get("", fail_template=True)):
        with pytest.raises(RuntimeError):
            fleet_setup.install_published_template(
                TEMPLATE_URL,
                workdir / "elastic-agent-ec2-cnvm.yml",
                fleet_setup.CNVM_TAGS,
            )
    assert (workdir / "elastic-agent-ec2-cnvm.yml").read_text() == original
    assert not (workdir / "elastic-agent-ec2-cnvm-orig.yml").exists()


@pytest.mark.usefixtures("workdir")
def test_gcp_requires_credentials(monkeypatch):
    monkeypatch.delenv("SERVICE_ACCOUNT_JSON_PATH")
    monkeypatch.delenv("GOOGLE_APPLICATION_CREDENTIALS", raising=False)
    with pytest.raises(ValueError):
        fleet_setup.gcp_credentials()


# ---------- U7: waiting for agents ----------


def agent(policy_id, tags):
    return munchify({"id": "a", "policy_id": policy_id, "tags": tags})


def test_agents_enrolled_counts_policy_and_tags():
    tags = fleet_setup.CNVM_AGENT_TAGS
    good = agent("p1", ["cft_version:9.6.0", "cft_arn:arn:aws:cloudformation:eu-west-1:1:stack/x"])
    assert fleet_setup.agents_enrolled([good], "p1", 1, tags)
    assert not fleet_setup.agents_enrolled([], "p1", 1, tags)
    assert not fleet_setup.agents_enrolled([agent("p2", good.tags)], "p1", 1, tags)  # other policy
    assert not fleet_setup.agents_enrolled([agent("p1", ["cft_version:9.6.0"])], "p1", 1, tags)  # missing tag
    assert not fleet_setup.agents_enrolled([good, good], "p1", 1, tags)  # more than expected
    assert fleet_setup.agents_enrolled([agent("p1", [])], "p1", 1, [])  # GCP: no tag requirements
    assert fleet_setup.agents_enrolled([Munch(id="a", policy_id="p1")], "p1", 1, [])  # agent without tags


@pytest.mark.usefixtures("workdir")
def test_wait_agents_success_and_timeout(monkeypatch):
    fleet_setup.save_state("p1", 1, [])
    cfg = fleet_setup.load_config()
    monkeypatch.setattr(fleet_setup, "get_agents", lambda cfg: [agent("p1", [])])
    assert fleet_setup.wait_agents(cfg, timeout=1, interval=0.01)
    assert fleet_setup.main(["wait-agents", "--timeout", "1"]) == 0

    monkeypatch.setattr(fleet_setup, "get_agents", lambda cfg: [])
    assert not fleet_setup.wait_agents(cfg, timeout=0.05, interval=0.01)
    assert fleet_setup.main(["wait-agents", "--timeout", "0"]) == 1


def test_build_package_policy_does_not_mutate_input():
    package = load_fixture("package_cloud_security_posture.json")
    before = copy.deepcopy(package)
    fleet_setup.build_package_policy(package, "n", "cis_gcp", "cspm", "gcp", GCP_STREAM_VARS)
    assert package == before
