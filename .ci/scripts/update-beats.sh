#!/bin/bash
set -euxo pipefail

BEATS_VERSION=${1:?Missing version argument}
go get "github.com/elastic/beats/v7@$BEATS_VERSION"
go mod tidy
# updatecli needs some specific stdout when using the shell target,
# see https://www.updatecli.io/docs/plugins/resource/shell/#_shell_target:
# > When the commands runs successfully (e.g. with an exit code of zero), the behavior depends on the content of the
# standard output:
# > - If it is empty, then updatecli report a success with no changes applied.
# > - Otherwise updatecli report a success with the content of the standard output as the resulting value of the change.
git diff
