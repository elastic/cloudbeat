# Kustomize Vanilla Agent

This manifests are used to deploy the agent on local kind cluster with dev image of the agent.
So that later we can inject into that agent a custom binray of cloudbeat and run E2E flow.

## How to use it

### Build the Agent
Note: this step might take a while
**From cloudbeat repo**
```bash
make PackageAgent
```

### Setup ERP

**From cloudbeat repo**
#### Step 1
Start your local ERP stack
```bash
elastic-package stack up -vd --version=8.5.0-SNAPSHOT
just create-kind-cluster
just elastic-stack-connect-kind
kind load docker-image elastic-agent:8.5.0-SNAPSHOT --name kind-mono
```

#### Step 2 - Get the entrollment token
To find the envorllment token, `app/fleet/enrollment-tokens` copy the token of the `Elastic-Agent (elastic-package)`
Run `export FLEET_ENROLLMENT_TOKEN=$TOKEN`

#### Step 3 - Take care of SSL
The SSL certificate was created by `elastic-package` and stored in `ELASTIC_PACKAGE_CA_CERT`.
Run 
```bash
eval "$(elastic-package stack shellinit)"
cp $ELASTIC_PACKAGE_CA_CERT deploy/kustomize/overlays/cloudbeat-vanilla-agent
```
#### Step 4 - Complete ERP setup
```bash
kubectl apply -k deploy/kustomize/overlays/cloudbeat-vanilla-agent
```

#### Step 4 - Verify
Go to `app/fleet/agents` and check that the new agent (`kind-mono-control-plane`) is healthy

### Load custom cloudbeat binary

**From cloudbeat repo**

To use custom binray of cloudbeat you need
1. Build binray + checksum
2. Copy the files to agent pod
3. Restart the cloudbeat process in the pod

#### Step 1 - build cloudbeat
```bash
DEV=true PLATFORMS=linux/arm64 SNAPSHOT=true mage -v package
```

#### Step 2 - Copy and restart
To copy all the assets and restart https://github.com/elastic/cloudbeat/blob/edfb7cad16eb7477a853f97c9a3d3cb906f5f6fb/scripts/remote_replace_cloudbeat.sh `scripts/remote_replace_cloudbeat.sh`
```bash
cd scripts
./remote_replace_cloudbeat.sh
```