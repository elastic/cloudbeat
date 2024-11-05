#!/bin/bash
set -euxo pipefail

# Delete password, stop and disable ssh
sudo passwd --delete cloudbeat &
(
    sudo systemctl disable --now ssh || true
    sudo systemctl mask ssh.service || true
    sudo killall sshd || true
) &

usage="$0 <elastic agent version> <elastic artifact server> <fleet url> <enrollment token>"
ElasticAgentVersion=${1:?$usage}
ElasticArtifactServer=${2:?$usage}
FleetUrl=${3:?$usage}
EnrollmentToken=${4:?$usage}

ElasticAgentArtifact="elastic-agent-$ElasticAgentVersion-linux-x86_64"
curl -L -O "${ElasticArtifactServer}/$ElasticAgentArtifact.tar.gz"
tar xzf "${ElasticAgentArtifact}.tar.gz"
cd "${ElasticAgentArtifact}"
sudo ./elastic-agent install --non-interactive --url="${FleetUrl}" --enrollment-token="${EnrollmentToken}"

wait
