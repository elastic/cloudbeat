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
>>>>>>> cc1ed67b ([Asset Inventory][Azure] Do not add --install-servers flag for versions < 9.0 (#3142))

wait
