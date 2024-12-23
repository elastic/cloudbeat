package compliance.cis_azure.rules.cis_7_1

import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

valid_bastion := {
	"extendedLocation": null,
	"id": "/subscriptions/sub-id/resourceGroups/cloudbeat/providers/Microsoft.Network/bastionHosts/cloudbeat",
	"identity": null,
	"kind": "",
	"location": "eastus",
	"managedBy": "",
	"name": "cloudbeat",
	"plan": null,
	"properties": {
		"disableCopyPaste": false,
		"dnsName": "dns.bastion.azure.com",
		"enableIpConnect": false,
		"enableKerberos": false,
		"enableShareableLink": false,
		"enableTunneling": false,
		"ipConfigurations": [{
			"etag": "W/\"57925490-08e8-4f8e-9a75-e66f4c54f7e2\"",
			"id": "/subscriptions/sub-id/resourceGroups/cloudbeat/providers/Microsoft.Network/bastionHosts/cloudbeat/bastionHostIpConfigurations/IpConf",
			"name": "IpConf",
			"properties": {
				"privateIPAllocationMethod": "Dynamic",
				"provisioningState": "Succeeded",
				"publicIPAddress": {
					"id": "/subscriptions/sub-id/resourceGroups/cloudbeat/providers/Microsoft.Network/publicIPAddresses/cloudbeatBastion-ip",
					"resourceGroup": "cloudbeat",
				},
				"subnet": {
					"id": "/subscriptions/sub-id/resourceGroups/cloudbeat/providers/Microsoft.Network/virtualNetworks/cloudbeatBastion/subnets/AzureBastionSubnet",
					"resourceGroup": "cloudbeat",
				},
			},
			"resourceGroup": "cloudbeat",
			"type": "Microsoft.Network/bastionHosts/bastionHostIpConfigurations",
		}],
		"provisioningState": "Succeeded",
		"scaleUnits": 2,
	},
	"resourceGroup": "cloudbeat",
	"sku": {"name": "Standard"},
	"subscriptionId": "sub-id",
	"tags": {},
	"tenantId": "tenant-id",
	"type": "microsoft.network/bastionhosts",
	"zones": null,
}

generate_bastions(assets) := {
	"subType": "azure-bastion",
	"resource": assets,
}

test_violation if {
	eval_fail with input as generate_bastions([])
	eval_fail with input as generate_bastions([{}])
}

test_pass if {
	eval_pass with input as generate_bastions([valid_bastion])
	eval_pass with input as generate_bastions([valid_bastion, valid_bastion])
}

test_not_evaluated if {
	not_eval with input as {}
	not_eval with input as {"subType": "other-type", "resource": {"assets": {}}}
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
