# Cloudbeat Minor Version Bump — Team Checklist

Reference for the team when a minor Feature Freeze (FF) triggers the version-bump pipeline. Grounded in what actually happened during the `9.5.0` bump in July 2026 — both what worked and what tripped us up.

## T-2 to T-1 days (pre-FF prep)

Nothing on this list requires code changes if the scripts are up to date, but these prevent every issue we hit on 7 Jul.

- [ ] **Create the new backport label** in `elastic/cloudbeat` before FF day
  ```bash
  gh label create "backport-v${NEW_MINOR}.0" --repo elastic/cloudbeat --color 8af5e5
  ```
  Without this, Mergify can't apply the backport rule the bump PR adds.
- [ ] **Check for stale bump branches** and delete any left over from previously-closed (unmerged) attempts:
  ```bash
  gh api repos/elastic/cloudbeat/git/refs/heads | jq -r '.[].ref' | grep bump-to-
  # Delete only branches whose PR is closed (not merged):
  gh api -X DELETE repos/elastic/cloudbeat/git/refs/heads/bump-to-X.Y.Z
  ```
  Focus on `bump-to-${NEW_MINOR}.0` and `bump-to-${NEXT_MINOR_AFTER_THAT}.0`. Scripts detect this and fail clearly (per [#7227](https://github.com/elastic/cloudbeat/pull/7227)), but pre-deleting saves a pipeline round-trip.
- [ ] Confirm `cloudsecmachine` bot has **Write+** on the repo (ping platform if unsure).
- [ ] Watch `#mission-control` for the FF-date announcement and any last-minute changes to `NEW_VERSION`/`WORKFLOW` semantics.

## FF day — pipeline runs automatically

The trigger comes from release-eng's `unified-release-centralized-version-bump` pipeline at ~02:00 UTC, which invokes `cloudbeat-version-bump` with:

```
WORKFLOW=minor  BRANCH=9.5  NEW_VERSION=9.5.0
```

Watch `#security-cloud-services-team` for pipeline notifications, and monitor [cloudbeat-version-bump](https://buildkite.com/elastic/cloudbeat-version-bump) directly.

### Step 1 — `bump-minor.sh` runs (auto)

Creates `bump-to-${NEW_MINOR}.0` branch and opens a PR that adds a Mergify backport rule for the new branch.

- [ ] Review the auto-generated bump PR (title: `Bump cloudbeat version to X.Y.0`)
- [ ] Click **Merge when ready** to add it to the queue (do **not** tick the "bypass rules" box)
- [ ] Back in Buildkite: click **Unblock** on the `Wait for PR merge` block step

### Step 2 — `post-minor-merge.sh` runs (auto)

Cuts the new release branch from main, then opens a second PR bumping main to the next minor (`X.(Y+1).0`).

- [ ] Verify the new `X.Y` release branch exists on origin
  ```bash
  git ls-remote --heads https://github.com/elastic/cloudbeat.git ${NEW_MINOR}
  ```
- [ ] Merge the second bump PR (title: `Bump cloudbeat version to X.(Y+1).0`) via **Merge when ready**

### Step 3 — DRA artifact polling

- [ ] Click **Unblock** on the `Ready to poll for DRA artifacts?` block step in Buildkite
- [ ] Pipeline then waits (up to ~4h) for:
  - `X.Y.0` staging artifact
  - `X.Y.0-SNAPSHOT` snapshot artifact
- [ ] Pipeline completes automatically once both are published

## T+1 and beyond (related follow-ups)

- [ ] Review the **integrations pre-release bump PR** on `elastic/integrations` when it appears — advances the integration's min-stack version to the new minor. Usually straightforward; may or may not need a coordinated cloudbeat-side change.
- [ ] Confirm `bin/hermit.hcl` gets synced automatically by [`scripts/sync_internal_cloudbeat_version.sh`](../../scripts/sync_internal_cloudbeat_version.sh) (runs on a daily cron via `sync-internal-cloudbeat-version.yml`, gated on the snapshot actually being published — no manual action needed).
- [ ] Watch `#mission-control` for `project-configs/${NEW_MINOR}` missing-config alerts from release-eng and any cross-team requests that impact us.

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `GH006: Protected branch update failed for refs/heads/bump-to-X.Y.Z` | Stale branch from a previously-closed unmerged PR | Delete via `gh api -X DELETE repos/elastic/cloudbeat/git/refs/heads/bump-to-X.Y.Z`, then re-run the failed Buildkite job. Scripts flag this clearly since [#7227](https://github.com/elastic/cloudbeat/pull/7227). |
| `Rule: auto-merge version bump PRs — GitHub refused to merge` | Was caused by Mergify's `merge:` action bypassing the merge queue | Fixed in [#7228](https://github.com/elastic/cloudbeat/pull/7228). If you see this again, check that no `merge:` action has been re-added to `.mergify.yml` for version-bump PRs. |
| Pipeline's bot rejected on push with `Changes must be made through a pull request` | Bot identity lacks Write+ on the repo | Ping platform (@gurevichdmitry) to bump permissions. |
| Bump PR sitting unmerged, blocking pipeline | Merge queue enforced on main; needs a human | Click **Merge when ready** (do NOT tick "bypass rules"). |
| `NEW_VERSION` seems wrong | Release-eng convention: `NEW_VERSION` is the version for the **new branch**, not for main. Main gets `next_minor_version(NEW_VERSION)` via the second bump PR. | Expected — the scripts handle the semver math. |

## Reference PRs from `9.5.0` (July 2026)

Immediate deliverables:

- Mergify rule for 9.5 backports: [#7225](https://github.com/elastic/cloudbeat/pull/7225)
- Main → 9.6.0 bump: [#7268](https://github.com/elastic/cloudbeat/pull/7268)

Pipeline enhancements shipped that cycle:

- [#7184](https://github.com/elastic/cloudbeat/pull/7184) — idempotency on retrigger
- [#7186](https://github.com/elastic/cloudbeat/pull/7186) — added `bump_main_to_next_minor` step
- [#7227](https://github.com/elastic/cloudbeat/pull/7227) — fail clearly on stale bump branches
- [#7228](https://github.com/elastic/cloudbeat/pull/7228) — removed Mergify auto-merge (conflicted with merge queue)

## Related scripts

- [`.buildkite/version-bump-pipeline.yml`](../../.buildkite/version-bump-pipeline.yml) — pipeline definition
- [`.buildkite/scripts/release/bump-minor.sh`](../../.buildkite/scripts/release/bump-minor.sh) — Step 1 entry point
- [`.buildkite/scripts/release/post-minor-merge.sh`](../../.buildkite/scripts/release/post-minor-merge.sh) — Step 2 entry point
- [`.buildkite/scripts/release/common.sh`](../../.buildkite/scripts/release/common.sh) — shared helpers
- [`scripts/sync_internal_cloudbeat_version.sh`](../../scripts/sync_internal_cloudbeat_version.sh) — daily hermit.hcl sync
