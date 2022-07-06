# This script will stop any Cloudbeat process within an agent and will replace the Cloudbeat artifacts with your local version
# Prerequisites:
# a. Create the Cloudbeat artifacts using the mage command (for example - "DEV=true PLATFORMS=linux/amd64 SNAPSHOT=true mage -v package")
# b. The running elastic agents must be complied with "DEV=TRUE"

#!/bin/bash
ARCH=`uname -a | rev | cut -d " " -f 1 | rev` # This should return arm64 on m1 and x86_64 on regular mac
PREFIX="/usr/share/elastic-agent/data/"
SUFFIX="/install/cloudbeat-8.4.0-SNAPSHOT-linux-${ARCH}"
DISTRIBUTION_LOCAL_FOLDER="../build/distributions"
CLOUDBEAT_IMAGE="cloudbeat-8.4.0-SNAPSHOT-linux-$ARCH"

PODS=`kubectl -n kube-system get pod -l app=elastic-agent -o name`
for P in $PODS
do
    POD=`echo $P | cut -d '/' -f 2`
    FOLDER=`kubectl -n kube-system exec $POD -- ls $PREFIX`
    # Kill Cloudbeat process
    CLOUDBEAT_PID=`kubectl -n kube-system exec "$POD" -- pidof cloudbeat`
    kubectl -n kube-system exec "$POD" -- kill -9 "${CLOUDBEAT_PID}"
    # Copy Cloudbeat from your local distribution folder to the agent
    kubectl -n kube-system cp ${DISTRIBUTION_LOCAL_FOLDER}/"${CLOUDBEAT_IMAGE}".tar.gz "$POD":/usr/share/elastic-agent/state/data/downloads/"${CLOUDBEAT_IMAGE}".tar.gz
    kubectl -n kube-system cp ${DISTRIBUTION_LOCAL_FOLDER}/"${CLOUDBEAT_IMAGE}".tar.gz.sha512 "$POD":/usr/share/elastic-agent/state/data/downloads/"${CLOUDBEAT_IMAGE}".tar.gz.sha512
    # Delete Cloudbeat artifacts
    kubectl -n kube-system exec "$POD" -- rm -r "${PREFIX}${FOLDER}${SUFFIX}"
done

# After the script finishes its work, you need to make Cloudbeat run again.
# This can be achieved by:
# a. Press the menu button on the top left of the screen
# b. Press on "Fleet"
# c. Press the "agent policies" tab and select the relevant policy
# d. Press the "..." next to the CIS Kubernetes benchmark integration and select `edit integration`
# e. Change the `integration name` and press on `save integration`
