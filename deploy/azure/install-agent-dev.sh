#!/bin/bash
set -euxo pipefail

usage="$0 <elastic agent version> <elastic artifact server> <fleet url> <enrollment token>"
ElasticAgentVersion=${1:?$usage}
ElasticArtifactServer=${2:?$usage}
FleetUrl=${3:?$usage}
EnrollmentToken=${4:?$usage}

install_servers_opt=""
major_version=$(echo "${ElasticAgentVersion}" | cut -d'.' -f1)
if [ "$major_version" -ge 9 ]; then
    install_servers_opt="--install-servers"
fi

ElasticAgentArtifact="elastic-agent-$ElasticAgentVersion-linux-x86_64"
curl -L -O "${ElasticArtifactServer}/$ElasticAgentArtifact.tar.gz"
tar xzf "${ElasticAgentArtifact}.tar.gz"
cd "${ElasticAgentArtifact}"
<<<<<<< HEAD
sudo ./elastic-agent install --non-interactive --url="${FleetUrl}" --enrollment-token="${EnrollmentToken}"
=======
sudo ./elastic-agent install --non-interactive --url="${FleetUrl}" --enrollment-token="${EnrollmentToken}" ${install_servers_opt:+$install_servers_opt}
>>>>>>> 83197406 (Update install-agent.sh script (#3023))

wait
