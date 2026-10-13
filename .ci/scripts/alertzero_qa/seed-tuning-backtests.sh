#!/bin/bash
# AlertZero QA data seeder: seed detection rules + false-positive alerts for
# testing every backtest in the rule tuning review workflow. Run the tuning
# worker/review from the UI afterwards.
# Usage: ./seed-tuning-backtests.sh [command] [scenario]
#
# Commands (each takes the scenario as second arg, default exception):
#   seed      — create a rule and 12 closed-as-false-positive alerts for it (default)
#   status    — show the seeded rule and its alert counts
#   clean     — delete the seeded rule + alerts + events
# Scenarios can be seeded side by side; one worker run then reviews each rule.
#
# Scenarios (second arg to `seed`, default exception). The diagnose step prefers
# exception > query > risk_score > threshold, so each scenario shapes the false
# positives to make the earlier options unreasonable:
#   exception   — query rule; 9 of 12 FPs share host.name="admin-workstation-04", 3 don't.
#                 20 source events (15 on that host) so the backtest reads ~"20 now, 5 after".
#   risk_score  — intentionally broad "successful logon audit" rule (its description says
#                 narrowing would defeat its purpose); 12 FPs on 12 different hosts and
#                 users spread over 6 days, so there is nothing to exclude or narrow and the
#                 rule just reads as over-scored (high/73). The proposal should lower risk
#                 score/severity.
#   threshold   — threshold rule on host.name with value 3; 12 FPs on 12 different hosts
#                 each barely over the line (count 3-5). Events: 15 on admin-workstation-04
#                 and 4 each on three other hosts, so the preview reads "4 now" and fewer
#                 with a higher threshold.
#   threshold_query
#               — the same threshold rule, but the 12 FPs are on 12 different hosts with
#                 counts well over the line (10-12), all from *_updater.exe processes, so
#                 neither an exception nor a higher threshold helps and the proposal should
#                 narrow the query. Events: 4 each on 4 hosts, 3 hosts all updaters, so the
#                 preview reads "~3 now, 1 after".
#   schedule    — query rule that looks back 65m on a 5m interval, so every run re-alerts
#                 on the same events. 12 FPs are 3 source events x 4 duplicate alerts
#                 (same original_time, alert timestamps 5m apart). 20 events spread over
#                 the hour before TS, so the preview reads "many now, ~3 with a 6m lookback".
#   data_view, timestamp_override, alert_suppression
#               — the exception scenario (same FPs and events, expect ~"20 now, 5 after")
#                 on a rule that uses a data view instead of index patterns, a timestamp
#                 override (event.ingested), or alert suppression on host.name. Before the
#                 widening PR these three got "No backtest is available".
#   eql         — the exception scenario on an EQL rule (process where process.name != null).
#   new_terms   — the exception scenario on a new terms rule (new host.name over a 7d
#                 history window), with its own "-nt" host names so the other scenarios'
#                 events do not make them old news. Each host alerts once when first seen,
#                 so the preview reads "~4 now, ~3 after". Re-seeding within 7 days without
#                 `clean` leaves the hosts in the history window and the preview at 0.
#   threat_match
#               — the exception scenario on an indicator match rule. Indicators for the four
#                 seeded hosts live in logs-ti_seedbacktest-default, matched on host.name, so
#                 every event matches and the preview reads ~"20 now, 5 after".
#   esql        — the schedule scenario on an ES|QL rule. Exceptions on ES|QL rules cannot
#                 be previewed (no filters field), so this exercises the schedule arm instead.
#
# 12 false positives clears the tuning worker's min_fp_count of 10, and since they are
# the rule's only alerts the FP rate is 100% (min is 50%).
#
# Source events live in logs-seedbacktest-default. The preview runs 2 invocations of
# the 5m interval with a now-6m lookback, ending at the newest alert timestamp, so
# events sit 1..10 minutes before TS to land inside that window.

set -euo pipefail

KB_HOST="${KB_HOST:-http://localhost:5601}"
ES_HOST="${ES_HOST:-http://localhost:9200}"
ES_CREDS="${ES_CREDS:-elastic:changeme}"
SPACE="${SPACE:-default}"
ALERTS_INDEX=".alerts-security.alerts-${SPACE}"
EVENTS_STREAM="logs-seedbacktest-default"
# Indicators for the threat_match scenario. Under logs-* because the tuning worker's
# service account only reads logs-* (plus beats/apm patterns); the preview silently
# finds no indicators otherwise. The indicator docs carry no process.name, so the
# rule's own query never reads them as source events.
THREAT_INDEX="logs-ti_seedbacktest-default"
# Everything is seeded this many minutes in the past. The seeded rule is live and
# looks back 6m, so events this old never become real alerts (which would dilute the
# FP rate below the worker's 50% floor); the preview backtest anchors at the newest
# false positive, so it still sees them, and the 7-day harvest window is unaffected.
SEED_OFFSET_MIN="${SEED_OFFSET_MIN:-120}"

cmd="${1:-seed}"
scenario="${2:-exception}"
RULE_ID_TAG="seed-${scenario}-backtest-$(date +%s)"
# One state file per scenario, so several can be seeded at once and the worker
# fans out one review per rule; `status`/`clean` take the same scenario argument.
STATE_FILE="${STATE_FILE:-/tmp/seed-${scenario}-backtest-state.json}"

es() { curl -s -u "$ES_CREDS" -H "Content-Type: application/json" "$@"; }
kb() { curl -s -u "$ES_CREDS" -H "kbn-xsrf: true" -H "elastic-api-version: 2023-10-31" -H "Content-Type: application/json" "$@"; }

ts_minutes_ago() {
    date -u -v-"$1"M +%Y-%m-%dT%H:%M:%S.000Z 2>/dev/null ||
        date -u -d "-$1 minutes" +%Y-%m-%dT%H:%M:%S.000Z
}

# Alerts land at TS; events 1..10 minutes before it (see SEED_OFFSET_MIN).
TS=$(ts_minutes_ago "$SEED_OFFSET_MIN")
event_ts() { ts_minutes_ago $((SEED_OFFSET_MIN + $1)); }

# Per-scenario rule/alert shape, set by configure_scenario. RULE_REVISION is read back
# from the created rule.
RULE_NAME=""
RULE_CATEGORY=""
RULE_TYPE_ID=""
SEVERITY=""
RISK_SCORE=""
RULE_REVISION=0

configure_scenario() {
    case "$scenario" in
    exception)
        RULE_NAME="Seeded rule for exception backtest"
        RULE_CATEGORY="Custom Query Rule"
        RULE_TYPE_ID="siem.queryRule"
        SEVERITY="high"
        RISK_SCORE=47
        ;;
    data_view)
        RULE_NAME="Seeded rule for data view backtest"
        RULE_CATEGORY="Custom Query Rule"
        RULE_TYPE_ID="siem.queryRule"
        SEVERITY="high"
        RISK_SCORE=47
        ;;
    timestamp_override)
        RULE_NAME="Seeded rule for timestamp override backtest"
        RULE_CATEGORY="Custom Query Rule"
        RULE_TYPE_ID="siem.queryRule"
        SEVERITY="high"
        RISK_SCORE=47
        ;;
    alert_suppression)
        RULE_NAME="Seeded rule for alert suppression backtest"
        RULE_CATEGORY="Custom Query Rule"
        RULE_TYPE_ID="siem.queryRule"
        SEVERITY="high"
        RISK_SCORE=47
        ;;
    eql)
        RULE_NAME="Seeded EQL rule for exception backtest"
        RULE_CATEGORY="Event Correlation Rule"
        RULE_TYPE_ID="siem.eqlRule"
        SEVERITY="high"
        RISK_SCORE=47
        ;;
    esql)
        RULE_NAME="Seeded ES|QL rule for schedule tuning"
        RULE_CATEGORY="ES|QL Rule"
        RULE_TYPE_ID="siem.esqlRule"
        SEVERITY="medium"
        RISK_SCORE=47
        ;;
    new_terms)
        RULE_NAME="Seeded new terms rule for exception backtest"
        RULE_CATEGORY="New Terms Rule"
        RULE_TYPE_ID="siem.newTermsRule"
        SEVERITY="high"
        RISK_SCORE=47
        HOST_SUFFIX="-nt"
        ;;
    threat_match)
        RULE_NAME="Seeded indicator match rule for exception backtest"
        RULE_CATEGORY="Indicator Match Rule"
        RULE_TYPE_ID="siem.indicatorRule"
        SEVERITY="high"
        RISK_SCORE=47
        ;;
    risk_score)
        RULE_NAME="Successful logon audit (informational)"
        RULE_CATEGORY="Custom Query Rule"
        RULE_TYPE_ID="siem.queryRule"
        SEVERITY="high"
        RISK_SCORE=73
        ;;
    threshold)
        RULE_NAME="Seeded rule for threshold tuning"
        RULE_CATEGORY="Threshold Rule"
        RULE_TYPE_ID="siem.thresholdRule"
        SEVERITY="medium"
        RISK_SCORE=47
        ;;
    threshold_query)
        RULE_NAME="Seeded threshold rule for query tuning"
        RULE_CATEGORY="Threshold Rule"
        RULE_TYPE_ID="siem.thresholdRule"
        SEVERITY="medium"
        RISK_SCORE=47
        ;;
    schedule)
        RULE_NAME="Seeded rule for schedule tuning"
        RULE_CATEGORY="Custom Query Rule"
        RULE_TYPE_ID="siem.queryRule"
        SEVERITY="medium"
        RISK_SCORE=47
        ;;
    *)
        echo "Unknown scenario '$scenario' (exception|risk_score|threshold|threshold_query|schedule|data_view|timestamp_override|alert_suppression|eql|esql|new_terms|threat_match)" >&2
        exit 1
        ;;
    esac
}

# Host names are suffixed for the new_terms scenario so they are new to the 7d history window.
HOST_SUFFIX=""
host_name() { printf '%s%s' "$1" "$HOST_SUFFIX"; }

# $1 id, $2 host, $3 rule uuid, $4 reason, $5 extra top-level JSON fields (may be empty),
# $6 alert timestamp (default TS), $7 original event time (default $6)
alert() {
    local id="$1" host rule_uuid="$3" reason="$4" extra="${5:-}" ts="${6:-$TS}"
    host=$(host_name "$2")
    local original_time="${7:-$ts}"
    es -X PUT "$ES_HOST/$ALERTS_INDEX/_doc/$id?refresh=true" -d "{
    \"@timestamp\": \"$ts\",
    \"kibana.alert.uuid\": \"$id\",
    \"kibana.alert.rule.uuid\": \"$rule_uuid\",
    \"kibana.alert.rule.name\": \"$RULE_NAME\",
    \"kibana.alert.rule.category\": \"$RULE_CATEGORY\",
    \"kibana.alert.rule.consumer\": \"siem\",
    \"kibana.alert.rule.producer\": \"siem\",
    \"kibana.alert.rule.rule_type_id\": \"$RULE_TYPE_ID\",
    \"kibana.alert.rule.revision\": $RULE_REVISION,
    \"kibana.alert.workflow_status\": \"closed\",
    \"kibana.alert.workflow_reason\": \"false_positive\",
    \"kibana.alert.workflow_tags\": [],
    \"kibana.alert.status\": \"closed\",
    \"kibana.alert.severity\": \"$SEVERITY\",
    \"kibana.alert.risk_score\": $RISK_SCORE,
    \"kibana.alert.reason\": \"$reason\",
    \"kibana.alert.original_time\": \"$original_time\",
    \"kibana.space_ids\": [\"$SPACE\"],
    \"host\": { \"name\": \"$host\" },
    \"event\": { \"kind\": \"signal\", \"action\": \"closed\" }
    ${extra:+, $extra}
  }" | jq -c '{id: ._id, result: .result, host: "'"$host"'"}'
}

# One bulk "create" pair for a process event $1 minutes before TS on host $2,
# process name $4 (default backup-agent.exe).
process_event_lines() {
    local idx="$3" process="${4:-backup-agent.exe}"
    printf '{"create":{"_id":"seed-ev-%s-%s"}}\n' "$RULE_ID_TAG" "$idx"
    local ts
    ts=$(event_ts "$1")
    # event.ingested mirrors @timestamp so the timestamp_override scenario sees the same window.
    printf '{"@timestamp":"%s","host":{"name":"%s"},"process":{"name":"%s"},"event":{"kind":"event","category":["process"],"dataset":"seedbacktest","ingested":"%s"},"data_stream":{"type":"logs","dataset":"seedbacktest","namespace":"default"},"tags":["%s"]}\n' \
        "$ts" "$(host_name "$2")" "$process" "$ts" "$RULE_ID_TAG"
}

# One bulk "create" pair for a successful logon $1 minutes before TS on host $2 by user $3.
auth_event_lines() {
    local idx="$4"
    printf '{"create":{"_id":"seed-ev-%s-%s"}}\n' "$RULE_ID_TAG" "$idx"
    printf '{"@timestamp":"%s","host":{"name":"%s"},"user":{"name":"%s"},"event":{"kind":"event","category":["authentication"],"outcome":"success","action":"logged-in","dataset":"seedbacktest"},"data_stream":{"type":"logs","dataset":"seedbacktest","namespace":"default"},"tags":["%s"]}\n' \
        "$(event_ts "$1")" "$2" "$3" "$RULE_ID_TAG"
}

bulk() {
    es -X POST "$ES_HOST/$EVENTS_STREAM/_bulk?refresh=true" -d "$1" |
        jq -c '{items: (.items | length), errors}'
}

# One indicator per seeded host, tagged like the events so clean can find them.
seed_indicators() {
    local body="" host i=0 ts
    ts=$(event_ts 30)
    echo "Seeding 4 host indicators into $THREAT_INDEX..."
    for host in admin-workstation-04 build-agent-02 db-host-11 jump-box-03; do
        body+=$(printf '{"create":{"_id":"seed-ioc-%s-%s"}}\n' "$RULE_ID_TAG" "$i")$'\n'
        body+=$(printf '{"@timestamp":"%s","threat":{"indicator":{"type":"host","host":{"name":"%s"}}},"event":{"kind":"enrichment","category":["threat"],"dataset":"ti_seedbacktest"},"data_stream":{"type":"logs","dataset":"ti_seedbacktest","namespace":"default"},"tags":["%s"]}' "$ts" "$host" "$RULE_ID_TAG")$'\n'
        i=$((i + 1))
    done
    es -X POST "$ES_HOST/$THREAT_INDEX/_bulk?refresh=true" -d "$body" | jq -c '{items: (.items | length), errors}'
}

# Kibana creates the alerts index (an alias onto .internal.alerts-security.alerts-*)
# the first time a security rule runs. Writing alerts before that auto-creates a plain,
# dynamically mapped index under the alias name, which breaks the worker's ES|QL
# harvest and blocks Kibana from ever creating the real one. The rule created above
# runs within seconds of being enabled, so poll for the alias instead of racing it.
wait_for_alerts_index() {
    local i resolved
    for i in $(seq 1 60); do
        resolved=$(es "$ES_HOST/_resolve/index/$ALERTS_INDEX")
        if [ "$(echo "$resolved" | jq '.aliases | length')" -gt 0 ]; then
            return 0
        fi
        if [ "$(echo "$resolved" | jq '.indices | length')" -gt 0 ]; then
            echo "$ALERTS_INDEX exists as a plain index, not Kibana's alias. Delete it and re-run:" >&2
            echo "  curl -u \$ES_CREDS -X DELETE $ES_HOST/$ALERTS_INDEX" >&2
            exit 1
        fi
        [ "$i" -eq 1 ] && echo "Waiting for Kibana to create $ALERTS_INDEX (first rule run)..."
        sleep 2
    done
    echo "Timed out waiting for $ALERTS_INDEX. Is the seeded rule running?" >&2
    exit 1
}

HOSTS_12=(admin-workstation-04 build-agent-02 db-host-11 jump-box-03 web-frontend-07
    hr-laptop-19 finance-pc-03 dev-vm-22 mail-relay-01 print-server-05
    kiosk-lobby-02 vpn-gateway-09)
USERS_12=(a.nguyen b.okafor c.silva d.mehta e.larsen f.kowalski g.haddad h.ito
    i.brennan j.okoye k.petrov l.santos)
PROCESSES_27=(chrome.exe outlook.exe teams.exe code.exe node.exe git.exe python.exe
    slack.exe zoom.exe explorer.exe notepad.exe powershell.exe cmd.exe
    svchost.exe wmiprvse.exe msedge.exe onedrive.exe spoolsv.exe
    java.exe docker.exe kubectl.exe terraform.exe psql.exe redis-server.exe
    nginx.exe sshd.exe rsync.exe)
UPDATERS_12=(chrome_updater.exe teams_updater.exe zoom_updater.exe slack_updater.exe
    code_updater.exe edge_updater.exe onedrive_updater.exe adobe_updater.exe
    java_updater.exe docker_updater.exe spotify_updater.exe dropbox_updater.exe)

create_rule() {
    local body
    case "$scenario" in
    exception)
        body="{\"type\":\"query\",\"query\":\"process.name: *\",\"language\":\"kuery\",
        \"description\":\"Created by seed-tuning-backtests.sh; safe to delete\"}"
        ;;
    risk_score)
        body="{\"type\":\"query\",\"query\":\"event.category: authentication and event.outcome: success\",\"language\":\"kuery\",
        \"description\":\"Audit trail of successful logons. Intentionally matches every successful authentication so the audit record is complete; narrowing the query or excluding hosts would defeat its purpose. Created by seed-tuning-backtests.sh; safe to delete\"}"
        ;;
    threshold)
        body="{\"type\":\"threshold\",\"query\":\"process.name: *\",\"language\":\"kuery\",
        \"threshold\":{\"field\":[\"host.name\"],\"value\":3},
        \"description\":\"Hosts running an unusual number of processes. Created by seed-tuning-backtests.sh; safe to delete\"}"
        ;;
    threshold_query)
        body="{\"type\":\"threshold\",\"query\":\"process.name: *\",\"language\":\"kuery\",
        \"threshold\":{\"field\":[\"host.name\"],\"value\":3},
        \"description\":\"Hosts running an unusual number of processes. Counts any process, including routine software updaters. Created by seed-tuning-backtests.sh; safe to delete\"}"
        ;;
    schedule)
        body="{\"type\":\"query\",\"query\":\"process.name: *\",\"language\":\"kuery\",
        \"from\":\"now-65m\",\"interval\":\"5m\",
        \"description\":\"Process activity. Looks back 65 minutes on a 5 minute interval, so each run re-alerts on the same events. Created by seed-tuning-backtests.sh; safe to delete\"}"
        ;;
    data_view)
        body="{\"type\":\"query\",\"query\":\"process.name: *\",\"language\":\"kuery\",
        \"data_view_id\":\"$RULE_ID_TAG\",
        \"description\":\"Process activity on a data view. Created by seed-tuning-backtests.sh; safe to delete\"}"
        ;;
    timestamp_override)
        # Only the seeded stream carries event.ingested; other logs-* indices would make
        # every run a partial failure and add a warning to the card.
        body="{\"type\":\"query\",\"query\":\"process.name: *\",\"language\":\"kuery\",
        \"index\":[\"logs-seedbacktest-*\"],\"timestamp_override\":\"event.ingested\",
        \"description\":\"Process activity, timed on event.ingested. Created by seed-tuning-backtests.sh; safe to delete\"}"
        ;;
    alert_suppression)
        body="{\"type\":\"query\",\"query\":\"process.name: *\",\"language\":\"kuery\",
        \"alert_suppression\":{\"group_by\":[\"host.name\"],\"duration\":{\"value\":5,\"unit\":\"m\"},\"missing_fields_strategy\":\"suppress\"},
        \"description\":\"Process activity, one alert per host per 5 minutes. Created by seed-tuning-backtests.sh; safe to delete\"}"
        ;;
    eql)
        body="{\"type\":\"eql\",\"query\":\"process where process.name != null\",\"language\":\"eql\",
        \"description\":\"Process activity (EQL). Created by seed-tuning-backtests.sh; safe to delete\"}"
        ;;
    esql)
        body="{\"type\":\"esql\",\"query\":\"FROM logs-* METADATA _id, _index, _version | WHERE process.name IS NOT NULL\",\"language\":\"esql\",
        \"from\":\"now-65m\",\"interval\":\"5m\",
        \"description\":\"Process activity (ES|QL). Looks back 65 minutes on a 5 minute interval, so each run re-alerts on the same events. Created by seed-tuning-backtests.sh; safe to delete\"}"
        ;;
    new_terms)
        body="{\"type\":\"new_terms\",\"query\":\"process.name: *\",\"language\":\"kuery\",
        \"new_terms_fields\":[\"host.name\"],\"history_window_start\":\"now-7d\",
        \"description\":\"Hosts seen for the first time in 7 days. Created by seed-tuning-backtests.sh; safe to delete\"}"
        ;;
    threat_match)
        body="{\"type\":\"threat_match\",\"query\":\"process.name: *\",\"language\":\"kuery\",
        \"threat_index\":[\"$THREAT_INDEX\"],\"threat_query\":\"*:*\",\"threat_language\":\"kuery\",
        \"threat_mapping\":[{\"entries\":[{\"field\":\"host.name\",\"type\":\"mapping\",\"value\":\"threat.indicator.host.name\"}]}],
        \"description\":\"Process activity on hosts listed in the indicator index. Created by seed-tuning-backtests.sh; safe to delete\"}"
        ;;
    esac
    # Scenario body wins over the defaults (schedule overrides from/interval). A rule on
    # a data view carries no index patterns, and an ES|QL rule names its index in the query.
    kb -X POST "$KB_HOST/api/detection_engine/rules" -d "$(jq -n \
        --arg name "$RULE_NAME" --arg rule_id "$RULE_ID_TAG" --arg severity "$SEVERITY" \
        --argjson risk "$RISK_SCORE" --argjson body "$body" \
        '{name: $name, rule_id: $rule_id, severity: $severity, risk_score: $risk,
      enabled: true, index: ["logs-*"], from: "now-6m", interval: "5m"} + $body
      | if .data_view_id or .type == "esql" then del(.index) else . end')" | jq -r '"\(.id) \(.revision)"'
}

# The data_view scenario points its rule at a Kibana data view over the same index
# pattern. The data view id is the rule tag, so clean can find it without extra state.
create_data_view() {
    kb -X POST "$KB_HOST/api/data_views/data_view" -d "$(jq -n --arg id "$RULE_ID_TAG" \
        '{data_view: {id: $id, title: "logs-*", name: ("Seed backtest " + $id), timeFieldName: "@timestamp"}}')" |
        jq -r '.data_view.id // ("data view creation failed: " + (.message // tostring))'
}

seed_alerts() {
    local rule_uuid="$1" i=0 host user from
    case "$scenario" in
    exception | data_view | timestamp_override | alert_suppression | eql | new_terms | threat_match)
        echo "Seeding 12 false-positive alerts (9x admin-workstation-04, 3 others)..."
        for host in admin-workstation-04 admin-workstation-04 admin-workstation-04 \
            admin-workstation-04 admin-workstation-04 admin-workstation-04 \
            admin-workstation-04 admin-workstation-04 admin-workstation-04 \
            build-agent-02 db-host-11 jump-box-03; do
            alert "seed-fp-$RULE_ID_TAG-$i" "$host" "$rule_uuid" "synthetic seeded alert"
            i=$((i + 1))
        done
        ;;
    risk_score)
        echo "Seeding 12 false-positive alerts (12 different hosts and users, routine logons spread over 6 days)..."
        # Spread over the window so they read as recurring noise, not one seeded batch.
        for i in $(seq 0 11); do
            host="${HOSTS_12[$i]}"
            user="${USERS_12[$i]}"
            alert "seed-fp-$RULE_ID_TAG-$i" "$host" "$rule_uuid" \
                "routine successful logon by $user on $host; expected audit event, closed as noise: the rule is informational and high severity is too loud for it" \
                "\"user\": {\"name\": \"$user\"}, \"event.category\": [\"authentication\"], \"event.outcome\": \"success\"" \
                "$(ts_minutes_ago $((SEED_OFFSET_MIN + i * 720)))"
        done
        ;;
    threshold)
        echo "Seeding 12 false-positive alerts (12 different hosts, counts barely over the threshold of 3)..."
        from=$(event_ts 6)
        for i in $(seq 0 11); do
            host="${HOSTS_12[$i]}"
            alert "seed-fp-$RULE_ID_TAG-$i" "$host" "$rule_uuid" \
                "$host ran $((3 + i % 3)) unrelated processes in 6 minutes, normal activity for this host" \
                "\"kibana.alert.threshold_result\": {\"count\": $((3 + i % 3)), \"from\": \"$from\", \"terms\": [{\"field\": \"host.name\", \"value\": \"$host\"}]}"
        done
        ;;
    threshold_query)
        echo "Seeding 12 false-positive alerts (12 different hosts, counts 10-12, all *_updater.exe processes)..."
        from=$(event_ts 6)
        for i in $(seq 0 11); do
            host="${HOSTS_12[$i]}"
            alert "seed-fp-$RULE_ID_TAG-$i" "$host" "$rule_uuid" \
                "all $((10 + i % 3)) processes on $host were software updaters (${UPDATERS_12[$i]} and similar *_updater.exe) during the patch window; routine, not unusual activity" \
                "\"process\": {\"name\": \"${UPDATERS_12[$i]}\"}, \"kibana.alert.threshold_result\": {\"count\": $((10 + i % 3)), \"from\": \"$from\", \"terms\": [{\"field\": \"host.name\", \"value\": \"$host\"}]}"
        done
        ;;
    schedule | esql)
        echo "Seeding 12 false-positive alerts (3 source events x 4 duplicate alerts, 5m apart)..."
        local run orig_ts alert_ts
        for i in $(seq 0 11); do
            host="${HOSTS_12[$((i % 3))]}"
            run=$((i / 3))
            orig_ts=$(event_ts $((45 + 5 * (i % 3))))
            alert_ts=$(ts_minutes_ago $((SEED_OFFSET_MIN + 5 * run)))
            alert "seed-fp-$RULE_ID_TAG-$i" "$host" "$rule_uuid" \
                "duplicate: event with process backup-agent.exe on $host at $orig_ts re-matched by the 65m lookback on run $((run + 1))" \
                "\"process\": {\"name\": \"backup-agent.exe\"}" "$alert_ts" "$orig_ts"
        done
        ;;
    esac
}

seed_events() {
    local body="" i=0 host
    case "$scenario" in
    exception | data_view | timestamp_override | alert_suppression | eql | new_terms | threat_match)
        echo "Seeding 20 source events (15x admin-workstation-04, 5 others)..."
        for host in admin-workstation-04 admin-workstation-04 admin-workstation-04 \
            admin-workstation-04 admin-workstation-04 admin-workstation-04 \
            admin-workstation-04 admin-workstation-04 admin-workstation-04 \
            admin-workstation-04 admin-workstation-04 admin-workstation-04 \
            admin-workstation-04 admin-workstation-04 admin-workstation-04 \
            build-agent-02 db-host-11 jump-box-03 build-agent-02 db-host-11; do
            body+=$(process_event_lines $((i % 10 + 1)) "$host" "$i")$'\n'
            i=$((i + 1))
        done
        ;;
    risk_score)
        echo "Seeding 24 successful-logon events across 12 hosts..."
        for i in $(seq 0 23); do
            body+=$(auth_event_lines $((i % 10 + 1)) "${HOSTS_12[$((i % 12))]}" "${USERS_12[$((i % 12))]}" "$i")$'\n'
        done
        ;;
    threshold)
        echo "Seeding 27 process events, each a different process (15x admin-workstation-04, 4 each on 3 others)..."
        for host in admin-workstation-04 admin-workstation-04 admin-workstation-04 \
            admin-workstation-04 admin-workstation-04 admin-workstation-04 \
            admin-workstation-04 admin-workstation-04 admin-workstation-04 \
            admin-workstation-04 admin-workstation-04 admin-workstation-04 \
            admin-workstation-04 admin-workstation-04 admin-workstation-04 \
            build-agent-02 build-agent-02 build-agent-02 build-agent-02 \
            db-host-11 db-host-11 db-host-11 db-host-11 \
            jump-box-03 jump-box-03 jump-box-03 jump-box-03; do
            # Distinct process names: a shared one reads as a known-good entity to except.
            body+=$(process_event_lines $((i % 10 + 1)) "$host" "$i" "${PROCESSES_27[$i]}")$'\n'
            i=$((i + 1))
        done
        ;;
    threshold_query)
        echo "Seeding 16 process events (4 each on 4 hosts; 3 hosts run only *_updater.exe, db-host-11 runs normal processes)..."
        for host in admin-workstation-04 build-agent-02 jump-box-03; do
            for j in 0 1 2 3; do
                body+=$(process_event_lines $((i % 10 + 1)) "$host" "$i" "${UPDATERS_12[$i]}")$'\n'
                i=$((i + 1))
            done
        done
        for j in 0 1 2 3; do
            body+=$(process_event_lines $((i % 10 + 1)) "db-host-11" "$i" "${PROCESSES_27[$j]}")$'\n'
            i=$((i + 1))
        done
        ;;
    schedule | esql)
        echo "Seeding 20 process events spread over the hour before TS (3..60 minutes)..."
        for i in $(seq 0 19); do
            body+=$(process_event_lines $((3 + 3 * i)) "${HOSTS_12[$((i % 3))]}" "$i")$'\n'
        done
        ;;
    esac
    bulk "$body"
}

case "$cmd" in
seed)
    configure_scenario
    echo "Scenario: $scenario"
    if [ "$scenario" = "data_view" ]; then
        echo "Creating data view: $(create_data_view)"
    fi
    # The indicator index must exist before the rule's first run.
    if [ "$scenario" = "threat_match" ]; then
        seed_indicators
    fi
    echo "Creating rule..."
    # The tuning worker only harvests false positives stamped with the rule's current
    # revision (kibana.alert.rule.revision), so the seeded alerts carry it.
    read -r rule_uuid RULE_REVISION <<<"$(create_rule)"
    if [ -z "$rule_uuid" ] || [ "$rule_uuid" = "null" ]; then
        echo "Rule creation failed" >&2
        exit 1
    fi
    echo "Rule created: $rule_uuid"
    echo "$RULE_ID_TAG" >"$STATE_FILE"

    echo ""
    wait_for_alerts_index
    seed_alerts "$rule_uuid"
    echo ""
    seed_events
    echo ""
    echo "Done. Run the rule tuning worker from the Workflows UI; it will pick this rule up."
    ;;

status)
    rule_id=$(cat "$STATE_FILE" 2>/dev/null || true)
    if [ -z "$rule_id" ]; then
        echo "Nothing seeded (no $STATE_FILE)."
        exit 0
    fi
    kb "$KB_HOST/api/detection_engine/rules?rule_id=$rule_id" | jq '{id, rule_id, name, type, enabled, severity, risk_score, threshold}'
    es "$ES_HOST/$ALERTS_INDEX/_count" -d "{\"query\":{\"prefix\":{\"kibana.alert.uuid\":\"seed-fp-$rule_id\"}}}" | jq '{seeded_alerts: .count}'
    es "$ES_HOST/$EVENTS_STREAM/_count" -d "{\"query\":{\"term\":{\"tags\":\"$rule_id\"}}}" | jq '{seeded_events: .count}'
    ;;

clean)
    rule_id=$(cat "$STATE_FILE" 2>/dev/null || true)
    if [ -z "$rule_id" ]; then
        echo "Nothing seeded (no $STATE_FILE)."
        exit 0
    fi
    rule_uuid=$(kb "$KB_HOST/api/detection_engine/rules?rule_id=$rule_id" | jq -r '.id // empty' || true)
    if [ -z "$rule_uuid" ]; then
        echo "Could not look up rule $rule_id (is Kibana up?). Nothing deleted." >&2
        exit 1
    fi
    echo "Deleting rule $rule_id..."
    kb -X DELETE "$KB_HOST/api/detection_engine/rules?rule_id=$rule_id" >/dev/null || true
    if [ "$scenario" = "data_view" ]; then
        echo "Deleting data view $rule_id..."
        kb -X DELETE "$KB_HOST/api/data_views/data_view/$rule_id" >/dev/null || true
    fi
    # Seeded alerts plus any the live rule produced on its own.
    echo "Deleting the rule's alerts..."
    es -X POST "$ES_HOST/$ALERTS_INDEX/_delete_by_query?refresh=true" \
        -d "{\"query\":{\"bool\":{\"should\":[{\"prefix\":{\"kibana.alert.uuid\":\"seed-fp-$rule_id\"}},{\"term\":{\"kibana.alert.rule.uuid\":\"${rule_uuid:-none}\"}}],\"minimum_should_match\":1}}}" | jq '{deleted}'
    echo "Deleting seeded source events..."
    es -X POST "$ES_HOST/$EVENTS_STREAM/_delete_by_query?refresh=true" \
        -d "{\"query\":{\"term\":{\"tags\":\"$rule_id\"}}}" | jq '{deleted}'
    if [ "$scenario" = "threat_match" ]; then
        echo "Deleting seeded indicators..."
        es -X POST "$ES_HOST/$THREAT_INDEX/_delete_by_query?refresh=true" \
            -d "{\"query\":{\"term\":{\"tags\":\"$rule_id\"}}}" | jq '{deleted}'
    fi
    rm -f "$STATE_FILE"
    echo "Cleaned up."
    ;;

*)
    echo "Usage: $0 {seed|status|clean} [exception|risk_score|threshold|threshold_query|schedule|data_view|timestamp_override|alert_suppression|eql|esql|new_terms|threat_match]" >&2
    exit 1
    ;;
esac
