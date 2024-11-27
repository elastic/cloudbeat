# Buildkite

This README provides an overview of the Buildkite pipeline used to automate the build and publish process for Cloudbeat artifacts.

## Artifacts

The pipeline generates the following artifacts:

- **dependencies-CLOUDBEAT_VERSION-WORKFLOW.csv**: This CSV file contains a list of dependencies for the specific Cloudbeat version being built. It helps track build dependencies.

- **cloudbeat-CLOUDBEAT_VERSION-WORKFLOW-linux-ARCH.tar.gz**: This tarball includes the Cloudbeat binary and its corresponding csp-policies archive. The supported architectures for the artifacts are amd64 and arm64.

## Triggering the Pipeline

The pipeline is triggered in the following scenarios:

- **Snapshot Builds**: A snapshot build is triggered when a pull request (PR) is merged into the 'main' branch or a version-specific branch. Additionally, if the environment variable RUN_RELEASE is set to "true", a snapshot build is also triggered.

- **Staging Builds**: A staging build is triggered when a PR is merged into a version-specific branch or when the environment variable RUN_RELEASE is set to "true". Staging builds are typically used for a release build candidate.

After a successful build, the pipeline publishes the generated artifacts to the Google Cloud Storage (GCS) bucket named [elastic-artifacts-snapshot/cloudbeat](https://console.cloud.google.com/storage/browser/elastic-artifacts-snapshot/cloudbeat). You can access the published artifacts in this bucket.

## Pipeline Configuration

To view the pipeline and its configuration, click [here](https://buildkite.com/elastic/cloudbeat).

## Notifications

The pipeline is [configured](https://buildkite.com/organizations/elastic/services/68636/edit) to send Slack notifications to the `#cloud-sec-ci` channel. Additionally, it includes a custom notification [script](./scripts/notify.sh) that pings specific users in the event of a build failure.
