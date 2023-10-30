package compliance.cis_aws.rules.cis_3_10

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test

s3_object_type = "AWS::S3::Object"

not_s3_object_type = "AWS::S3ObjectLambda::AccessPoint"

test_violation {
	eval_fail with input as rule_input(null, true)
	eval_fail with input as rule_input([], true)
	eval_fail with input as rule_input([{"ReadWriteType": "ReadOnly"}], true)
	eval_fail with input as rule_input([{"ReadWriteType": "ReadOnly", "DataResources": [{"Type": not_s3_object_type}]}], true)
	eval_fail with input as rule_input([{"ReadWriteType": "WriteOnly", "DataResources": [{"Type": not_s3_object_type}]}], true)
	eval_fail with input as rule_input([{"ReadWriteType": "ReadOnly", "DataResources": [{"Type": s3_object_type, "Values": ["arn:aws:s3"]}]}], true)
	eval_fail with input as rule_input([{"ReadWriteType": "WriteOnly", "DataResources": [{"Type": s3_object_type, "Values": ["arn:aws:s3"]}]}], false)
}

test_pass {
	eval_pass with input as rule_input([{"ReadWriteType": "WriteOnly", "DataResources": [{"Type": s3_object_type, "Values": ["arn:aws:s3"]}]}], true)
	eval_pass with input as rule_input([{"ReadWriteType": "All", "DataResources": [{"Type": s3_object_type, "Values": ["arn:aws:s3"]}]}], true)
	eval_pass with input as rule_input([{"ReadWriteType": "All", "DataResources": [{"Type": s3_object_type, "Values": ["arn:aws:s3:::some-bucket"]}]}], true)
	eval_pass with input as rule_input([{"ReadWriteType": "All", "DataResources": [{"Type": not_s3_object_type, "Values": ["arn:aws:s3"]}, {"Type": s3_object_type, "Values": ["arn:aws:s3"]}]}], true)
}

test_not_evaluated {
	not_eval with input as test_data.not_evaluated_trail
}

rule_input(entries, is_multi_region) = test_data.generate_event_selectors(entries, is_multi_region)

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
