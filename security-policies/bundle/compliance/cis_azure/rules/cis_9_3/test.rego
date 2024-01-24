package compliance.cis_azure.rules.cis_9_3

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

kind := "azure-web-site"

test_violation if {
	eval_fail with input as test_data.generate_azure_asset(kind, {})
	eval_fail with input as test_data.generate_azure_asset(kind, {"httpsOnly": true})
	eval_fail with input as test_data.generate_azure_asset(kind, {"siteConfig": null})
	eval_fail with input as test_data.generate_azure_asset(kind, {"siteConfig": {}})
	eval_fail with input as test_data.generate_azure_asset(kind, {"siteConfig": {"minTlsVersion": null}})
	eval_fail with input as test_data.generate_azure_asset(kind, {"siteConfig": {"minTlsVersion": "1.0"}})
	eval_fail with input as test_data.generate_azure_asset(kind, {"siteConfig": {"minTlsVersion": "1.1"}})
}

test_pass if {
	eval_pass with input as test_data.generate_azure_asset(kind, {"siteConfig": {"minTlsVersion": "1.2"}})
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
