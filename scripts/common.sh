#!/bin/bash

exec_pod() {
    pod=$1
    cmd=$2
    kubectl -n kube-system exec $pod -- $cmd
}

cp_to_pod() {
    pod=$1
    source=$2
    dest=$3
    kubectl cp $2 kube-system/$1:$dest
}

get_agents() {
    kubectl -n kube-system get pod -l app=elastic-agent -o name
}

find_target_os() {
    _kubectl_node_info operatingSystem
}

find_target_arch() {
    _kubectl_node_info architecture
}

is_eks() {
    _kubectl_node_info kubeletVersion | grep "eks"
}

get_agent_sha() {
    out=$(exec_pod $1 "elastic-agent version --yaml --binary-only")
    echo $out | cut -d ":" -f4 | awk '{$1=$1};1' | awk '{ print substr($0, 0, 6) }'
}

_kubectl_node_info() {
    kubectl get node -o go-template="{{(index .items 0 ).status.nodeInfo.$1}}"
}

# Iterates over the agents, and copies a list of files into each one of them to the `components` folder
copy_to_agents() {
    for P in $(get_agents); do
        POD=$(echo $P | cut -d '/' -f 2)
        SHA=$(get_agent_sha "$POD")
        echo "Found sha=$SHA in pod=$POD"

        DEST=/usr/share/elastic-agent/data/elastic-agent-$SHA/components

        for FILE in "$@"; do
            cp_to_pod "$POD" "$FILE" "$DEST"
        done

        echo "Copied all the assets to $POD"
    done
}

restart_agents() {
    for P in $(get_agents); do
        POD=$(echo $P | cut -d '/' -f 2)
        exec_pod $POD "elastic-agent restart"
    done
}

bump_preview_version() {
    local current_version="${1:?Missing current version}"
    preview_number="${current_version##*-preview}"
    ((next_preview_number = 10#${preview_number} + 1))
    echo "${current_version%-*}-preview$(printf "%02d" $next_preview_number)"
}

bump_minor_version() {
    local version="${1:?Missing version}"
    IFS='.' read -r major minor _ <<<"$1"
    ((minor++))
    echo "$major.$minor.0"
}

get_integration_version() {
    local changelog_path="${1:?Missing changelog.yml path}"
    current_version=$(yq '.[0].version' "$changelog_path" | tr -d '"')
    echo "$current_version"
}

get_new_integration_version_map_entry() {
    local latest_version="${1:?Missing latest versions entry}"
    IFS='-' read -r group1 group2 <<<"$latest_version"
    IFS='.' read -r major1 minor1 _ <<<"$group1"
    IFS='.' read -r major2 minor2 _ <<<"$group2"
    ((minor1++))
    ((minor2++))
    local first_version="${major1}.${minor1}.x"
    local second_version="$(echo "${major2}.${minor2}.x" | xargs)"
    echo "$first_version - $second_version"
}

# bumps existing preview version: 1.0.0-preview01 -> 1.0.0-preview02, or
# creates a new preview version: 1.0.0 -> 1.1.0-preview01, and
# updates the manifest and changelog files
bump_integration_version() {
    changelog_path="${1:?Missing changelog.yml path}"
    manifest_path="${2:?Missing manifest.yml path}"
    pr_url="${3:?Missing PR URL}"
    changelog_description="${4:-Bump version}"
    # exports required for yq's env()
    export changelog_description
    export pr_url
    version="$(get_integration_version "$changelog_path")"
    if [[ $version == *"preview"* ]]; then
        next_version=$(bump_preview_version "$version")
        export next_version
        # update current version and add new changes entry
        yq -i ".[0].version = \"$next_version\"" "$changelog_path"
        yq -i '.[0].changes += [{"description": env(changelog_description), "type": "enhancement", "link": env(pr_url) }]' "$changelog_path"
    else
        next_version="$(bump_minor_version "$version")-preview01"
        export next_version
        # add new version + changes entry
        yq -i '. = [{"version": env(next_version), "changes": [{"description": env(changelog_description), "type": "enhancement", "link": env(pr_url) }]}] + .' "$changelog_path"

        # add new version map for integration - kibana
        latest_entry="$(sed -n '3p' "$changelog_path")"
        next_entry=$(get_new_integration_version_map_entry "$latest_entry")
        sed -i '' -e '3i\'$'\n'"$next_entry" "$changelog_path"

        # update manifest with new kibana version
        IFS='-' read -r _ next_kibana_version <<<"$next_entry"
        IFS='.' read -r major minor _ _ <<<"$(echo "$next_kibana_version" | xargs)"
        yq -i ".conditions.kibana.version = \"^$major.$minor.0\"" "$manifest_path"
    fi
    yq -i ".version = \"$next_version\"" "$manifest_path"
}
