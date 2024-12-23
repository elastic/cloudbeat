package compliance.cis_azure.rules.cis_5_5

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

# No filter, all resources will be checked
finding := common.generate_result_without_expected(
	common.calculate_result(ensure_sku_valid),
	{"Resource": data_adapter.resource},
)

ensure_sku_tier if {
	data_adapter.resource.sku.tier != "Basic"
} else := false

ensure_sku_valid := r if {
	data_adapter.resource.sku != null
	r = ensure_sku_tier
}
