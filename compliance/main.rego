package main

import data.compliance.cis_k8s
import data.compliance.lib.common

# input is a resource
# data is policy/configuration
# output is findings

resource = input.resource

findings := f {
	# iterate over activated benchmarks
	benchmarks := [key | data.activated_rules[key]]

	# aggregate findings from activated benchmarks
	f := [finding | data.compliance[benchmarks[_]].findings[finding]]
}

metadata = common.metadata
