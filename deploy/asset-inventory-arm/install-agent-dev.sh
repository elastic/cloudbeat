#!/bin/bash
set -euxo pipefail

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
