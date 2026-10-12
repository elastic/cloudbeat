#!/usr/bin/env python3
"""Replay one Defend episode onto an enrolled endpoint.

The bundled fixture is the SRVNIX05 half of Oh My Malware episode 5
(https://ohmymalware.com/). SRVNIX05 is the Linux host in that capture.
This script rewrites that chain onto the enrolled endpoint you pass in
and renames the archive, payload, binary, rule, and shared hashes so the
replay does not join the original episode. SRVNIX05-PRIV is not in the fixture.

Auth is an API key or a username and password. The Elasticsearch bulk path is
the same on serverless and on a classic stack. --deployment only changes the
setup notes this script prints.

  python3 replay.py --help
  ES_URL=... ES_API_KEY=... python3 replay.py \\
    --host-name HOST --agent-id AGENT --host-id HOSTID --run-id demo1
  ES_URL=... ES_USER=elastic ES_PASSWORD=... python3 replay.py \\
    --deployment stack --host-name HOST --agent-id AGENT --host-id HOSTID \\
    --run-id demo1 --upload --install-rule --kibana-url https://...
"""

from __future__ import annotations

import argparse
import base64
import getpass
import hashlib
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

HERE = Path(__file__).resolve().parent
FIXTURES = HERE / "fixtures"
SRC_EVENTS = FIXTURES / "source-events.ndjson"
SRC_ALERTS = FIXTURES / "source-alerts.ndjson"
MAPPING = FIXTURES / "mapping.json"
RULE_FILE = FIXTURES / "detection-rule.ndjson"

OLD_HOST = "SRVNIX05"
OLD_AGENT = "79ad2620-720c-4aa3-8877-816f6aef87bc"
OLD_HOST_ID = "909950ee14d641e8bcba6ffeb9db2e7b"
OLD_PAYLOAD = "74ef6cc38f5a1a80148752b63c117e6846984debd2af806c65887195a8eccc56"
OLD_BINARY = "kdmtmpflush"
OLD_ARCHIVE = "9415656314.zip"
OLD_RULE = "Linux.Trojan.BPFDoor"
RULE_ID = "61e90241-c8f2-47bc-8e47-238420a34fb6"
PIPELINE = "insights_pipeline"

SHARED_HASHES = [
    "caef04d258ec8e7089227c69ffe0407c",
    "d80c31353a4f404d56fa7d10efa15705",
    "c96c2246fd1b5bf80fccc2c733763bb2",
    "c4b73a5f6e20144f3d1c250262661b5f",
    "11227b11f565de042c48654a241e9d1c",
    "351f929d8f9ed915bc19a96d904cfa49",
    "eeb83d755e54ff08a3f085c3e641fe65",
    "c84305c175251af7c7bca4b1856597d8",
    "19143d7259b7b19b2a02ea0519f0302a181014d3",
    "25ab0fb1fd822d69ab665f247db8800e34f5043a",
    "aa5a7327c63f60ca3ea75a6d9ffd59d87b62229b",
    "31adee239fb9cfeb7f0416fe97e4e7de1372d2d6",
    "6b6843dae13e009a3267eed8f472bebe47e39e31",
    "143ff5589e99eefe8ad5664d110269139916ce42",
    "3100ce15e35c3e45cfac8cf77392ef58a0e8f0c6",
    "50e5caec70401734574a9fdefc068b5b61740bf8",
    "bafe6f949e99897a01024d66650cccce36a1eb982cb100cfbe56a686830b9e10",
    "b88ffa26bc4f1a1c1a53bb6580a0d77ad418f324ea82138c610a1b6d5950b0f3",
    "2003c43060a03af33ff3ef5ba4b700f36eb85e17d7947b285bdece53b1cd3d43",
    "391e6221124df1be64aa64dc84e094dffd2165f6a708ee1db29aabffa0559aaf",
    "b4280c4589f0bcabdd2db596552883a33bd67a41653bade2ff17d9ac49e0d7a1",
    "8677919550a90ff7106584285f25a70ac9e7aa38bdb4ed17d34bbfb366fd71b7",
    "206204d48c4861f754ca766805aaa8410b5674da552e091e5fff25ee2c0ea523",
    "2c336c63e26881d2f02f34379024e7c314bce572c08cbaa319bacbbec29f93ed",
]

DATASETS = {
    "endpoint.events.process": "logs-endpoint.events.process-default",
    "endpoint.events.file": "logs-endpoint.events.file-default",
    "endpoint.events.network": "logs-endpoint.events.network-default",
}


def die(message: str) -> None:
    sys.exit(message)


def rewrite_entity_id(value: str, new_agent: str, new_host_id: str) -> str | None:
    pad = "=" * ((4 - len(value) % 4) % 4)
    try:
        text = base64.b64decode(value + pad, validate=True).decode("utf-8")
    except (ValueError, UnicodeDecodeError):
        return None
    if OLD_AGENT not in text and OLD_HOST_ID not in text:
        return None
    text = text.replace(OLD_AGENT, new_agent).replace(OLD_HOST_ID, new_host_id)
    encoded = base64.b64encode(text.encode()).decode()
    return encoded.rstrip("=") if "=" not in value else encoded


def rewrite_string(value: str, subs: list[tuple[str, str]], hash_map: dict[str, str], new_agent: str, new_host_id: str) -> str:
    rewritten = rewrite_entity_id(value, new_agent, new_host_id)
    if rewritten is not None:
        return rewritten
    for old, new in subs:
        if old in value:
            value = value.replace(old, new)
    return hash_map.get(value, value)


def rewrite(node, subs, hash_map, new_agent, new_host_id):
    if isinstance(node, dict):
        return {key: rewrite(value, subs, hash_map, new_agent, new_host_id) for key, value in node.items()}
    if isinstance(node, list):
        return [rewrite(value, subs, hash_map, new_agent, new_host_id) for value in node]
    if isinstance(node, str):
        return rewrite_string(node, subs, hash_map, new_agent, new_host_id)
    return node


def stamp(doc: dict, host_name: str, agent_id: str, host_id: str, agent_name: str, tag: str) -> dict:
    host = doc.setdefault("host", {})
    host["name"] = host_name
    host["hostname"] = host_name
    host["id"] = host_id
    agent = doc.setdefault("agent", {})
    agent["id"] = agent_id
    agent["name"] = agent_name
    elastic = doc.get("elastic")
    if isinstance(elastic, dict) and isinstance(elastic.get("agent"), dict):
        elastic["agent"]["id"] = agent_id
    tags = doc.get("tags")
    if not isinstance(tags, list):
        tags = []
    if tag not in tags:
        tags.append(tag)
    doc["tags"] = tags
    return doc


def load_docs(path: Path, subs, hash_map, host_name, agent_id, host_id, agent_name, tag) -> list[dict]:
    docs = []
    with path.open() as fh:
        for line in fh:
            line = line.strip()
            if not line:
                continue
            doc = json.loads(line)
            host = doc.get("host") if isinstance(doc.get("host"), dict) else {}
            if host.get("name") != OLD_HOST:
                continue
            docs.append(stamp(rewrite(doc, subs, hash_map, agent_id, host_id), host_name, agent_id, host_id, agent_name, tag))
    return docs


def dataset_of(doc: dict) -> str | None:
    stream = doc.get("data_stream")
    if isinstance(stream, dict) and stream.get("dataset"):
        return stream["dataset"]
    event = doc.get("event")
    return event.get("dataset") if isinstance(event, dict) else None


def bulk_body(index: str, docs: list[dict]) -> bytes:
    lines = []
    for doc in docs:
        lines.append(json.dumps({"create": {"_index": index}}, separators=(",", ":")))
        lines.append(json.dumps(doc, separators=(",", ":")))
    return ("\n".join(lines) + "\n").encode()


def replace_exact(node, old: str, new: str):
    if isinstance(node, dict):
        return {key: replace_exact(value, old, new) for key, value in node.items()}
    if isinstance(node, list):
        return [replace_exact(value, old, new) for value in node]
    if isinstance(node, str):
        return node.replace(old, new)
    return node


def split_alert_rules(alerts: list[dict], base_rule: str) -> tuple[str, str]:
    """Give one replay two endpoint rule names.

    Attack Discovery refuses a chain made of a single rule. Episode 5's alerts
    are all Linux.Trojan.BPFDoor, so the drop (unzip) and the execute (/dev/shm)
    each get their own name. Both stay on the enrolled host.
    """
    drop_rule = f"{base_rule}.Drop"
    exec_rule = f"{base_rule}.Exec"
    for alert in alerts:
        proc = alert.get("process") if isinstance(alert.get("process"), dict) else {}
        file = alert.get("file") if isinstance(alert.get("file"), dict) else {}
        command = proc.get("command_line") or ""
        path = file.get("path") or ""
        is_drop = proc.get("name") == "unzip" or command.startswith("unzip ") or (
            "/home/" in path and "/dev/shm/" not in path
        )
        chosen = drop_rule if is_drop else exec_rule
        rewritten = replace_exact(alert, base_rule, chosen)
        alert.clear()
        alert.update(rewritten)
    return drop_rule, exec_rule


def assert_clean(docs: list[dict], run_id: str) -> None:
    blob = "\n".join(json.dumps(doc) for doc in docs)
    forbidden = [OLD_HOST, "SRVNIX05-PRIV", OLD_BINARY, "9415656314", OLD_PAYLOAD, OLD_AGENT, OLD_HOST_ID, OLD_RULE, *SHARED_HASHES]
    found = [token for token in forbidden if token in blob]
    if found:
        die("replay still contains original episode identifiers: " + ", ".join(found[:8]))
    if f"alertzero-replay:{run_id}" not in blob:
        die("replay tag was not applied")


class Client:
    def __init__(self, es_url: str, kibana_url: str | None, space: str, authorization: str):
        self.es_url = es_url.rstrip("/")
        self.kibana_url = kibana_url.rstrip("/") if kibana_url else None
        self.space = space
        self.authorization = authorization

    def request(self, url: str, method: str, body: bytes | None = None, content_type: str | None = None, extra: dict | None = None) -> tuple[int, bytes]:
        headers = {"Authorization": self.authorization, "Accept": "application/json"}
        if content_type:
            headers["Content-Type"] = content_type
        if extra:
            headers.update(extra)
        req = urllib.request.Request(url, data=body, headers=headers, method=method)
        try:
            with urllib.request.urlopen(req, timeout=120) as resp:
                return resp.status, resp.read()
        except urllib.error.HTTPError as exc:
            return exc.code, exc.read()

    def es(self, path: str, method: str, body: bytes | None = None, content_type: str | None = None) -> tuple[int, bytes]:
        return self.request(f"{self.es_url}{path}", method, body, content_type)

    def kibana(self, path: str, method: str, body: bytes | None = None, content_type: str | None = None) -> tuple[int, bytes]:
        if not self.kibana_url:
            die("Kibana URL is required for this step. Pass --kibana-url or KIBANA_URL.")
        prefix = f"/s/{self.space}" if self.space else ""
        return self.request(
            f"{self.kibana_url}{prefix}{path}",
            method,
            body,
            content_type,
            {"kbn-xsrf": "true", "elastic-api-version": "2023-10-31"},
        )


def ensure_pipeline(client: Client) -> None:
    status, _ = client.es(f"/_ingest/pipeline/{PIPELINE}", "GET")
    if status == 200:
        print(f"pipeline exists: {PIPELINE}")
        return
    payload = json.dumps({
        "description": "Stamp replayed endpoint alerts with the ingest time so the insights detection rule window matches.",
        "processors": [
            {"set": {"field": "event.ingested", "value": "{{_ingest.timestamp}}"}},
            {"set": {"field": "@timestamp", "value": "{{_ingest.timestamp}}"}},
        ],
    }).encode()
    status, raw = client.es(f"/_ingest/pipeline/{PIPELINE}", "PUT", payload, "application/json")
    if status not in (200, 201):
        die(f"create pipeline failed: HTTP {status} {raw[:400]!r}")
    print(f"created pipeline: {PIPELINE}")


def ensure_alert_index(client: Client, index: str) -> None:
    status, raw = client.es(f"/{index}", "HEAD")
    if status == 200:
        print(f"index exists: {index}")
        return
    if status != 404:
        die(f"HEAD {index} failed: HTTP {status} {raw[:300]!r}")
    payload = json.dumps({
        "settings": {"index.mapping.total_fields.limit": "6000"},
        "mappings": json.loads(MAPPING.read_text()),
    }).encode()
    status, raw = client.es(f"/{index}", "PUT", payload, "application/json")
    if status not in (200, 201):
        die(f"create {index} failed: HTTP {status} {raw[:500]!r}")
    print(f"created index: {index}")


def bulk(client: Client, index: str, docs: list[dict]) -> None:
    if not docs:
        print(f"skip empty {index}")
        return
    status, raw = client.es(f"/_bulk?pipeline={PIPELINE}", "POST", bulk_body(index, docs), "application/x-ndjson")
    if status != 200:
        die(f"bulk {index} failed: HTTP {status} {raw[:500]!r}")
    body = json.loads(raw)
    if body.get("errors"):
        failed = [item for item in body.get("items", []) if next(iter(item.values())).get("error")]
        first = failed[0] if failed else {}
        die(f"bulk {index}: {len(failed)} item errors, first={json.dumps(first)[:500]}")
    print(f"uploaded {len(docs):4d}  ->  {index}  (took {body.get('took')} ms)")


def find_rule(client: Client) -> dict | None:
    query = urllib.parse.urlencode({
        "filter": f'alert.attributes.params.ruleId:"{RULE_ID}"',
        "per_page": "1",
    })
    status, raw = client.kibana(f"/api/detection_engine/rules/_find?{query}", "GET")
    if status != 200:
        die(f"rule lookup failed: HTTP {status} {raw[:400]!r}")
    data = json.loads(raw).get("data") or []
    return data[0] if data else None


def import_rule(client: Client) -> None:
    boundary = "----alertzero-replay"
    rule_bytes = RULE_FILE.read_bytes()
    body = (
        f"--{boundary}\r\n"
        'Content-Disposition: form-data; name="file"; filename="detection-rule.ndjson"\r\n'
        "Content-Type: application/octet-stream\r\n\r\n"
    ).encode() + rule_bytes + f"\r\n--{boundary}--\r\n".encode()
    status, raw = client.kibana(
        "/api/detection_engine/rules/_import?overwrite=true",
        "POST",
        body,
        f"multipart/form-data; boundary={boundary}",
    )
    if status not in (200, 201):
        die(f"rule import failed: HTTP {status} {raw[:500]!r}")
    result = json.loads(raw)
    if not result.get("success"):
        die(f"rule import rejected the file: {raw[:500]!r}")
    print(f"imported detection rule {RULE_ID}")


def enable_rule(client: Client, rule: dict) -> None:
    status, raw = client.kibana(
        "/api/detection_engine/rules/_bulk_action",
        "POST",
        json.dumps({"action": "enable", "ids": [rule["id"]]}).encode(),
        "application/json",
    )
    if status != 200:
        die(f"enable rule failed: HTTP {status} {raw[:400]!r}")
    print("enabled detection rule Endpoint Security [Insights]")


def install_rule(client: Client) -> None:
    existing = find_rule(client)
    if existing and existing.get("enabled"):
        print("detection rule already enabled: Endpoint Security [Insights]")
        return
    if existing is None:
        import_rule(client)
        existing = find_rule(client)
        if existing is None:
            die("rule import returned success but the rule id was not found")
    if not existing.get("enabled"):
        enable_rule(client, existing)


def count(client: Client, index: str, tag: str) -> int:
    status, raw = client.es(
        f"/{index}/_count",
        "POST",
        json.dumps({"query": {"term": {"tags": tag}}}).encode(),
        "application/json",
    )
    if status != 200:
        die(f"count {index} failed: HTTP {status} {raw[:300]!r}")
    return int(json.loads(raw).get("count", 0))


def verify(client: Client, batches: list[tuple[str, int]], alert_index: str, tag: str) -> None:
    targets = ",".join([index for index, _ in batches] + [alert_index])
    status, raw = client.es(f"/{targets}/_refresh", "POST")
    if status != 200:
        print(f"refresh returned HTTP {status}; counts may be short until the next refresh")
    for index, expected in batches:
        got = count(client, index, tag)
        mark = "ok" if got == expected else "MISMATCH"
        print(f"verify {mark:8s} {got:4d}/{expected:<4d}  {index}")
        if got != expected:
            die("uploaded counts do not match the replay. See the lines above.")
    got = count(client, alert_index, tag)
    expected = 4
    mark = "ok" if got == expected else "MISMATCH"
    print(f"verify {mark:8s} {got:4d}/{expected:<4d}  {alert_index}")
    if got != expected:
        die("alert count does not match the replay")


def print_signal_check(alert_index: str, host_name: str, agent_id: str, rule_name: str) -> None:
    print("Detection signals can take up to 5 minutes. Then run this in Kibana Dev Tools:")
    print("GET .alerts-security.alerts-default/_search")
    print(json.dumps({
        "size": 10,
        "_source": ["host.name", "agent.id", "rule.name", "kibana.alert.rule.name"],
        "query": {"term": {"kibana.alert.ancestors.index": alert_index}},
    }, indent=2))
    print(f"Expect 4 hits, two rule names. host.name={host_name} agent.id={agent_id} rule.name={rule_name}")


def deployment_notes(deployment: str) -> None:
    if deployment == "serverless":
        print("serverless: Fleet Server is hosted. Use the Fleet URL from the Add agent flyout. Do not deploy Fleet Server.")
    else:
        print("stack: enroll the VM with the Fleet Server URL from this cluster's Add agent flyout. Elasticsearch bulk is the same API as serverless.")
    print("Copy host.name, Agent ID, and Host ID from Fleet -> Agents. host.name is the FQDN, not the short shell hostname.")
    print("The account needs a project Admin role. An Editor can open Fleet and still cannot isolate.")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--deployment", choices=("serverless", "stack"), default=os.environ.get("REPLAY_DEPLOYMENT", "serverless"))
    parser.add_argument("--es-url", default=os.environ.get("ES_URL", ""))
    parser.add_argument("--kibana-url", default=os.environ.get("KIBANA_URL", ""))
    parser.add_argument("--space", default=os.environ.get("KIBANA_SPACE", "default"))
    parser.add_argument("--api-key", default=os.environ.get("ES_API_KEY", ""), help="Encoded API key. Used as Authorization: ApiKey.")
    parser.add_argument("--user", default=os.environ.get("ES_USER", ""), help="Elasticsearch username. Used with --password.")
    parser.add_argument("--password", default=os.environ.get("ES_PASSWORD", ""), help="Prefer ES_PASSWORD over putting this on the command line.")
    parser.add_argument("--host-name", default=os.environ.get("REPLAY_HOST_NAME", ""))
    parser.add_argument("--agent-id", default=os.environ.get("REPLAY_AGENT_ID", ""))
    parser.add_argument("--host-id", default=os.environ.get("REPLAY_HOST_ID", ""))
    parser.add_argument("--agent-name", default=os.environ.get("REPLAY_AGENT_NAME", ""))
    parser.add_argument("--run-id", default=os.environ.get("REPLAY_RUN_ID", ""), help="Short id baked into names and the alert index. Example: demo1")
    parser.add_argument("--upload", action="store_true")
    parser.add_argument("--install-rule", action="store_true", help="Import and enable the insights detection rule in Kibana.")
    parser.add_argument("--verify", action="store_true", help="After upload, count the documents by the replay tag.")
    return parser.parse_args()


def authorization(args: argparse.Namespace) -> str:
    api_key = args.api_key.strip()
    user = args.user.strip()
    # An API key wins over ES_USER left in the shell. Both count only when --user is on the command line.
    if api_key and user and "--user" in sys.argv:
        die("Pass either an API key or a username and password, not both.")
    if api_key:
        return f"ApiKey {api_key}"
    if not user:
        die("Set ES_API_KEY, or set ES_USER and ES_PASSWORD.")
    password = args.password
    if not password:
        password = getpass.getpass("Elasticsearch password: ")
    if not password:
        die("Password is empty.")
    token = base64.b64encode(f"{user}:{password}".encode()).decode()
    return f"Basic {token}"


def main() -> None:
    args = parse_args()
    for label, value in (("--host-name", args.host_name), ("--agent-id", args.agent_id), ("--host-id", args.host_id), ("--run-id", args.run_id)):
        if not value.strip():
            die(f"{label} is required")
    run_id = args.run_id.strip()
    if not run_id.replace("-", "").isalnum() or not run_id[0].isalnum() or len(run_id) > 24:
        die("--run-id must be 1-24 letters, numbers, or hyphens")
    host_name = args.host_name.strip()
    agent_id = args.agent_id.strip()
    host_id = args.host_id.strip()
    agent_name = args.agent_name.strip() or host_name.split(".")[0]
    tag = f"alertzero-replay:{run_id}"
    new_payload = hashlib.sha256(f"{run_id}:payload".encode()).hexdigest()
    new_binary = f"az{run_id.replace('-', '')}flush"
    new_archive = f"{run_id}-stage.zip"
    new_rule = f"Linux.Trojan.Replay.{run_id}"
    day = datetime.now(timezone.utc).strftime("%Y.%m.%d")
    alert_index = f"insights-alerts-replay-{run_id}-{day}"
    subs = [
        (OLD_PAYLOAD, new_payload),
        (OLD_ARCHIVE, new_archive),
        (OLD_BINARY, new_binary),
        (OLD_RULE, new_rule),
        (OLD_AGENT, agent_id),
        (OLD_HOST_ID, host_id),
        (OLD_HOST, host_name),
    ]
    hash_map = {
        old: hashlib.sha256(f"{run_id}:{old}".encode()).hexdigest()[: len(old)]
        for old in SHARED_HASHES
    }
    events = load_docs(SRC_EVENTS, subs, hash_map, host_name, agent_id, host_id, agent_name, tag)
    alerts = load_docs(SRC_ALERTS, subs, hash_map, host_name, agent_id, host_id, agent_name, tag)
    drop_rule, exec_rule = split_alert_rules(alerts, new_rule)
    assert_clean(events + alerts, run_id)

    batches: list[tuple[str, list[dict]]] = []
    skipped = 0
    for kind, index in DATASETS.items():
        docs = [doc for doc in events if dataset_of(doc) == kind]
        batches.append((index, docs))
    skipped = sum(1 for doc in events if dataset_of(doc) not in DATASETS)

    print(f"deployment {args.deployment}")
    print(f"host {OLD_HOST} -> {host_name}")
    print(f"agent {OLD_AGENT} -> {agent_id}")
    print(f"binary {OLD_BINARY} -> {new_binary}")
    print(f"archive {OLD_ARCHIVE} -> {new_archive}")
    print(f"rule {OLD_RULE} -> {drop_rule} (unzip) and {exec_rule} (/dev/shm)")
    drop_count = sum(1 for alert in alerts if (alert.get("rule") or {}).get("name") == drop_rule)
    exec_count = sum(1 for alert in alerts if (alert.get("rule") or {}).get("name") == exec_rule)
    print(f"alert rules {drop_rule}={drop_count}  {exec_rule}={exec_count}")
    print(f"tag {tag}")
    print(f"events {len(events)}  alerts {len(alerts)}  skipped syslog {skipped}")
    for index, docs in batches:
        print(f"ready {len(docs):4d}  ->  {index}")
    print(f"ready {len(alerts):4d}  ->  {alert_index}")
    deployment_notes(args.deployment)

    if not args.upload and not args.install_rule:
        print("files were not uploaded. Re-run with --upload.")
        print_signal_check(alert_index, host_name, agent_id, f"{drop_rule} and {exec_rule}")
        return

    if not args.es_url and args.upload:
        die("--es-url or ES_URL is required to upload")
    client = Client(args.es_url or "http://unused", args.kibana_url or None, args.space, authorization(args))
    if args.install_rule:
        install_rule(client)
    if not args.upload:
        return
    client = Client(args.es_url, args.kibana_url or None, args.space, client.authorization)
    ensure_pipeline(client)
    ensure_alert_index(client, alert_index)
    for index, docs in batches:
        bulk(client, index, docs)
    bulk(client, alert_index, alerts)
    if args.verify:
        verify(client, [(index, len(docs)) for index, docs in batches], alert_index, tag)
    print_signal_check(alert_index, host_name, agent_id, f"{drop_rule} and {exec_rule}")


if __name__ == "__main__":
    main()
