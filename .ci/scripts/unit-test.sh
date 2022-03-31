#!/usr/bin/env bash
set -exo pipefail

source ./scripts/make/common.bash

jenkins_setup

make update
mage goTestUnit
