#!/usr/bin/env bash
#
# Creates the Elastic Cloud (ECH) deployment used by the template tests, with ecctl.
# The size defaults match the former Terraform elk-stack root.
#
# This is a PUBLIC repository, so nothing sensitive may reach the job log:
#   - ecctl's JSON output (it contains the generated credentials) goes to a 0600 temp
#     file, never to stdout/stderr, and the file is deleted on exit.
#   - every secret and every URL is masked (::add-mask::) BEFORE it is written to
#     GITHUB_ENV or used anywhere else.
#   - no `set -x`, and no secret is passed on a command line.
#
# Inputs (environment):
#   EC_API_KEY        Elastic Cloud API key (secret, read by ecctl from the environment)
#   DEPLOYMENT_NAME   Name for the deployment
#   STACK_VERSION     Stack version, e.g. 9.7.0 or 9.7.0-SNAPSHOT
#   EC_REGION         Region, default gcp-us-west2
#   DEPLOYMENT_TEMPLATE, ES_SIZE_MB, ES_ZONES, KIBANA_SIZE_MB  Optional size overrides
#
# Exports to GITHUB_ENV: DEPLOYMENT_ID, KIBANA_URL, ES_URL, ES_USER, ES_PASSWORD,
#   TEST_KIBANA_URL, TEST_ES_URL (the last two embed the credentials).

set -euo pipefail
set +x

: "${EC_API_KEY:?EC_API_KEY must be set}"
: "${DEPLOYMENT_NAME:?DEPLOYMENT_NAME must be set}"
: "${STACK_VERSION:?STACK_VERSION must be set}"
: "${GITHUB_ENV:?GITHUB_ENV must be set}"

export EC_REGION="${EC_REGION:-gcp-us-west2}"
DEPLOYMENT_TEMPLATE="${DEPLOYMENT_TEMPLATE:-gcp-general-purpose}"
ES_SIZE_MB="${ES_SIZE_MB:-8192}"
ES_ZONES="${ES_ZONES:-2}"
KIBANA_SIZE_MB="${KIBANA_SIZE_MB:-4096}"

tmp_dir="$(mktemp -d "${RUNNER_TEMP:-/tmp}/ecctl-stack.XXXXXX")"
chmod 700 "${tmp_dir}"
trap 'rm -rf "${tmp_dir}"' EXIT
payload="${tmp_dir}/payload.json"
result="${tmp_dir}/create.json"

mask() {
    # Masks a value in the job log. Empty values are skipped.
    [ -n "${1:-}" ] && echo "::add-mask::$1"
    return 0
}

export_env() {
    # Writes NAME=value to GITHUB_ENV. Refuses multi-line values (GITHUB_ENV injection).
    case "$2" in
    *$'\n'*)
        echo "Refusing to export $1: value contains a newline" >&2
        exit 1
        ;;
    esac
    echo "$1=$2" >>"${GITHUB_ENV}"
}

# The template payload includes Enterprise Search, which is not allowed on 9.x stacks and
# was never part of the Terraform deployment; it is removed by the jq filter below.
echo "Generating deployment payload (${DEPLOYMENT_TEMPLATE}, ${EC_REGION}, ${STACK_VERSION})"
ecctl deployment create \
    --generate-payload \
    --name "${DEPLOYMENT_NAME}" \
    --version "${STACK_VERSION}" \
    --deployment-template "${DEPLOYMENT_TEMPLATE}" |
    jq \
        --argjson es "${ES_SIZE_MB}" \
        --argjson zones "${ES_ZONES}" \
        --argjson kb "${KIBANA_SIZE_MB}" \
        '
        (.resources.elasticsearch[0].plan.cluster_topology[] | select(.id == "hot_content"))
            |= (.size.value = $es | .zone_count = $zones)
        | .resources.kibana[0].plan.cluster_topology[0].size.value = $kb
        | del(.resources.enterprise_search)
        | .metadata.tags = [
            {key: "division", value: "engineering"},
            {key: "org", value: "security"},
            {key: "team", value: "contextual-security"},
            {key: "project", value: "template-tests"},
            {key: "owner", value: "cloudbeat"}
          ]
        ' >"${payload}"

echo "Creating deployment ${DEPLOYMENT_NAME} (this takes a few minutes)"
create_rc=0
# stdout is the JSON stream with the credentials: keep it in the temp file.
ecctl deployment create --file "${payload}" --track --output json >"${result}" || create_rc=$?

# Export the ID first so the cleanup step can shut the deployment down even if the
# create failed half way. Deployment IDs are not secret.
deployment_id="$(jq -r --slurp '[.[] | select(.id != null) | .id][0] // empty' "${result}" 2>/dev/null || true)"
if [ -n "${deployment_id}" ]; then
    export_env DEPLOYMENT_ID "${deployment_id}"
fi

if [ "${create_rc}" -ne 0 ] || [ -z "${deployment_id}" ]; then
    echo "ecctl deployment create failed (exit ${create_rc}); output withheld because it may contain credentials" >&2
    exit "${create_rc:-1}"
fi

es_user="$(jq -r --slurp '[.[] | select(.resources != null) | .resources[] | select(.credentials != null) | .credentials.username][0] // empty' "${result}")"
es_password="$(jq -r --slurp '[.[] | select(.resources != null) | .resources[] | select(.credentials != null) | .credentials.password][0] // empty' "${result}")"
# Mask the password before anything else touches it.
mask "${es_password}"
if [ -z "${es_user}" ] || [ -z "${es_password}" ]; then
    echo "Could not read the deployment credentials from the ecctl output" >&2
    exit 1
fi

# ecctl show output has the endpoints and no credentials; it still goes straight into jq.
kibana_url="$(ecctl deployment show "${deployment_id}" --output json |
    jq -r '.resources.kibana[0].info.metadata | (.service_url // .aliased_url) // empty')"
es_url="$(ecctl deployment show "${deployment_id}" --output json |
    jq -r '.resources.elasticsearch[0].info.metadata | (.service_url // .aliased_url) // empty')"
if [ -z "${kibana_url}" ] || [ -z "${es_url}" ]; then
    echo "Could not read the Kibana/Elasticsearch URLs of the deployment" >&2
    exit 1
fi

# URLs are treated as secrets too: mask the full URL and the bare host, with and without port.
for url in "${kibana_url}" "${es_url}"; do
    bare="${url#https://}"
    mask "${url}"
    mask "${bare}"
    mask "${bare%%:*}"
    mask "${bare%%/*}"
done

kibana_stripped="${kibana_url#https://}"
es_stripped="${es_url#https://}"
test_kibana_url="https://${es_user}:${es_password}@${kibana_stripped}"
test_es_url="https://${es_user}:${es_password}@${es_stripped}"
mask "${test_kibana_url}"
mask "${test_es_url}"

export_env KIBANA_URL "${kibana_url}"
export_env ES_URL "${es_url}"
export_env ES_USER "${es_user}"
export_env ES_PASSWORD "${es_password}"
export_env TEST_KIBANA_URL "${test_kibana_url}"
export_env TEST_ES_URL "${test_es_url}"

echo "Deployment ${DEPLOYMENT_NAME} (${deployment_id}) is ready"
