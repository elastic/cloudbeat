#!/usr/bin/env bash
#
# AlertZero QA data seeder: seeds detection coverage "gap" documents
# (knowledge indicators with type security.coverage) for exercising the
# AlertZero Rule Coverage worker. Originally vendored from
# https://gist.github.com/hannahbrooks/3746ab4935d3599ed3f8f436328ad463 and
# reworked so that every `seed` case leads to its intended verdict even when
# the full prebuilt rule catalog is installed:
#
#   covered_disabled    Gap built from a disabled rule's own MITRE technique and
#                       a paraphrase of its description (never its name). No
#                       enabled rule maps to that technique.
#                       Expected verdict: enable the existing rule.
#   covered_enabled     Same, from an enabled rule. Skipped when no enabled rule
#                       has a MITRE mapping.
#                       Expected verdict: coverage confirmed.
#   prebuilt_available  DELETES every installed rule mapped to PREBUILT_TECHNIQUE,
#                       confirms a prebuilt rule for it is installable, then
#                       writes a gap from that rule's description.
#                       Expected verdict: install the prebuilt rule.
#                       Undo: Rules > Detection rules > Add Elastic rules,
#                       reinstall the deleted rules (listed in the output).
#   no_coverage         First technique in NO_COVERAGE_TECHNIQUES with no
#                       installed rule and no installable prebuilt rule.
#                       Expected verdict: create a new rule.
#
# The worker only picks gaps with attributes.status "pending", a matching
# attributes.space_id, and an @timestamp inside its lookback (14 days by
# default). It reads title, content and attributes.technique.
#
# Usage: seed-coverage-kis.sh [command]
#
# Commands:
#   seed       one gap per verdict, as described above (default)
#   generic    GENERIC_COUNT generic gaps for GENERIC_TECHNIQUE
#   malformed  one gap each with: no technique, no title, no content, no
#              status, and `attributes` as a string (the index mapping may
#              reject the last one; that is reported, not treated as an error)
#   no-space   one otherwise valid gap without attributes.space_id
#   reset      mark all coverage gaps pending (re-run sweep)
#   clear      mark all coverage gaps processed (start fresh)
#   status     list pending coverage gaps
#   rules      show installed rules and prebuilt rule status
#
# Required env:
#   KB_HOST    Kibana URL, may include a base path (http://localhost:5611/kbn)
#   ES_HOST    Elasticsearch URL
#   ES_CREDS   user:password, used for both Kibana and Elasticsearch
#
# Optional env:
#   SPACE                   Kibana space id; rules are read from this space and
#                           gaps get attributes.space_id=SPACE (default: default)
#   GAP_AGE_DAYS            @timestamp of written gaps is now minus this many
#                           days; use > 14 to fall outside the default
#                           lookback (default: 0)
#   PREBUILT_TECHNIQUE      ATT&CK id for prebuilt_available (default: T1003.001)
#   NO_COVERAGE_TECHNIQUES  space-separated ATT&CK ids tried in order for
#                           no_coverage (default: T1610 T1612 T1535 T1564.011 T1205.002)
#   GENERIC_COUNT           number of gaps written by `generic` (default: 5)
#   GENERIC_TECHNIQUE       ATT&CK id used by `generic`, `malformed` and
#                           `no-space` (default: T1059)
#
# Requires: jq, curl

set -euo pipefail

command -v jq >/dev/null 2>&1 || {
    echo "seed-coverage-kis.sh: jq is required" >&2
    exit 1
}

: "${KB_HOST:?KB_HOST must be set}"
: "${ES_HOST:?ES_HOST must be set}"
: "${ES_CREDS:?ES_CREDS must be set}"

readonly SPACE="${SPACE:-default}"
readonly GAP_AGE_DAYS="${GAP_AGE_DAYS:-0}"
readonly PREBUILT_TECHNIQUE="${PREBUILT_TECHNIQUE:-T1003.001}"
readonly NO_COVERAGE_TECHNIQUES="${NO_COVERAGE_TECHNIQUES:-T1610 T1612 T1535 T1564.011 T1205.002}"
readonly GENERIC_COUNT="${GENERIC_COUNT:-5}"
readonly GENERIC_TECHNIQUE="${GENERIC_TECHNIQUE:-T1059}"

readonly INDEX="ai-index-idx-security-investigations"
readonly CURL_COMMON=(--connect-timeout 10 --max-time 120 -sS)
readonly PUBLIC_API_VERSION="2023-10-31"
# Internal API: Kibana rejects calls without x-elastic-internal-origin when
# server.restrictInternalApis is on (the default).
readonly REVIEW_INSTALLATION_PATH="/internal/detection_engine/prebuilt_rules/installation/_review"
readonly KQL_TECHNIQUE_FIELD="alert.attributes.params.threat.technique.id"
readonly KQL_SUBTECHNIQUE_FIELD="alert.attributes.params.threat.technique.subtechnique.id"

cmd="${1:-seed}"

# Everything below ends up in a URL, a KQL filter, a JSON body or a
# timestamp; validate it before use.
[[ "${SPACE}" =~ ^[a-z0-9_-]+$ ]] || {
    echo "Invalid SPACE '${SPACE}': expected lowercase letters, digits, '_' or '-'" >&2
    exit 1
}
[[ "${GAP_AGE_DAYS}" =~ ^(0|[1-9][0-9]{0,3})$ ]] || {
    echo "Invalid GAP_AGE_DAYS '${GAP_AGE_DAYS}': expected a whole number of days (0-9999)" >&2
    exit 1
}
[[ "${GENERIC_COUNT}" =~ ^[1-9][0-9]{0,2}$ ]] || {
    echo "Invalid GENERIC_COUNT '${GENERIC_COUNT}': expected 1-999" >&2
    exit 1
}

validate_technique() {
    [[ "$1" =~ ^T[0-9]{4}(\.[0-9]{3})?$ ]] || {
        echo "Invalid ATT&CK technique id '$1': expected Tnnnn or Tnnnn.nnn" >&2
        exit 1
    }
}
validate_technique "${PREBUILT_TECHNIQUE}"
validate_technique "${GENERIC_TECHNIQUE}"
for t in ${NO_COVERAGE_TECHNIQUES}; do
    validate_technique "${t}"
done

readonly KB_BASE="${KB_HOST%/}"
readonly ES_BASE="${ES_HOST%/}"
if [[ "${SPACE}" == "default" ]]; then
    readonly KB_SPACE_BASE="${KB_BASE}"
else
    readonly KB_SPACE_BASE="${KB_BASE}/s/${SPACE}"
fi

TS="$(jq -nr --argjson days "${GAP_AGE_DAYS}" '(now - $days * 86400) | strftime("%Y-%m-%dT%H:%M:%SZ")')"
readonly TS

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "${WORK_DIR}"' EXIT

# Credentials are passed via curl's -u flag and never echoed or written to logs.
kb_curl() {
    curl "${CURL_COMMON[@]}" -u "${ES_CREDS}" -H 'kbn-xsrf: true' "$@"
}

es_curl() {
    curl "${CURL_COMMON[@]}" -u "${ES_CREDS}" "$@"
}

# expect_status <what> <expected http code> <actual http code> <response file>
expect_status() {
    if [[ "$3" != "$2" ]]; then
        echo "$1 failed (HTTP $3)" >&2
        head -c 2000 "$4" >&2 || true
        echo >&2
        exit 1
    fi
}

# Shared jq definitions for reading rules. Installed rules (_find) and
# installable prebuilt rules (installation review) share the RuleResponse shape.
JQ_LIB=$(
    cat <<'EOF'
def first_or_null: if length > 0 then .[0] else null end;

# Most specific mapping of the first MITRE technique: its first subtechnique
# when present, else the technique itself.
def gap_technique:
  [ .threat[]? | select(.framework == "MITRE ATT&CK") | .technique[]?
    | if ((.subtechnique // []) | length) > 0 then .subtechnique[0] else . end
    | {id, name} ] | first_or_null;

def technique_entry($want):
  if $want == "" then gap_technique
  else [ .threat[]?.technique[]? | (., .subtechnique[]?) | select(.id == $want) | {id, name} ] | first_or_null
  end;

# Same matching as the KQL filter built by technique_kql: technique id for a
# technique, subtechnique id for a subtechnique.
def maps_to($t): any(.threat[]?.technique[]?; .id == $t or any(.subtechnique[]?; .id == $t));

def technique_ids: [ .threat[]?.technique[]? | .id, .subtechnique[]?.id ];

def detection_lead: "^(this rule |the rule )?(identifies|detects|indicates|monitors for|looks for|alerts on|alerts when|captures|flags|generates an alert (for|when)) ";

# The sentence of a rule description that says what it detects (else its first
# sentence), reworded as observed activity so the gap describes behaviour
# instead of quoting the rule.
def paraphrase:
  [ gsub("\\s+"; " ") | ltrimstr(" ") | splits("(?<=[.!?]) ") | select(length > 0) ] as $sentences
  | ([ $sentences[] | select(test(detection_lead; "i")) ] + $sentences) | first_or_null // ""
  | sub(detection_lead; ""; "i")
  | if test("^[A-Z][a-z]") then (.[0:1] | ascii_downcase) + .[1:] else . end;

def gap_content:
  "A threat hunt surfaced the following activity and we need to know whether it is detected: " + paraphrase;

def says_what_it_detects: any(splits("(?<=[.!?]) "); test(detection_lead; "i"));

# Skips rule kinds that make poor gap sources: ML rules, building blocks, and
# rules that only promote alerts from another product (e.g. Elastic Endgame).
def usable_rule:
  .type != "machine_learning"
  and (.building_block_type // null) == null
  and ((.tags // []) | any(. == "Rule Type: Promotion" or . == "Rule Type: Higher-Order") | not)
  and ((.description // "") | length) > 0;

# First usable rule whose technique is not in $avoid (neither the id nor its
# parent technique) and whose gap text does not contain the rule name. Rules
# whose description says what they detect are preferred.
def pick_rule($want; $avoid):
  [ .[] | select(usable_rule) | . as $r
    | ($r | technique_entry($want)) as $t
    | select($t != null)
    | select(($avoid | index($t.id)) == null and ($avoid | index($t.id | split(".")[0])) == null)
    | ($r.description | gap_content) as $c
    | select($c | ascii_downcase | contains($r.name | ascii_downcase) | not)
    | {id: $r.id, name: $r.name, technique: $t.id, technique_name: $t.name, content: $c,
       preferred: ($r.description | gsub("\\s+"; " ") | says_what_it_detects)} ]
  | (map(select(.preferred)) + map(select(.preferred | not))) | first_or_null
  | if . == null then null else del(.preferred) end;
EOF
)
readonly JQ_LIB

# Technique ids are validated by validate_technique, so quoting them in KQL is safe.
technique_kql() {
    if [[ "$1" == *.* ]]; then
        echo "${KQL_SUBTECHNIQUE_FIELD}: \"$1\""
    else
        echo "${KQL_TECHNIQUE_FIELD}: \"$1\""
    fi
}

# find_rules_page <out> <kql> <page> <per_page>
find_rules_page() {
    local out=$1 kql=$2 page=$3 per_page=$4 http_code
    http_code=$(kb_curl -o "${out}" -w '%{http_code}' -G \
        -H "elastic-api-version: ${PUBLIC_API_VERSION}" \
        "${KB_SPACE_BASE}/api/detection_engine/rules/_find" \
        --data-urlencode "filter=${kql}" \
        --data-urlencode "page=${page}" \
        --data-urlencode "per_page=${per_page}" \
        --data-urlencode "sort_field=name" \
        --data-urlencode "sort_order=asc")
    expect_status "Rule search (${kql})" 200 "${http_code}" "${out}"
}

# find_all_rules <kql> <out>: writes a JSON array of all matching rules.
find_all_rules() {
    local kql=$1 out=$2 page=1 total
    local page_file="${WORK_DIR}/find_page.json"
    echo '[]' >"${out}"
    while :; do
        find_rules_page "${page_file}" "${kql}" "${page}" 100
        jq -s '.[0] + .[1].data' "${out}" "${page_file}" >"${out}.tmp"
        mv "${out}.tmp" "${out}"
        total=$(jq '.total' "${page_file}")
        ((page * 100 < total)) || break
        page=$((page + 1))
    done
}

count_rules() {
    local out="${WORK_DIR}/count.json"
    find_rules_page "${out}" "$1" 1 1
    jq '.total' "${out}"
}

# fetch_installable <out>: writes a JSON array of every prebuilt rule that can
# be installed right now (not installed, allowed by license and ML privileges).
fetch_installable() {
    local out=$1 page=1 total http_code
    local page_file="${WORK_DIR}/review_page.json"
    echo '[]' >"${out}"
    while :; do
        http_code=$(kb_curl -o "${page_file}" -w '%{http_code}' -X POST \
            -H 'elastic-api-version: 1' \
            -H 'x-elastic-internal-origin: Kibana' \
            -H 'Content-Type: application/json' \
            "${KB_SPACE_BASE}${REVIEW_INSTALLATION_PATH}" \
            -d "$(jq -n --argjson page "${page}" '{page: $page, per_page: 500}')")
        expect_status "Prebuilt rule installation review" 200 "${http_code}" "${page_file}"
        jq -s '.[0] + .[1].rules' "${out}" "${page_file}" >"${out}.tmp"
        mv "${out}.tmp" "${out}"
        total=$(jq '.total' "${page_file}")
        ((page * 500 < total)) || break
        page=$((page + 1))
    done
}

delete_rules_by_kql() {
    local kql=$1 out="${WORK_DIR}/bulk_delete.json" http_code
    http_code=$(kb_curl -o "${out}" -w '%{http_code}' -X POST \
        -H "elastic-api-version: ${PUBLIC_API_VERSION}" \
        -H 'Content-Type: application/json' \
        "${KB_SPACE_BASE}/api/detection_engine/rules/_bulk_action" \
        -d "$(jq -n --arg q "${kql}" '{action: "delete", query: $q}')")
    expect_status "Bulk delete (${kql})" 200 "${http_code}" "${out}"
    if ! jq -e '.attributes.summary.failed == 0' "${out}" >/dev/null; then
        echo "Bulk delete reported failures:" >&2
        jq -c '.attributes.errors' "${out}" >&2
        exit 1
    fi
}

gap_doc() {
    local title=$1 content=$2 technique=$3
    jq -n --arg ts "${TS}" --arg title "${title}" --arg content "${content}" \
        --arg technique "${technique}" --arg space "${SPACE}" '{
        "@timestamp": $ts,
        type: "security.coverage",
        title: $title,
        content: $content,
        attributes: {status: "pending", space_id: $space, technique: $technique}
    }'
}

# write_gap <case label> <document> [allow_mapping_reject]
write_gap() {
    local label=$1 doc=$2 allow_reject=${3:-} out="${WORK_DIR}/gap_out.json" http_code
    http_code=$(es_curl -o "${out}" -w '%{http_code}' -X POST \
        "${ES_BASE}/${INDEX}/_doc?refresh=wait_for" \
        -H 'Content-Type: application/json' --data-binary "${doc}")
    if [[ "${http_code}" == "400" && -n "${allow_reject}" ]] &&
        jq -e '.error.type | test("parsing_exception")' "${out}" >/dev/null 2>&1; then
        echo "  ${label}: rejected by the index mapping ($(jq -r '.error.reason' "${out}")); the worker can never see it"
        return 0
    fi
    expect_status "Writing gap (${label})" 201 "${http_code}" "${out}"
    jq -c --arg label "${label}" --argjson doc "${doc}" '{
        case: $label,
        id: ._id,
        title: ($doc.title // null),
        technique: (($doc.attributes | objects | .technique) // null)
    }' "${out}"
}

# Sets GAP_TITLE and GAP_CONTENT for a no_coverage technique.
no_coverage_text() {
    case "$1" in
    T1610)
        GAP_TITLE="Privileged container deployed from an unvetted image"
        GAP_CONTENT="A new pod running a privileged container with hostPID: true was deployed into the kube-system namespace from an image pulled from an unknown public registry. Nothing alerted on the deployment."
        ;;
    T1612)
        GAP_TITLE="Container image built directly on a production node"
        GAP_CONTENT="docker build was run from an interactive session on a production Kubernetes node, and the resulting image was started locally instead of being pulled from the organization's registry."
        ;;
    T1535)
        GAP_TITLE="Cloud compute launched in a region the organization never uses"
        GAP_CONTENT="Several virtual machine instances were launched in a cloud region where the organization has no workloads and no prior API activity, using a long-lived access key."
        ;;
    T1564.011)
        GAP_TITLE="Process detached from its terminal to survive logout"
        GAP_CONTENT="A shell spawned by a web server process started a long-running binary under nohup with its output discarded, so it kept running after the session that launched it ended."
        ;;
    T1205.002)
        GAP_TITLE="Raw socket filter attached by an unexpected process"
        GAP_CONTENT="A process with no networking role attached a classic BPF filter to a raw socket and then stayed idle, consistent with a passive backdoor waiting for a specially crafted packet."
        ;;
    *)
        GAP_TITLE="Unmonitored activity mapped to MITRE ATT&CK $1"
        GAP_CONTENT="A threat hunt observed activity consistent with MITRE ATT&CK technique $1 that no alert fired on."
        ;;
    esac
}

seed_prebuilt_available() {
    local technique=$1 kql pick custom_count prebuilt_installed
    local installed="${WORK_DIR}/prebuilt_installed.json"
    local installable="${WORK_DIR}/installable.json"
    kql=$(technique_kql "${technique}")

    echo "--- prebuilt_available (${technique})"
    find_all_rules "${kql}" "${installed}"

    custom_count=$(jq '[.[] | select(.immutable != true)] | length' "${installed}")
    if ((custom_count > 0)); then
        echo "Custom rules map to ${technique}; deleting them cannot be undone, so stopping:" >&2
        jq -r '.[] | select(.immutable != true) | "  - \(.name) (id \(.id))"' "${installed}" >&2
        echo "Delete them yourself or set PREBUILT_TECHNIQUE to another technique." >&2
        exit 1
    fi

    # Installed prebuilt rules become installable once deleted; make sure there
    # will be something to install before deleting anything.
    fetch_installable "${installable}"
    prebuilt_installed=$(jq 'length' "${installed}")
    if ((prebuilt_installed == 0)) &&
        [[ "$(jq --arg t "${technique}" "${JQ_LIB}"' [.[] | select(maps_to($t))] | length' "${installable}")" == "0" ]]; then
        echo "No prebuilt rule maps to ${technique}; set PREBUILT_TECHNIQUE to another technique." >&2
        exit 1
    fi

    if ((prebuilt_installed > 0)); then
        echo "  Deleting ${prebuilt_installed} installed prebuilt rule(s) mapped to ${technique}:"
        jq -r '.[] | "    - \(.name)"' "${installed}"
        delete_rules_by_kql "${kql}"
        if [[ "$(count_rules "${kql}")" != "0" ]]; then
            echo "Rules mapped to ${technique} are still installed after the bulk delete" >&2
            exit 1
        fi
        echo "  To undo: Rules > Detection rules > Add Elastic rules, search for the names above and install them."
        fetch_installable "${installable}"
    fi

    pick=$(jq -c --arg t "${technique}" "${JQ_LIB}"' [.[] | select(maps_to($t))] | sort_by(.name) | pick_rule($t; [])' "${installable}")
    if [[ "${pick}" == "null" ]]; then
        echo "No usable installable prebuilt rule maps to ${technique} (license or ML privileges may hide them)." >&2
        exit 1
    fi
    echo "  Installable prebuilt rule the worker should propose: $(jq -r '.name' <<<"${pick}")"
    write_gap prebuilt_available "$(gap_doc \
        "Detection coverage gap: $(jq -r '.technique_name' <<<"${pick}") (${technique})" \
        "$(jq -r '.content' <<<"${pick}")" \
        "${technique}")"
}

seed_no_coverage() {
    local installable="${WORK_DIR}/installable.json" t chosen="" installed_count installable_count
    echo "--- no_coverage"
    [[ -f "${installable}" ]] || fetch_installable "${installable}"
    for t in ${NO_COVERAGE_TECHNIQUES}; do
        installed_count=$(count_rules "$(technique_kql "${t}")")
        installable_count=$(jq --arg t "${t}" "${JQ_LIB}"' [.[] | select(maps_to($t))] | length' "${installable}")
        if [[ "${installed_count}" == "0" && "${installable_count}" == "0" ]]; then
            chosen="${t}"
            break
        fi
        echo "  ${t}: ${installed_count} installed, ${installable_count} installable rule(s); trying the next technique"
    done
    if [[ -z "${chosen}" ]]; then
        echo "Every technique in NO_COVERAGE_TECHNIQUES (${NO_COVERAGE_TECHNIQUES}) has an installed or installable rule; add a technique with no Elastic rule." >&2
        exit 1
    fi
    no_coverage_text "${chosen}"
    write_gap no_coverage "$(gap_doc "${GAP_TITLE}" "${GAP_CONTENT}" "${chosen}")"
}

# seed_covered <case label> <enabled: true|false> <techniques to avoid, JSON array>
seed_covered() {
    local label=$1 enabled=$2 avoid=$3 page=1 total pick="null"
    local page_file="${WORK_DIR}/covered_page.json"
    echo "--- ${label}"
    while :; do
        find_rules_page "${page_file}" "alert.attributes.enabled: ${enabled}" "${page}" 100
        pick=$(jq -c --argjson avoid "${avoid}" "${JQ_LIB}"' .data | pick_rule(""; $avoid)' "${page_file}")
        [[ "${pick}" == "null" ]] || break
        total=$(jq '.total' "${page_file}")
        ((page * 100 < total)) || break
        page=$((page + 1))
    done
    if [[ "${pick}" == "null" ]]; then
        echo "  No usable rule with enabled=${enabled} and a MITRE technique mapping; skipping ${label}."
        return 0
    fi
    echo "  Rule the worker should find: $(jq -r '.name' <<<"${pick}")"
    write_gap "${label}" "$(gap_doc \
        "Detection coverage gap: $(jq -r '.technique_name' <<<"${pick}") ($(jq -r '.technique' <<<"${pick}"))" \
        "$(jq -r '.content' <<<"${pick}")" \
        "$(jq -r '.technique' <<<"${pick}")")"
}

update_all_gaps_status() {
    local new_status=$1 out="${WORK_DIR}/update_by_query.json" http_code
    http_code=$(es_curl -o "${out}" -w '%{http_code}' -X POST \
        "${ES_BASE}/${INDEX}/_update_by_query?conflicts=proceed&refresh=true" \
        -H 'Content-Type: application/json' \
        -d "$(jq -n --arg s "${new_status}" '{
            query: {term: {type: "security.coverage"}},
            script: {
                source: "if (ctx._source.attributes instanceof Map) { ctx._source.attributes.status = params.status } else { ctx.op = \"noop\" }",
                params: {status: $s}
            }
        }')")
    expect_status "Updating gap status" 200 "${http_code}" "${out}"
    jq -c '{updated, noops}' "${out}"
}

case "${cmd}" in
seed)
    echo "Seeding coverage gaps into space '${SPACE}' (@timestamp ${TS})"
    # Cases that delete rules or check for their absence run first, so the
    # covered_* lookups see the final rule state.
    seed_prebuilt_available "${PREBUILT_TECHNIQUE}"
    seed_no_coverage

    enabled_rules="${WORK_DIR}/enabled_rules.json"
    find_all_rules "alert.attributes.enabled: true" "${enabled_rules}"
    enabled_techniques=$(jq -c "${JQ_LIB}"' [.[] | technique_ids[]] | unique' "${enabled_rules}")
    seed_covered covered_disabled false "${enabled_techniques}"
    seed_covered covered_enabled true '[]'

    echo ""
    echo "Done. Trigger system-security-coverage-worker in /app/alertzero."
    ;;

generic)
    echo "Seeding ${GENERIC_COUNT} generic coverage gaps (${GENERIC_TECHNIQUE}) into space '${SPACE}' (@timestamp ${TS})"
    for ((i = 1; i <= GENERIC_COUNT; i++)); do
        write_gap "generic_${i}" "$(gap_doc \
            "QA generic coverage gap ${i}/${GENERIC_COUNT} (${GENERIC_TECHNIQUE})" \
            "Generic QA gap ${i} of ${GENERIC_COUNT}: a threat hunt observed activity mapped to MITRE ATT&CK ${GENERIC_TECHNIQUE} and we need to know whether it is detected." \
            "${GENERIC_TECHNIQUE}")"
    done
    ;;

malformed)
    echo "Seeding malformed coverage gaps into space '${SPACE}' (@timestamp ${TS})"
    base=$(gap_doc "QA malformed coverage gap" \
        "A threat hunt observed activity mapped to MITRE ATT&CK ${GENERIC_TECHNIQUE}; this gap is deliberately malformed." \
        "${GENERIC_TECHNIQUE}")
    write_gap malformed_no_technique "$(jq -c '.title += " (no technique)" | del(.attributes.technique)' <<<"${base}")"
    write_gap malformed_no_title "$(jq -c 'del(.title)' <<<"${base}")"
    write_gap malformed_no_content "$(jq -c '.title += " (no content)" | del(.content)' <<<"${base}")"
    write_gap malformed_no_status "$(jq -c '.title += " (no status)" | del(.attributes.status)' <<<"${base}")"
    write_gap malformed_attributes_string "$(jq -c '.title += " (attributes is a string)"
        | .attributes = "status=\(.attributes.status) space_id=\(.attributes.space_id) technique=\(.attributes.technique)"' <<<"${base}")" \
        allow_mapping_reject
    ;;

no-space)
    echo "Seeding a coverage gap without attributes.space_id (@timestamp ${TS})"
    write_gap no_space_id "$(gap_doc "QA coverage gap without a space (${GENERIC_TECHNIQUE})" \
        "A threat hunt observed activity mapped to MITRE ATT&CK ${GENERIC_TECHNIQUE}; this gap has no space_id." \
        "${GENERIC_TECHNIQUE}" | jq -c 'del(.attributes.space_id)')"
    ;;

reset)
    echo "Resetting all coverage gaps to pending..."
    update_all_gaps_status pending
    ;;

clear)
    echo "Marking all coverage gaps processed..."
    update_all_gaps_status processed
    ;;

status)
    echo "Pending coverage gaps:"
    out="${WORK_DIR}/status.json"
    http_code=$(es_curl -o "${out}" -w '%{http_code}' -X POST "${ES_BASE}/${INDEX}/_search" \
        -H 'Content-Type: application/json' -d '{
      "size": 100,
      "sort": [{ "@timestamp": "desc" }],
      "query": {
        "bool": {
          "filter": [
            { "term": { "type": "security.coverage" } },
            { "term": { "attributes.status": "pending" } }
          ]
        }
      }
    }')
    expect_status "Searching gaps" 200 "${http_code}" "${out}"
    jq -c '.hits.hits[] | {id: ._id, "@timestamp": ._source["@timestamp"], title: ._source.title,
        technique: ._source.attributes.technique, space_id: ._source.attributes.space_id}' "${out}"
    ;;

rules)
    out="${WORK_DIR}/rules.json"
    for enabled in false true; do
        echo "=== Rules with enabled=${enabled} (first 10) ==="
        find_rules_page "${out}" "alert.attributes.enabled: ${enabled}" 1 10
        jq -r '"total: \(.total)"' "${out}"
        jq -c "${JQ_LIB}"' .data[] | {name, rule_id, techniques: technique_ids}' "${out}"
    done
    echo "=== Prebuilt rule status ==="
    http_code=$(kb_curl -o "${out}" -w '%{http_code}' -H "elastic-api-version: ${PUBLIC_API_VERSION}" \
        "${KB_SPACE_BASE}/api/detection_engine/rules/prepackaged/_status")
    expect_status "Prebuilt rule status" 200 "${http_code}" "${out}"
    jq -c '{rules_installed, rules_not_installed}' "${out}"
    ;;

*)
    echo "Unknown command: ${cmd}" >&2
    echo "Usage: $0 [seed|generic|malformed|no-space|reset|clear|status|rules]" >&2
    exit 1
    ;;
esac
