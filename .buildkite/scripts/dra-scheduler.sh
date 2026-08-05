#!/usr/bin/env bash
# Weekly DRA scheduler: derive maintained branches from .mergify.yml + main,
# then upload Buildkite trigger steps for the cloudbeat DRA pipeline.
#
# Optional env:
#   EXCLUDE_BRANCHES — comma-separated branches to skip (e.g. "9.3")
#   SKIP_UPLOAD      — if "true", print generated steps and exit (local dry-run)
#   SKIP_REMOTE_CHECK — if "true", do not verify branches exist on origin
#   MERGIFY_FILE     — path to Mergify config (default: .mergify.yml)
#   PIPELINE_TO_TRIGGER — Buildkite pipeline slug (default: cloudbeat)

set -euo pipefail

MERGIFY_FILE="${MERGIFY_FILE:-.mergify.yml}"
PIPELINE_TO_TRIGGER="${PIPELINE_TO_TRIGGER:-cloudbeat}"
SKIP_UPLOAD="${SKIP_UPLOAD:-false}"
SKIP_REMOTE_CHECK="${SKIP_REMOTE_CHECK:-false}"
EXCLUDE_BRANCHES="${EXCLUDE_BRANCHES:-}"

install_yq() {
    if command -v yq >/dev/null 2>&1 && yq --version 2>/dev/null | grep -q mikefarah; then
        return
    fi
    echo "--- Downloading yq"
    local yq_bin="/usr/local/bin/yq"
    if [[ ! -w "$(dirname "${yq_bin}")" ]]; then
        yq_bin="${TMPDIR:-/tmp}/yq"
    fi
    # Prefer the host arch when downloading outside Linux CI.
    local yq_asset="yq_linux_amd64"
    case "$(uname -s)-$(uname -m)" in
    Darwin-arm64) yq_asset="yq_darwin_arm64" ;;
    Darwin-x86_64) yq_asset="yq_darwin_amd64" ;;
    Linux-aarch64 | Linux-arm64) yq_asset="yq_linux_arm64" ;;
    esac
    curl -fsSL --retry-max-time 60 --retry 3 --retry-delay 5 \
        -o "${yq_bin}" \
        "https://github.com/mikefarah/yq/releases/latest/download/${yq_asset}"
    chmod a+x "${yq_bin}"
    PATH="$(dirname "${yq_bin}"):${PATH}"
    export PATH
}

is_excluded() {
    local branch="$1"
    local excl
    [[ -z "${EXCLUDE_BRANCHES}" ]] && return 1
    local IFS=','
    # shellcheck disable=SC2086
    for excl in ${EXCLUDE_BRANCHES}; do
        if [[ "${branch}" == "${excl}" ]]; then
            return 0
        fi
    done
    return 1
}

already_listed() {
    local needle="$1"
    local b
    for b in "${BRANCHES[@]+"${BRANCHES[@]}"}"; do
        if [[ "${b}" == "${needle}" ]]; then
            return 0
        fi
    done
    return 1
}

branch_exists_on_origin() {
    local branch="$1"
    git ls-remote --exit-code --heads origin "${branch}" >/dev/null 2>&1
}

add_branch() {
    local branch="$1"
    if is_excluded "${branch}"; then
        echo "Skipping excluded branch: ${branch}"
        return
    fi
    if already_listed "${branch}"; then
        return
    fi
    BRANCHES+=("${branch}")
}

if [[ ! -f "${MERGIFY_FILE}" ]]; then
    echo "^^^ +++"
    echo "ERROR: Mergify file not found at ${MERGIFY_FILE}"
    exit 1
fi

install_yq

echo "--- Deriving maintained branches from ${MERGIFY_FILE}"

BRANCHES=()
add_branch "main"

# Collect unique X.Y destinations under actions.backport.branches (not "main").
while IFS= read -r branch; do
    [[ -z "${branch}" ]] && continue
    add_branch "${branch}"
done < <(
    yq -r '
    .pull_request_rules[]
    | select(.actions.backport.branches != null)
    | .actions.backport.branches[]
  ' "${MERGIFY_FILE}" |
        grep -E '^[0-9]+\.[0-9]+$' |
        sort -Vu
)

TARGET_BRANCHES=()
for branch in "${BRANCHES[@]+"${BRANCHES[@]}"}"; do
    if [[ "${SKIP_REMOTE_CHECK}" == "true" ]]; then
        TARGET_BRANCHES+=("${branch}")
        continue
    fi
    if branch_exists_on_origin "${branch}"; then
        TARGET_BRANCHES+=("${branch}")
    else
        echo "WARNING: branch ${branch} is listed in Mergify but missing on origin — skipping"
    fi
done

if [[ ${#TARGET_BRANCHES[@]} -eq 0 ]]; then
    echo "^^^ +++"
    echo "ERROR: No target branches to trigger. Check ${MERGIFY_FILE} and EXCLUDE_BRANCHES=${EXCLUDE_BRANCHES}"
    exit 1
fi

echo "Target branches: ${TARGET_BRANCHES[*]}"

STEPS_FILE="$(mktemp)"
trap 'rm -f "${STEPS_FILE}"' EXIT

{
    echo "# yaml-language-server: \$schema=https://raw.githubusercontent.com/buildkite/pipeline-schema/main/schema.json"
    echo "steps:"
    for branch in "${TARGET_BRANCHES[@]}"; do
        cat <<EOF
  - trigger: ${PIPELINE_TO_TRIGGER}
    label: ":rocket: DRA refresh / ${branch}"
    async: true
    build:
      branch: "${branch}"
      message: "Weekly DRA refresh (${branch}) — prevent snapshot expiry"
EOF
    done
} >"${STEPS_FILE}"

echo "--- Generated pipeline steps"
cat "${STEPS_FILE}"

if [[ "${SKIP_UPLOAD}" == "true" ]]; then
    echo "SKIP_UPLOAD=true — not uploading to Buildkite"
    exit 0
fi

echo "--- Uploading steps to Buildkite"
buildkite-agent pipeline upload "${STEPS_FILE}"
