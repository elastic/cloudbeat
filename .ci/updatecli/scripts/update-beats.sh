#!/bin/bash
set -euxo pipefail

BEATS_VERSION=${1:?Missing version argument}
go get "github.com/elastic/beats/v7@$BEATS_VERSION"
go mod tidy
# updatecli needs standard output to not be empty for changes to be detected
# see https://www.updatecli.io/docs/plugins/resource/shell/#_shell_target:
# > When the commands runs successfully (e.g. with an exit code of zero), the behavior depends on the content of the
# standard output:
# > - If it is empty, then updatecli report a success with no changes applied.
# > - Otherwise updatecli report a success with the content of the standard output as the resulting value of the change.
# Non-empty stdout will trigger the "git add" stage which will commit all differences to the updatecli branch.
# git diff is a safe choice because it will be non-empty when changes need to be committed and it's also good for
# debugging.
git diff
# Re-generate config files, ignore failure
mage config || true
