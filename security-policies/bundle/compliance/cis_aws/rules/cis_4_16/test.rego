package compliance.cis_aws.rules.cis_4_16

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	# no enalbed field
	eval_fail with input as rule_input({})
	eval_fail with input as rule_input({"Enabled": false})
}

test_pass if {
	eval_pass with input as rule_input({"Enabled": true})
}

rule_input(entry) := test_data.generate_securityhub(entry)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
