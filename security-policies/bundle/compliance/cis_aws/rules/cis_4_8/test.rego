package compliance.cis_aws.rules.cis_4_8

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.compliance.policy.aws_cloudtrail.pattern
import data.lib.test
import future.keywords.if

test_pass if {
	eval_pass with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
		},
		"MetricFilters": [{
			"FilterName": "filter_1",
			"FilterPattern": "{ ($.eventSource = s3.amazonaws.com) && (($.eventName = PutBucketAcl) || ($.eventName = PutBucketPolicy) || ($.eventName = PutBucketCors) || ($.eventName = PutBucketLifecycle) || ($.eventName = PutBucketReplication) || ($.eventName = DeleteBucketPolicy) || ($.eventName = DeleteBucketCors) || ($.eventName = DeleteBucketLifecycle) || ($.eventName = DeleteBucketReplication)) }",
			"ParsedFilterPattern": pattern.complex_expression("&&", [
				pattern.complex_expression("||", [
					pattern.simple_expression("$.eventName", "=", "PutBucketPolicy"),
					pattern.simple_expression("$.eventName", "=", "PutBucketCors"),
					pattern.simple_expression("$.eventName", "=", "DeleteBucketReplication"),
					pattern.simple_expression("$.eventName", "=", "DeleteBucketPolicy"),
					pattern.simple_expression("$.eventName", "=", "PutBucketLifecycle"),
					pattern.simple_expression("$.eventName", "=", "DeleteBucketLifecycle"),
					pattern.simple_expression("$.eventName", "=", "PutBucketAcl"),
					pattern.simple_expression("$.eventName", "=", "DeleteBucketCors"),
					pattern.simple_expression("$.eventName", "=", "PutBucketReplication"),
				]),
				pattern.simple_expression("$.eventSource", "=", "s3.amazonaws.com"),
			]),
		}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])
}

rule_input(entry) := test_data.generate_monitoring_resources(entry)

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
