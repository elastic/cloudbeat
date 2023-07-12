```







1. THE USER INSTALLS THE CSPM INTEGRATION

2. KIBANA GENERATES A COMMAND FOR THE USER TO COPY

3. THE USER IS COPYING THE COMMAND

4. THE USER IS PRESSING THE LINK TO THE CLOUD SHELL

5. THE USER IS RUNNING THE COMMAND IN GCP's CLOUD SHELL







```




```
gcloud deployment-manager deployments create --automatic-rollback-on-error uri-test --template compute-engine.py --properties zone:europe-west2-a,elasticAgentVersion:8.8.0,fleetUrl:https://0bcfe3aec94240f0ab3731e4f007daf0.fleet.us-central1.gcp.foundit.no:443,enrollmentToken:OVRpbFNZa0JOekpSZTBsRDhXYWw6MjdsYzdRRXVRNmFxUzFhUkl5X1Mtdw==
```




[![Open in Cloud Shell](https://gstatic.com/cloudssh/images/open-btn.svg)](https://shell.cloud.google.com/cloudshell/editor?cloudshell_git_repo=https%3A%2F%2Fgithub.com%2Furi-weisman%2Fcloudbeat&cloudshell_git_branch=deployment_manager&cloudshell_print=instructions.txt&cloudshell_workspace=deploy%2Fdeployment-manager&show=terminal)

```
