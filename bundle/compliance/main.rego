package main

import data.compliance
import data.compliance.lib.common

# input is a resource
# data is policy/configuration
# output is findings

resource = input.resource

findings = f {
	data.activated_rules

	# iterate over activated benchmarks
	benchmarks := [key | data.activated_rules[key]]

	# aggregate findings from activated benchmarks
	f := {finding | compliance[benchmarks[_]].findings[finding]}
}

findings = f {
	not data.activated_rules

	# aggregate findings from all benchmarks
	f := {finding | compliance[benchmarks].findings[finding]}
}

metadata = common.metadata
