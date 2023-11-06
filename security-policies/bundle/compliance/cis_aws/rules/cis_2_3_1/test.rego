package compliance.cis_aws.rules.cis_2_3_1

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as rule_input(false)
}

test_pass {
	eval_pass with input as rule_input(true)
}

test_not_evaluated {
	not_eval with input as test_data.not_evaluated_rds_db_instance
}

rule_input(encryption_enabled) = test_data.generate_rds_db_instance(encryption_enabled, true, false, [])

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
