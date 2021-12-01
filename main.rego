package main

import data.compliance.cis_k8s
import data.compliance.lib.common

# input is a resource
# data is policy/configuration
# output is findings

resource = input

findings = cis_k8s.findings

metadata = common.metadata
