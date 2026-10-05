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

- **Weekly DRA refresh**: The [cloudbeat-dra-scheduler](https://buildkite.com/elastic/cloudbeat-dra-scheduler) pipeline runs every Monday at 06:00 UTC. It derives maintained branches from `.mergify.yml` backport destinations plus `main`, then triggers a DRA build of this pipeline for each branch. This keeps snapshot manifests in GCS fresh within the 15-day retention window even when a branch has no commits.

After a successful build, the pipeline publishes the generated artifacts to the Google Cloud Storage (GCS) bucket named [elastic-artifacts-snapshot/cloudbeat](https://console.cloud.google.com/storage/browser/elastic-artifacts-snapshot/cloudbeat). You can access the published artifacts in this bucket.

### Weekly scheduler maintenance

- **Branch source of truth**: `main` + unique `X.Y` values under Mergify `actions.backport.branches` (not Elastic `branches.json` / future-releases, which can lag team-maintained minors such as `9.3`).
- **New minor**: when the version-bump pipeline adds a Mergify backport rule, the next weekly run includes that branch automatically — no `catalog-info.yaml` schedule edits.
- **EOL**: remove the Mergify backport rule for that branch; the scheduler stops triggering it.
- **Manual run**: start a new build of [cloudbeat-dra-scheduler](https://buildkite.com/elastic/cloudbeat-dra-scheduler) on `main`.
- **Verify**: confirm child cloudbeat builds pass Snapshot publish, then check [dra-info](https://artifacts-staging.elastic.co/dra-info/index.html) and `https://storage.googleapis.com/elastic-artifacts-snapshot/cloudbeat/latest/<branch>.json`.
- **Failures**: the scheduler notifies `#cloud-sec-ci` when the scheduler build itself fails. Child `cloudbeat` DRA builds keep their existing failure notifications.

Optional env vars for the scheduler script (`.buildkite/scripts/dra-scheduler.sh`): `EXCLUDE_BRANCHES` (CSV, whitespace around values is ignored), `SKIP_UPLOAD=true` (dry-run), `SKIP_REMOTE_CHECK=true`, `YQ_VERSION` (pinned mikefarah/yq tag), `YQ_SHA256` (optional expected binary hash; when unset, verified against the release checksums file).

## Pipeline Configuration

To view the DRA pipeline and its configuration, click [here](https://buildkite.com/elastic/cloudbeat). Scheduler pipeline: [cloudbeat-dra-scheduler](https://buildkite.com/elastic/cloudbeat-dra-scheduler).

## Notifications

The pipeline is [configured](https://buildkite.com/organizations/elastic/services/68636/edit) to send Slack notifications to the `#cloud-sec-ci` channel. Additionally, it includes a custom notification [script](./scripts/notify.sh) that pings specific users in the event of a build failure.
