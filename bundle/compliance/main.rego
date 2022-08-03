package main

import data.compliance
import data.compliance.lib.common

# input contains the resource and the configuration
# output is findings

resource = input.resource

findings = f {
	input.activated_rules

	# iterate over activated benchmarks
	benchmarks := [key | input.activated_rules[key]]

	# aggregate findings from activated benchmarks
	f := {finding | compliance[benchmarks[_]].findings[finding]}
}

findings = f {
	not input.activated_rules

	# aggregate findings from all benchmarks
	f := {finding | compliance[benchmarks].findings[finding]}
}

metadata = common.metadata
