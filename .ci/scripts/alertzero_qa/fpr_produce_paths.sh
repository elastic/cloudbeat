#!/usr/bin/env bash
#
# AlertZero QA data seeder: produces open false-positive detection alerts
# (optionally auto-closed via the alert-analysis workflow) for exercising
# the rule-tuning worker. Not product code -- talks to Kibana/ES over HTTP.
#
# fpr_produce_paths.sh — like fpr_produce.sh, but seeds events shaped to push the
# rule-tuning worker toward a specific tuning recommendation.
#
# Usage:
#   ./fpr_produce_paths.sh --path exception      # tight entity pattern → add exception
#   ./fpr_produce_paths.sh --path query          # shared safe field → narrow the query
#   ./fpr_produce_paths.sh --path suppression    # same entity repeated → suppress duplicates
#   ./fpr_produce_paths.sh --path risk_score     # broad spread, low stakes → lower risk score
#   ./fpr_produce_paths.sh --path disable        # 100% FP rule, no detection value → disable
#   ./fpr_produce_paths.sh --path threshold      # threshold count too low → raise threshold value
#   ./fpr_produce_paths.sh --path schedule       # lookback window too wide → tighten schedule
#   ./fpr_produce_paths.sh --path all            # run all 7 paths sequentially
#
# Add --no-workflow to stop after producing open alerts (skip alert-analysis).
#
# Env (same defaults as fpr_produce.sh):
#   ES_URL=http://localhost:9200
#   KIBANA_URL=http://localhost:5651
#   KIBANA_AUTH=elastic:changeme
#   WORKFLOW_CONNECTOR_ID=
#   AUTO_CLOSE_MIN=0.5
set -euo pipefail

ES_URL="${ES_URL:-http://localhost:9200}"
ES_URL="${ES_URL%/}"
KB_URL="${KIBANA_URL:-http://localhost:5601}"
KB_URL="${KB_URL%/}"
AUTH="${KIBANA_AUTH:-elastic:changeme}"
AUTO_CLOSE_MIN="${AUTO_CLOSE_MIN:-0.5}"

RUN_TOKEN="${RUN_TOKEN:-$(date -u +%Y%m%d-%H%M%S)-$RANDOM}"
DEMO_TAG="false-positive-demo"
WANT=5
WANT_ALERTS="${WANT_ALERTS:-$WANT}"
WITH_WORKFLOW=1
PATH_ARG=""

while [[ $# -gt 0 ]]; do
    case "$1" in
    --path=*)
        PATH_ARG="${1#--path=}"
        shift
        ;;
    --path)
        PATH_ARG="${2:-}"
        shift 2
        ;;
    --no-workflow)
        WITH_WORKFLOW=0
        shift
        ;;
    *)
        echo "unknown argument: $1" >&2
        exit 64
        ;;
    esac
done

case "$PATH_ARG" in
exception | query | suppression | risk_score | disable | threshold | schedule) ;;
all)
    SCRIPT="$(realpath "$0")"
    for p in exception query suppression risk_score disable threshold schedule; do
        echo ""
        echo "════════════════ path=$p ════════════════"
        if [ "$WITH_WORKFLOW" -eq 0 ]; then
            "$SCRIPT" --path "$p" --no-workflow
        else
            "$SCRIPT" --path "$p"
        fi
    done
    exit 0
    ;;
"")
    echo "required: --path exception|query|suppression|risk_score|disable|threshold|schedule|all" >&2
    exit 64
    ;;
*)
    echo "unknown path: $PATH_ARG (expected exception|query|suppression|risk_score|disable|threshold|schedule|all)" >&2
    exit 64
    ;;
esac

command -v jq >/dev/null || {
    echo "jq is required" >&2
    exit 1
}

INDEX="fpr-events-$RUN_TOKEN"
RULE_ID="fpr-produce-$RUN_TOKEN"

es() { curl -sS -u "$AUTH" -H 'content-type: application/json' "$@"; }
kbp() { curl -sS -u "$AUTH" -H 'kbn-xsrf: true' -H 'x-elastic-internal-origin: kibana' \
    -H 'elastic-api-version: 2023-10-31' -H 'content-type: application/json' "$@"; }
kbi() { curl -sS -u "$AUTH" -H 'kbn-xsrf: true' -H 'x-elastic-internal-origin: kibana' \
    -H 'elastic-api-version: 1' -H 'content-type: application/json' "$@"; }

AA="$KB_URL/internal/security_solution/alert_analysis_workflow"
NOW="$(date -u +%Y-%m-%dT%H:%M:%S.000Z)"

echo "ES=$ES_URL  KB=$KB_URL  path=$PATH_ARG  run=$RUN_TOKEN"

# ── 1. Pick events + rule shape for each tuning path ───────────────────────────────────────
#
# exception: 5 alerts all involve the same process.name + process.executable on the same host.
#   The AI should see a tight entity cluster and propose an exception on those fields.
#
# query: 5 alerts all share process.command_line matching "*.ps1 -ExecutionPolicy Bypass*",
#   spread across different hosts/users. The AI should see a query-narrowable pattern and
#   suggest tightening the KQL (e.g. add `NOT process.command_line: *RemoteSigned*`).
#   The rule description hints that the intent was only to catch unsigned PowerShell.
#
# suppression: the same host + user + process fires 5 times (duplicate alert bursts). The AI
#   should see high-volume repeated entity and suggest alert suppression rather than an exception.
#
# risk_score: 5 alerts spread across 5 different hosts, 5 different users, no concentration.
#   The rule catches real behavior (mixed code_signature.trusted) but it's all low-stakes
#   sysadmin tooling. The AI should see spread + low severity and propose lowering the risk score.
#
# disable: the rule matches every single process event on the host, including OS noise.
#   All 5 FPs are completely routine OS activity with zero detection value, and the rule
#   description says it was a temporary breadth test that should be retired.

case "$PATH_ARG" in

# ── exception ──────────────────────────────────────────────────────────────────────────────
exception)
    RULE_NAME="Suspicious Admin Tool Execution [FPR-exception $RUN_TOKEN]"
    RULE_DESC="Detects process activity on admin hosts. Broad rule — may fire on known-good backup tooling."
    RULE_QUERY='event.category: "process"'
    RULE_SEVERITY="medium"
    RULE_RISK=47

    read -r -d '' EVENTS <<'JSON' || true
[
 {"message":"Nightly VSS shadow copy — backup service account, same host every night","host":{"name":"WIN-BACKUP-01"},"user":{"name":"svc_backup"},"process":{"name":"vssadmin.exe","executable":"C:\\Windows\\System32\\vssadmin.exe","command_line":"vssadmin.exe create shadow /for=C:","pid":3201,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"VSS shadow copy (second nightly run, same host)","host":{"name":"WIN-BACKUP-01"},"user":{"name":"svc_backup"},"process":{"name":"vssadmin.exe","executable":"C:\\Windows\\System32\\vssadmin.exe","command_line":"vssadmin.exe create shadow /for=D:","pid":3210,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"VSS shadow copy (weekly full backup, same host)","host":{"name":"WIN-BACKUP-01"},"user":{"name":"svc_backup"},"process":{"name":"vssadmin.exe","executable":"C:\\Windows\\System32\\vssadmin.exe","command_line":"vssadmin.exe create shadow /for=E:","pid":3215,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"VSS list shadows (integrity check, same host)","host":{"name":"WIN-BACKUP-01"},"user":{"name":"svc_backup"},"process":{"name":"vssadmin.exe","executable":"C:\\Windows\\System32\\vssadmin.exe","command_line":"vssadmin.exe list shadows","pid":3220,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"VSS shadow copy (monthly offsite backup, same host)","host":{"name":"WIN-BACKUP-01"},"user":{"name":"svc_backup"},"process":{"name":"vssadmin.exe","executable":"C:\\Windows\\System32\\vssadmin.exe","command_line":"vssadmin.exe create shadow /for=F:","pid":3225,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}}
]
JSON
    ;;

# ── query ──────────────────────────────────────────────────────────────────────────────────
query)
    RULE_NAME="Unsigned PowerShell Execution [FPR-query $RUN_TOKEN]"
    RULE_DESC="Detects PowerShell launched with -ExecutionPolicy Bypass. Intent: catch unsigned scripts run ad-hoc by attackers. Known-good paths: IT scripts signed with RemoteSigned or AllSigned policy are out of scope."
    RULE_QUERY='process.name: "powershell.exe" AND process.command_line: *ExecutionPolicy*'
    RULE_SEVERITY="high"
    RULE_RISK=73

    read -r -d '' EVENTS <<'JSON' || true
[
 {"message":"Signed IT asset-inventory script (AllSigned policy)","host":{"name":"WIN-MON-01"},"user":{"name":"svc_monitor"},"process":{"name":"powershell.exe","executable":"C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe","command_line":"powershell.exe -ExecutionPolicy AllSigned -File C:\\Scripts\\Get-AssetInventory.ps1","pid":4001,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"Signed patch deployment script (AllSigned policy)","host":{"name":"WIN-MON-07"},"user":{"name":"svc_patching"},"process":{"name":"powershell.exe","executable":"C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe","command_line":"powershell.exe -ExecutionPolicy AllSigned -File C:\\Scripts\\Deploy-Patches.ps1","pid":4055,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"Signed compliance audit script (AllSigned policy)","host":{"name":"WIN-SEC-03"},"user":{"name":"svc_compliance"},"process":{"name":"powershell.exe","executable":"C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe","command_line":"powershell.exe -ExecutionPolicy AllSigned -File C:\\Scripts\\Run-ComplianceAudit.ps1","pid":4112,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"Signed certificate renewal script (AllSigned policy)","host":{"name":"WIN-CERT-02"},"user":{"name":"svc_pki"},"process":{"name":"powershell.exe","executable":"C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe","command_line":"powershell.exe -ExecutionPolicy AllSigned -File C:\\Scripts\\Renew-Certificates.ps1","pid":4201,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"Signed log rotation script (AllSigned policy)","host":{"name":"WIN-LOG-05"},"user":{"name":"svc_logs"},"process":{"name":"powershell.exe","executable":"C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe","command_line":"powershell.exe -ExecutionPolicy AllSigned -File C:\\Scripts\\Rotate-Logs.ps1","pid":4310,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}}
]
JSON
    ;;

# ── suppression ────────────────────────────────────────────────────────────────────────────
suppression)
    RULE_NAME="Port Scanner Activity Detected [FPR-suppression $RUN_TOKEN]"
    RULE_DESC="Detects processes that appear to perform port scanning. The corporate vulnerability scanner runs nightly from WIN-SCAN-01 — its alerts are known-good but the scanner is legitimately active and should stay running."
    RULE_QUERY='event.category: "process"'
    RULE_SEVERITY="high"
    RULE_RISK=73

    read -r -d '' EVENTS <<'JSON' || true
[
 {"message":"Corporate Nessus scanner — nightly sweep (run 1 of 5)","host":{"name":"WIN-SCAN-01"},"user":{"name":"svc_vulnscan"},"process":{"name":"nessusd.exe","executable":"C:\\Program Files\\Tenable\\Nessus\\nessusd.exe","command_line":"nessusd.exe --target-list corp-targets.txt --policy nightly","pid":5001,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Tenable Network Security"}}},
 {"message":"Corporate Nessus scanner — nightly sweep (run 2 of 5)","host":{"name":"WIN-SCAN-01"},"user":{"name":"svc_vulnscan"},"process":{"name":"nessusd.exe","executable":"C:\\Program Files\\Tenable\\Nessus\\nessusd.exe","command_line":"nessusd.exe --target-list corp-targets.txt --policy nightly","pid":5002,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Tenable Network Security"}}},
 {"message":"Corporate Nessus scanner — nightly sweep (run 3 of 5)","host":{"name":"WIN-SCAN-01"},"user":{"name":"svc_vulnscan"},"process":{"name":"nessusd.exe","executable":"C:\\Program Files\\Tenable\\Nessus\\nessusd.exe","command_line":"nessusd.exe --target-list corp-targets.txt --policy nightly","pid":5003,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Tenable Network Security"}}},
 {"message":"Corporate Nessus scanner — nightly sweep (run 4 of 5)","host":{"name":"WIN-SCAN-01"},"user":{"name":"svc_vulnscan"},"process":{"name":"nessusd.exe","executable":"C:\\Program Files\\Tenable\\Nessus\\nessusd.exe","command_line":"nessusd.exe --target-list corp-targets.txt --policy nightly","pid":5004,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Tenable Network Security"}}},
 {"message":"Corporate Nessus scanner — nightly sweep (run 5 of 5)","host":{"name":"WIN-SCAN-01"},"user":{"name":"svc_vulnscan"},"process":{"name":"nessusd.exe","executable":"C:\\Program Files\\Tenable\\Nessus\\nessusd.exe","command_line":"nessusd.exe --target-list corp-targets.txt --policy nightly","pid":5005,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Tenable Network Security"}}}
]
JSON
    ;;

# ── risk_score ─────────────────────────────────────────────────────────────────────────────
risk_score)
    RULE_NAME="Admin Tool Execution — High Severity [FPR-risk_score $RUN_TOKEN]"
    RULE_DESC="Detects admin/management tools on workstations. Currently severity=high / risk_score=73, but routine IT management generates many low-stakes alerts across unrelated endpoints."
    RULE_QUERY='event.category: "process"'
    RULE_SEVERITY="high"
    RULE_RISK=73

    read -r -d '' EVENTS <<'JSON' || true
[
 {"message":"IT helpdesk ran gpupdate to push a policy change (ticket HD-8812)","host":{"name":"WIN-DESK-04"},"user":{"name":"helpdesk_user"},"process":{"name":"gpupdate.exe","executable":"C:\\Windows\\System32\\gpupdate.exe","command_line":"gpupdate.exe /force","pid":6001,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"IT ran diskpart to extend a volume after resize (ticket HD-8813)","host":{"name":"WIN-DESK-11"},"user":{"name":"it_admin_2"},"process":{"name":"diskpart.exe","executable":"C:\\Windows\\System32\\diskpart.exe","command_line":"diskpart.exe","pid":6055,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"SysInternals PsExec used to restart a service remotely (approved)","host":{"name":"WIN-SRV-09"},"user":{"name":"it_admin_3"},"process":{"name":"PsExec.exe","executable":"C:\\Tools\\SysInternals\\PsExec.exe","command_line":"PsExec.exe \\\\print-server restart spooler","pid":6112,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Sysinternals"}}},
 {"message":"IT used net.exe to add a workstation to a maintenance group","host":{"name":"WIN-DESK-22"},"user":{"name":"it_admin_4"},"process":{"name":"net.exe","executable":"C:\\Windows\\System32\\net.exe","command_line":"net.exe localgroup Maintenance WIN-DESK-22 /add","pid":6201,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"Automated certificate check via certutil — scheduled task","host":{"name":"WIN-SRV-15"},"user":{"name":"system"},"process":{"name":"certutil.exe","executable":"C:\\Windows\\System32\\certutil.exe","command_line":"certutil.exe -verify corp-root-ca.crt","pid":6310,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}}
]
JSON
    ;;

# ── disable ────────────────────────────────────────────────────────────────────────────────
disable)
    RULE_NAME="Catch-All Process Breadth Test [FPR-disable $RUN_TOKEN]"
    RULE_DESC="Temporary breadth-test rule created during a tuning sprint to measure alert volume. Every process event on every host fires this rule — it was never scoped. This rule was supposed to be deleted before it went to production and has zero detection value. It is fully superseded by targeted rules in the detection library. Should be disabled and removed."
    RULE_QUERY='event.category: "process"'
    RULE_SEVERITY="low"
    RULE_RISK=21

    read -r -d '' EVENTS <<'JSON' || true
[
 {"message":"Windows system idle process heartbeat (OS noise, fires every minute)","host":{"name":"WIN-DESK-01"},"user":{"name":"system"},"process":{"name":"System","executable":"System","command_line":"","pid":4,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"Windows session manager (normal OS startup sequence)","host":{"name":"WIN-DESK-02"},"user":{"name":"system"},"process":{"name":"smss.exe","executable":"C:\\Windows\\System32\\smss.exe","command_line":"\\SystemRoot\\System32\\smss.exe","pid":312,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"Windows service host (generic SvcHost for network services)","host":{"name":"WIN-SRV-03"},"user":{"name":"network_service"},"process":{"name":"svchost.exe","executable":"C:\\Windows\\System32\\svchost.exe","command_line":"C:\\Windows\\System32\\svchost.exe -k NetworkService","pid":876,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"Windows runtime broker for UWP apps (routine)","host":{"name":"WIN-DESK-04"},"user":{"name":"a.user"},"process":{"name":"RuntimeBroker.exe","executable":"C:\\Windows\\System32\\RuntimeBroker.exe","command_line":"C:\\Windows\\System32\\RuntimeBroker.exe -Embedding","pid":1024,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"Windows security health service (Defender background task)","host":{"name":"WIN-DESK-05"},"user":{"name":"system"},"process":{"name":"SecurityHealthService.exe","executable":"C:\\Windows\\System32\\SecurityHealthService.exe","command_line":"C:\\Windows\\System32\\SecurityHealthService.exe","pid":1248,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}}
]
JSON
    ;;

# ── threshold ──────────────────────────────────────────────────────────────────────────────
threshold)
    RULE_NAME="Repeated Authentication Failures — Low Threshold [FPR-threshold $RUN_TOKEN]"
    RULE_DESC="Detects hosts with repeated authentication failures using a threshold rule. Current threshold is 2 events — far too low for this environment where service accounts trigger 3-5 auth failures per minute during routine certificate renewals."
    RULE_QUERY='event.category: "process" AND process.name: "certreq.exe"'
    RULE_SEVERITY="medium"
    RULE_RISK=47
    RULE_TYPE="threshold"
    RULE_EXTRA='{"threshold":{"field":["host.name"],"value":2}}'
    WANT_ALERTS=2 # threshold rules fire 1 alert per group; 2 hosts → 2 alerts

    read -r -d '' EVENTS <<'JSON' || true
[
 {"message":"Certificate renewal auth failure (1 of 3) — WIN-CERT-01","host":{"name":"WIN-CERT-01"},"user":{"name":"svc_pki"},"event":{"category":"authentication","outcome":"failure","type":"start","action":"logon","dataset":"endpoint.process"},"process":{"name":"certreq.exe","executable":"C:\\Windows\\System32\\certreq.exe","command_line":"certreq.exe -submit renewal.req","pid":7001,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"Certificate renewal auth failure (2 of 3) — WIN-CERT-01","host":{"name":"WIN-CERT-01"},"user":{"name":"svc_pki"},"event":{"category":"authentication","outcome":"failure","type":"start","action":"logon","dataset":"endpoint.process"},"process":{"name":"certreq.exe","executable":"C:\\Windows\\System32\\certreq.exe","command_line":"certreq.exe -submit renewal2.req","pid":7002,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"Certificate renewal auth failure (3 of 3) — WIN-CERT-01","host":{"name":"WIN-CERT-01"},"user":{"name":"svc_pki"},"event":{"category":"authentication","outcome":"failure","type":"start","action":"logon","dataset":"endpoint.process"},"process":{"name":"certreq.exe","executable":"C:\\Windows\\System32\\certreq.exe","command_line":"certreq.exe -submit renewal3.req","pid":7003,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"Certificate renewal auth failure (1 of 2) — WIN-CERT-02","host":{"name":"WIN-CERT-02"},"user":{"name":"svc_pki"},"event":{"category":"authentication","outcome":"failure","type":"start","action":"logon","dataset":"endpoint.process"},"process":{"name":"certreq.exe","executable":"C:\\Windows\\System32\\certreq.exe","command_line":"certreq.exe -submit renewal.req","pid":8001,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}},
 {"message":"Certificate renewal auth failure (2 of 2) — WIN-CERT-02","host":{"name":"WIN-CERT-02"},"user":{"name":"svc_pki"},"event":{"category":"authentication","outcome":"failure","type":"start","action":"logon","dataset":"endpoint.process"},"process":{"name":"certreq.exe","executable":"C:\\Windows\\System32\\certreq.exe","command_line":"certreq.exe -submit renewal2.req","pid":8002,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Windows"}}}
]
JSON
    ;;

# ── schedule ───────────────────────────────────────────────────────────────────────────────
schedule)
    RULE_NAME="Outbound Connection to Rare External IP [FPR-schedule $RUN_TOKEN]"
    RULE_DESC="Detects outbound connections to external IPs that are rare in the environment. The rule runs every 5 minutes with a 24h lookback window — far too wide. This causes every connection seen at any point in the last day to fire repeatedly, generating FPs from known-good CDN and cloud provider IPs that are common but only rare on this specific query's window."
    RULE_QUERY='event.category: "process"'
    RULE_SEVERITY="medium"
    RULE_RISK=47
    RULE_TYPE="query"
    RULE_EXTRA=""
    WANT_ALERTS=5 # all 5 events match; diverse processes/hosts make exception/query inappropriate
    # Override the default 1h window with a very wide one to force schedule recommendation
    RULE_FROM="now-24h"

    read -r -d '' EVENTS <<'JSON' || true
[
 {"message":"Outbound HTTPS to Cloudflare CDN — routine SaaS traffic","host":{"name":"WIN-DESK-10"},"user":{"name":"a.user"},"event":{"category":"network","type":"start","action":"connection_attempted","outcome":"success","dataset":"endpoint.process"},"network":{"direction":"egress","transport":"tcp","protocol":"https"},"destination":{"ip":"104.21.45.1","port":443},"process":{"name":"chrome.exe","executable":"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe","command_line":"chrome.exe https://app.slack.com","pid":8001,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Google LLC"}}},
 {"message":"Outbound HTTPS to AWS S3 — artifact download from CI pipeline","host":{"name":"WIN-DEV-03"},"user":{"name":"svc_ci"},"event":{"category":"network","type":"start","action":"connection_attempted","outcome":"success","dataset":"endpoint.process"},"network":{"direction":"egress","transport":"tcp","protocol":"https"},"destination":{"ip":"52.217.32.1","port":443},"process":{"name":"node.exe","executable":"C:\\Program Files\\nodejs\\node.exe","command_line":"node scripts/build.js","pid":8055,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Node.js Foundation"}}},
 {"message":"Outbound HTTPS to Azure — Teams client sync","host":{"name":"WIN-DESK-07"},"user":{"name":"b.user"},"event":{"category":"network","type":"start","action":"connection_attempted","outcome":"success","dataset":"endpoint.process"},"network":{"direction":"egress","transport":"tcp","protocol":"https"},"destination":{"ip":"13.107.42.1","port":443},"process":{"name":"Teams.exe","executable":"C:\\Users\\b.user\\AppData\\Local\\Microsoft\\Teams\\current\\Teams.exe","command_line":"Teams.exe","pid":8112,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Microsoft Corporation"}}},
 {"message":"Outbound HTTPS to GCP — Google Workspace sync","host":{"name":"WIN-DESK-15"},"user":{"name":"c.user"},"event":{"category":"network","type":"start","action":"connection_attempted","outcome":"success","dataset":"endpoint.process"},"network":{"direction":"egress","transport":"tcp","protocol":"https"},"destination":{"ip":"142.250.80.1","port":443},"process":{"name":"GoogleDriveFSHelper.exe","executable":"C:\\Program Files\\Google\\Drive File Stream\\GoogleDriveFSHelper.exe","command_line":"GoogleDriveFSHelper.exe","pid":8201,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Google LLC"}}},
 {"message":"Outbound HTTPS to Okta — SSO token refresh","host":{"name":"WIN-DESK-22"},"user":{"name":"d.user"},"event":{"category":"network","type":"start","action":"connection_attempted","outcome":"success","dataset":"endpoint.process"},"network":{"direction":"egress","transport":"tcp","protocol":"https"},"destination":{"ip":"34.100.12.1","port":443},"process":{"name":"chrome.exe","executable":"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe","command_line":"chrome.exe https://corp.okta.com","pid":8310,"code_signature":{"status":"trusted","trusted":true,"subject_name":"Google LLC"}}}
]
JSON
    ;;
esac

# Defaults for path-specific overrides
RULE_TYPE="${RULE_TYPE:-query}"
RULE_EXTRA="${RULE_EXTRA:-}"
RULE_FROM="${RULE_FROM:-now-1h}"
# ── 1. Index ───────────────────────────────────────────────────────────────────────────────
es -X PUT "$ES_URL/$INDEX" -d '{
  "mappings":{"properties":{
    "@timestamp":{"type":"date"},
    "disposition":{"type":"keyword"},
    "message":{"type":"text"},
    "event":{"properties":{"category":{"type":"keyword"},"type":{"type":"keyword"},"action":{"type":"keyword"},"outcome":{"type":"keyword"},"dataset":{"type":"keyword"}}},
    "host":{"properties":{"name":{"type":"keyword"},"os":{"properties":{"family":{"type":"keyword"}}}}},
    "user":{"properties":{"name":{"type":"keyword"}}},
    "process":{"properties":{
      "name":{"type":"keyword"},"executable":{"type":"keyword"},"command_line":{"type":"wildcard"},"pid":{"type":"long"},
      "code_signature":{"properties":{"status":{"type":"keyword"},"trusted":{"type":"boolean"},"subject_name":{"type":"keyword"}}}
    }}
  }}
}' | jq -e '.acknowledged==true' >/dev/null || {
    echo "index create failed" >&2
    exit 1
}

BULK="$(jq -rc --arg ts "$NOW" '
  .[] | ({index:{}}), (. + {"@timestamp":$ts,"disposition":"benign",
    "event":{"category":"process","type":"start","action":"start","outcome":"success","dataset":"endpoint.process"}})
' <<<"$EVENTS")"
es -X POST "$ES_URL/$INDEX/_bulk?refresh=wait_for" --data-binary "$BULK"$'\n' |
    jq -e '.errors==false' >/dev/null || {
    echo "bulk index failed" >&2
    exit 1
}
echo "[data] indexed 5 events into $INDEX"

# ── 2. Rule ───────────────────────────────────────────────────────────────────────────────
RULE_BASE="$(jq -n \
    --arg rid "$RULE_ID" --arg name "$RULE_NAME" \
    --arg desc "$RULE_DESC" --arg type "$RULE_TYPE" \
    --arg query "$RULE_QUERY" --arg index "$INDEX*" \
    --arg from "$RULE_FROM" \
    --arg sev "$RULE_SEVERITY" --argjson risk "$RULE_RISK" \
    --arg demo "$DEMO_TAG" --arg path "$PATH_ARG" \
    '{rule_id:$rid,name:$name,description:$desc,type:$type,language:"kuery",
    query:$query,index:[$index],enabled:false,from:$from,to:"now",interval:"1m",
    severity:$sev,risk_score:$risk,max_signals:100,
    tags:["FPR",$demo,("tuning-path:"+$path)]}')"
RULE_EXTRA_OBJ="${RULE_EXTRA}"
[ -z "$RULE_EXTRA_OBJ" ] && RULE_EXTRA_OBJ="{}"
RID="$(echo "$RULE_BASE" | jq -s --argjson extra "${RULE_EXTRA_OBJ}" '.[0] * $extra' |
    kbp -X POST "$KB_URL/api/detection_engine/rules" -H 'content-type: application/json' -d @- |
    jq -r '.id // empty')"
[ -n "$RID" ] || {
    echo "rule create failed" >&2
    exit 1
}
echo "[rule] created \"$RULE_NAME\" id=$RID"

# ── 3. Attach workflow ─────────────────────────────────────────────────────────────────────
if [ "$WITH_WORKFLOW" -eq 1 ]; then
    SETTINGS="$(kbi "$AA/settings")"
    jq -e '.settings' >/dev/null <<<"$SETTINGS" || {
        echo "[workflow] cannot read settings (Enterprise license + privileges required):" >&2
        jq -c . <<<"$SETTINGS" >&2
        exit 1
    }

    CONNECTOR="${WORKFLOW_CONNECTOR_ID:-$(jq -r '.settings.connectorId // empty' <<<"$SETTINGS")}"
    if [ -z "$CONNECTOR" ]; then
        CONNECTOR="$(kbp "$KB_URL/api/actions/connectors" |
            jq -r '[.[] | select(.connector_type_id | test("^\\.(inference|gen-ai|bedrock|gemini)$"))]
               | sort_by(.connector_type_id != ".inference") | .[0].id // empty')"
    fi
    [ -n "$CONNECTOR" ] || {
        echo "[workflow] no LLM connector found; set WORKFLOW_CONNECTOR_ID or create one" >&2
        exit 1
    }

    UPDATED="$(jq -c --arg c "$CONNECTOR" --argjson min "$AUTO_CLOSE_MIN" '.settings
    | { workflowEnabled: true,
        connectorId: $c,
        agentId: .agentId,
        createConversation: .createConversation,
        tagPrefix: .tagPrefix,
        autoCloseEnabled: true,
        autoCloseConfidenceScoreMinThreshold: $min,
        autoCloseConfidenceScoreMaxThreshold: 1 }' <<<"$SETTINGS")"
    kbi -X PUT "$AA/settings" -d "$UPDATED" | jq -e '.settings.workflowEnabled==true' >/dev/null ||
        {
            echo "[workflow] settings update failed" >&2
            exit 1
        }
    echo "[workflow] enabled connector=$CONNECTOR auto-close in [$AUTO_CLOSE_MIN, 1]"

    kbi -X POST "$AA/rules/_update" -d "{\"attachRuleIds\":[\"$RID\"]}" |
        jq -e '.updated >= 1' >/dev/null || {
        echo "[workflow] attach failed" >&2
        exit 1
    }
    echo "[workflow] attached alert-analysis to rule $RID"
else
    echo "[workflow] skipped (--no-workflow)"
fi

# ── 4. Enable + poll ──────────────────────────────────────────────────────────────────────
kbp -X PATCH "$KB_URL/api/detection_engine/rules" -d "{\"rule_id\":\"$RULE_ID\",\"enabled\":true}" |
    jq -e '.enabled==true' >/dev/null || {
    echo "enable failed" >&2
    exit 1
}
echo "[rule] enabled (runs every 1m)"

if [ "$WITH_WORKFLOW" -eq 0 ]; then
    FROM="$(date -u +%Y-%m-%dT%H:%M:%S.000Z)"
    kbp -X POST "$KB_URL/api/detection_engine/rules/_bulk_action" \
        -d "{\"action\":\"run\",\"ids\":[\"$RID\"],\"run\":{\"start_date\":\"$NOW\",\"end_date\":\"$FROM\"}}" -o /dev/null || true
fi

count_alerts() {
    es -X POST "$ES_URL/.alerts-security.alerts-*/_search" \
        -d "{\"size\":0,\"query\":{\"bool\":{\"filter\":[
          {\"term\":{\"kibana.alert.rule.uuid\":\"$RID\"}}${1:+,$1}]}}}" |
        jq -r '.hits.total.value // 0'
}
count_closed_fp() {
    count_alerts '{"term":{"kibana.alert.workflow_status":"closed"}},
                {"term":{"kibana.alert.workflow_reason":"false_positive"}}'
}

WANT_ALERTS="${WANT_ALERTS:-$WANT}"
echo "[verify] polling for >=$WANT_ALERTS alerts..."
N=0
for i in $(seq 1 18); do
    sleep 10
    N="$(count_alerts)"
    echo "  t=$((i * 10))s alerts=$N"
    [ "$N" -ge "$WANT_ALERTS" ] && break
done

if [ "$N" -lt "$WANT_ALERTS" ]; then
    echo "[WARN] only $N alerts after 3 min" >&2
    exit 2
fi
echo "[OK] $N alerts created for path=$PATH_ARG"

[ "$WITH_WORKFLOW" -eq 1 ] || exit 0

echo "[verify] polling for alert-analysis closures (up to ~5 min)..."
for i in $(seq 1 30); do
    sleep 10
    FP="$(count_closed_fp)"
    echo "  t=$((i * 10))s closed_fp=$FP/$N"
    if [ "$FP" -ge "$WANT_ALERTS" ]; then
        echo "[OK] $FP alerts closed as false_positive — ready for tuning worker (path=$PATH_ARG)"
        exit 0
    fi
done

echo "[WARN] only $(count_closed_fp) of $N alerts closed. Check connector health and AUTO_CLOSE_MIN=$AUTO_CLOSE_MIN" >&2
exit 2
