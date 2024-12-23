package compliance.cis_aws.rules.cis_3_3

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

forbidden_principal1 := "http://acs.amazonaws.com/groups/global/AllUsers"

forbidden_principal2 := "http://acs.amazonaws.com/groups/global/AuthenticatedUsers"

test_violation if {
	# Fail with forbidden principal
	eval_fail with input as rule_input(forbidden_principal1, [test_data.generate_s3_bucket_policy_statement("Deny", "*", "s3:*", "true")])
	eval_fail with input as rule_input(forbidden_principal2, [test_data.generate_s3_bucket_policy_statement("Deny", "*", "s3:*", "true")])

	# Fail with forbidden policy statement
	eval_fail with input as rule_input("", [test_data.generate_s3_bucket_policy_statement("Allow", {"AWS": "*"}, "s3:*", "true")])
	eval_fail with input as rule_input("", [test_data.generate_s3_bucket_policy_statement("Allow", "*", "s3:*", "true")])

	# Fail with multiple policy statements, one of them forbidden
	eval_fail with input as rule_input("", [test_data.generate_s3_bucket_policy_statement("Allow", "*", "s3:*", "true"), test_data.generate_s3_bucket_policy_statement("Deny", "blabla", "s3:*", "true")])

	# Fail with both forbidden principal and forbidden policy statement
	eval_fail with input as rule_input(forbidden_principal1, [test_data.generate_s3_bucket_policy_statement("Allow", "*", "s3:*", "true")])
}

test_pass if {
	eval_pass with input as rule_input("allowed-test-principal", [test_data.generate_s3_bucket_policy_statement("Deny", `{"AWS":"111122223333"}`, "s3:*", "true")])
	eval_pass with input as rule_input(null, [test_data.generate_s3_bucket_policy_statement("Deny", `{"AWS":["arn:aws:iam::111122223333:role/JohnDoe"]}`, "s3:*", "true")])
	eval_pass with input as rule_input("", [test_data.generate_s3_bucket_policy_statement("Deny", "*", "s3:*", "true")])
	eval_pass with input as rule_input("", [
		test_data.generate_s3_bucket_policy_statement("Allow", {"AWS": "arn:aws:iam::111122223333:root"}, "s3:*", "true"),
		test_data.generate_s3_bucket_policy_statement("Deny", "*", "s3:*", "true"),
		test_data.generate_s3_bucket_policy_statement("Deny", {"AWS": "*"}, "s3:*", "true"),
	])
}

test_not_evaluated if {
	not_eval with input as test_data.not_evaluated_trail
}

rule_input(principal_uri, policy_statement) := test_data.generate_trail_bucket_info(principal_uri, policy_statement)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
