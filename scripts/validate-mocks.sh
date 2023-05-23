#!/bin/sh

## validate that all mocks exists and up to date
## the script will run the "mockery" command and we dont expect any changes

mock="mockery --dir . --inpackage --all --with-expecter --case underscore --recursive"
eval $mock > /dev/null 2>&1
mage AddLicenseHeaders > /dev/null 2>&1

diff="git diff --name-only | grep \".*/mock_.*.go\""
diff_count=$(eval $diff | wc -l)


untracked="git ls-files --others --exclude-standard | grep \".*/mock_.*.go\""
untracked_count=$(eval $untracked | wc -l)

if [ $diff_count -eq 0 ] && [ $untracked_count -eq 0 ]; then
    exit 0
fi


echo "There are missing mocks for interfaces, please run \"$mock\" to update all mocks"
echo "Mock files:"
eval $diff
eval $untracked
exit 1
