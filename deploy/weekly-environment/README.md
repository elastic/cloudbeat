# Weekly environment

**Motivation**

Create a long-running environment to monitor the latest version of our KSPM integration.
This will allow us to detect any regression that may happen in a long-running environment.

**How to Use**

The weekly environment is deployed automatically every two weeks on Monday at 00:00 UTC.
The environment will be destroyed automatically after 2 weeks by the Elastic staging environment cleanup job.

The job will deploy an Elastic Cloud environment in staging, while using the latest Elastic Agent SNAPSHOT version with the latest KSPM integration.
It will use a pre-defined EC2 instance to run the KSPM integration on a vanilla Kubernetes cluster.

The job will also create the relevant benchmark alerts and connectors in Kibana, and will trigger slack alerts in case of wrong benchmark results.

**How to Deploy Manually**

You can deploy the job manually by running the flow as a Github dispatch workflow.
In order to do it, you need to have the right permissions to run the workflow.
You can find the workflow [here](https://github.com/elastic/cloudbeat/actions/workflows/weekly-enviroment.yml).

** In this case, please delete the old environment before deploying a new one.

**Next Steps**
- Support EKS KSPM integration.
