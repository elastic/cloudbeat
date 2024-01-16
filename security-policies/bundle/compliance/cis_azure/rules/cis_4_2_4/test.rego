package compliance.cis_azure.rules.cis_4_2_4

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlVulnerabilityAssessmentSettings": []})
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlVulnerabilityAssessmentSettings": [{"notificationEmails": []}]})
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlVulnerabilityAssessmentSettings": [
		{
			"notificationEmails": [],
			"storageAccountName": "full",
		},
		{
			"recurringScansEnabled": ["test@elastic.co"],
			"notificationEmails": "full",
		},
	]})
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlVulnerabilityAssessmentSettings": [
		{"storageAccountName": "full"},
		{
			"notificationEmails": ["test@elastic.co"],
			"storageAccountName": "full",
		},
	]})
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlVulnerabilityAssessmentSettings": [{"notificationEmails": ["test@elastic.co"]}]})
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {})
}

test_pass if {
	eval_pass with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlVulnerabilityAssessmentSettings": [{
		"notificationEmails": ["test@elastic.co"],
		"storageAccountName": "sa",
	}]})
	eval_pass with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlVulnerabilityAssessmentSettings": [{
		"notificationEmails": ["test@elastic.co", "test2@elastic.co"],
		"storageAccountName": "sa",
	}]})
	eval_pass with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlVulnerabilityAssessmentSettings": [
		{
			"notificationEmails": ["test@elastic.co"],
			"storageAccountName": "sa",
		},
		{
			"notificationEmails": ["test@elastic.co"],
			"storageAccountName": "sa",
		},
	]})
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_non_exist_type
}

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
