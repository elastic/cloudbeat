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
	f := {finding |
		benchmark := benchmarks[_]
		rule := input.activated_rules[benchmark][_]
		result := compliance[benchmark].rules[rule].finding with data.benchmark_data_adapter as compliance[benchmark].data_adapter
		finding = {
			"result": result,
			"rule": compliance[benchmark].rules[rule].metadata,
		}
	}
}

findings = f {
	not input.activated_rules

	# aggregate findings from all benchmarks
	f := {finding |
		result := compliance[benchmark].rules[rule].finding with data.benchmark_data_adapter as compliance[benchmark].data_adapter
		finding = {
			"result": result,
			"rule": compliance[benchmark].rules[rule].metadata,
		}
	}
}

metadata = common.metadata
