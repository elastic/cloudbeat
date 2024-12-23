package compliance.cis_aws.rules.cis_2_3_3

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	# Publicly available with an exposed subnet
	eval_fail with input as rule_input(true, [test_data.generate_rds_db_instance_subnet_with_route("0.0.0.0/0", "igw-12345678")])

	# Publicly availble with multiple subnets, only one of them exposed
	eval_fail with input as rule_input(true, [test_data.generate_rds_db_instance_subnet_with_route("10.1.0.0/16", "nat-12345678"), test_data.generate_rds_db_instance_subnet_with_route("0.0.0.0/0", "igw-12345678")])
}

test_pass if {
	# Publicly accessible, no subnets
	eval_pass with input as rule_input(true, [])

	# Publicly accessible, not an internet gateway
	eval_pass with input as rule_input(true, [test_data.generate_rds_db_instance_subnet_with_route("0.0.0.0/0", "nat-12345678")])

	# Publicly accessible, destination not 0.0.0.0
	eval_pass with input as rule_input(true, [test_data.generate_rds_db_instance_subnet_with_route("10.1.0.0/16", "igw-12345678")])

	# Publicly accessible, one subnet with internet gateway, one subnet with destination 0.0.0.0
	eval_pass with input as rule_input(true, [test_data.generate_rds_db_instance_subnet_with_route("0.0.0.0/0", "nat-12345678"), test_data.generate_rds_db_instance_subnet_with_route("10.1.0.0/16", "igw-12345678")])

	# Not publicly accessible, subnet is exposed
	eval_pass with input as rule_input(false, [test_data.generate_rds_db_instance_subnet_with_route("0.0.0.0/0", "igw-12345678")])
}

test_not_evaluated if {
	not_eval with input as test_data.not_evaluated_rds_db_instance

	# An RDS db instance with a null route table in one of the subnets
	not_eval with input as rule_input(true, [test_data.generate_rds_db_instance_subnet_with_route("0.0.0.0/0", "igw-12345678"), {"ID": "subnet-abcdef12", "RouteTable": null}])
}

rule_input(publicly_accessible, subnets) := test_data.generate_rds_db_instance(true, true, publicly_accessible, subnets)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
