# This script will stop any Cloudbeat process within an agent and will replace the Cloudbeat artifacts with your local built artifacts

# Prerequisites:
# a. Create the Cloudbeat artifacts using the mage command (for example - "DEV=true PLATFORMS=linux/amd64 SNAPSHOT=true mage -v package")
# b. The running elastic-agents were built with "DEV=TRUE" flag (https://github.com/elastic/security-team/blob/main/docs/cloud-security-posture-team/Onboarding/cloudbeat-ec2.md#:~:text=Build%20the%20agent%20(We%20are%20building%20with%20DEV%3Dtrue%2C%20so%20we%20can%20load%20cloudbeat.tar.gz%20without%20signature%20verification))

#!/bin/bash
ARCH=$(uname -a | rev | cut -d " " -f 1 | rev) # This should return arm64 on m1 and x86_64 on regular mac
PREFIX="/usr/share/elastic-agent/data/"
CLOUDBEAT_VERSION=$(grep defaultBeatVersion ../cmd/version.go | cut -d'=' -f2 | tr -d '" ')
CLOUDBEAT_IMAGE="cloudbeat-${CLOUDBEAT_VERSION}-SNAPSHOT-linux-${ARCH}"
SUFFIX="/install/${CLOUDBEAT_IMAGE}"
DISTRIBUTION_LOCAL_FOLDER="../build/distributions"
DOWNLOAD_PATH="/usr/share/elastic-agent/state/data/downloads/"

PODS=$(kubectl -n kube-system get pod -l app=elastic-agent -o name)
for P in $PODS
do
    POD=$(echo $P | cut -d '/' -f 2)
    FOLDER=$(kubectl -n kube-system exec $POD -- ls $PREFIX)
    # Kill Cloudbeat process
    CLOUDBEAT_PID=$(kubectl -n kube-system exec "$POD" -- pidof cloudbeat)
    kubectl -n kube-system exec "$POD" -- kill -9 "${CLOUDBEAT_PID}"
    # Copy Cloudbeat from your local distribution folder to the agent
    kubectl -n kube-system cp ${DISTRIBUTION_LOCAL_FOLDER}/"${CLOUDBEAT_IMAGE}".tar.gz "$POD":"${DOWNLOAD_PATH}""${CLOUDBEAT_IMAGE}".tar.gz
    kubectl -n kube-system cp ${DISTRIBUTION_LOCAL_FOLDER}/"${CLOUDBEAT_IMAGE}".tar.gz.sha512 "$POD":${DOWNLOAD_PATH}"${CLOUDBEAT_IMAGE}".tar.gz.sha512

    # Delete Cloudbeat artifacts
    INSTALLATION_DIRECTORY=${PREFIX}${FOLDER}"/install"
    kubectl -n kube-system exec "$POD" -- rm -r "${INSTALLATION_DIRECTORY}/${CLOUDBEAT_IMAGE}"
    # Create new Cloudbeat artifcats
    kubectl -n kube-system exec "$POD" -- tar xf "${DOWNLOAD_PATH}""${CLOUDBEAT_IMAGE}".tar.gz -C "${INSTALLATION_DIRECTORY}"
    kubectl -n kube-system exec "$POD" -- chown -R elastic-agent:elastic-agent "${INSTALLATION_DIRECTORY}"
    kubectl -n kube-system exec "$POD" -- chown -R root:root "${INSTALLATION_DIRECTORY}"/"${CLOUDBEAT_IMAGE}"/cloudbeat.yml
    kubectl -n kube-system exec "$POD" -- chmod 0744 "${INSTALLATION_DIRECTORY}"/"${CLOUDBEAT_IMAGE}"/cloudbeat.yml
#    chown -R root:root "${INSTALLATION_DIRECTORY}"/*/*.yml && \
#    chmod 0644 "${INSTALLATION_DIRECTORY}"/*/*.yml && \
#    "${INSTALLATION_DIRECTORY}"/"${CLOUDBEAT_IMAGE}"/cloudbeat -E setup.ilm.enabled=false -E setup.template.enabled=false -E management.enabled
done

# After the script finishes its work, you need to make Cloudbeat run again.
# This can be achieved by:
# a. Open Kibana and press the menu button on the top left of the screen.
# b. Select "Fleet".
# c. Press the "agent policies" tab and select the relevant policy.
# d. Select the "..." next to the CIS Kubernetes benchmark integration and select `edit integration`.
# e. Change the `integration name` and press on `save integration`.
