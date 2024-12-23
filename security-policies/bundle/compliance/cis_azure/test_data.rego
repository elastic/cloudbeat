package cis_azure.test_data

not_eval_resource := {
	"type": "azure-resource-type",
	"subType": "azure-resource-subtype",
	"resource": {},
}

generate_disk_encryption_settings(type) := {"encryption": {
	"diskEncryptionSetId": "/subscriptions/dead-beef/resourceGroups/RESOURCEGROUP/providers/Microsoft.Compute/diskEncryptionSets/double-disk-encryption-set",
	"type": type,
}}

generate_attached_disk_with_encryption(settings) := generate_disk_with_encryption("Attached", settings)

generate_unattached_disk_with_encryption(settings) := generate_disk_with_encryption("Unattached", settings)

generate_disk_with_encryption(state, settings) := {
	"subType": "azure-disk",
	"resource": {
		"id": "/subscriptions/dead-beef/resourceGroups/resourceGroup/providers/Microsoft.Compute/disks/unattached-disk",
		"location": "eastus",
		"name": "unattached-disk",
		"properties": object.union(
			{
				"creationData": {"createOption": "Empty"},
				"dataAccessAuthMode": "None",
				"diskIOPSReadWrite": 500,
				"diskMBpsReadWrite": 60,
				"diskSizeBytes": 4294967296,
				"diskSizeGB": 4,
				"diskState": state,
				"networkAccessPolicy": "DenyAll",
				"provisioningState": "Succeeded",
				"publicNetworkAccess": "Disabled",
				"timeCreated": "2023-09-28T19:05:41.631Z",
				"uniqueId": "12345-abcdef",
			},
			settings,
		),
		"resource_group": "resourceGroup",
		"subscription_id": "dead-beef",
		"tenant_id": "beef-dead",
		"type": "microsoft.compute/disks",
	},
}

generate_storage_account_with_property(key, value) := {
	"subType": "azure-storage-account",
	"resource": {"properties": {key: value}},
}

generate_storage_account_with_extensions(properties, extension) := {
	"subType": "azure-storage-account",
	"resource": {"properties": properties, "extension": extension},
}

generate_azure_asset(type, properties) := {
	"subType": type,
	"resource": {"properties": properties},
}

generate_azure_asset_with_ext(type, properties, ext) := {
	"subType": type,
	"resource": {"properties": properties, "extension": ext},
}

generate_azure_asset_resource(type, properties) := {
	"subType": type,
	"resource": properties,
}

generate_azure_sku_asset_with_properties(type, properties) := {
	"subType": type,
	"resource": {
		"sku": properties,
		"properties": {},
	},
}

generate_azure_non_sku_asset(type) := {
	"subType": type,
	"resource": {"properties": {}},
}

not_eval_storage_account_empty := {
	"subType": "azure-storage-account",
	"resource": {"properties": {}},
}

not_eval_non_exist_type := {
	"subType": "azure-non-exist",
	"resource": {"properties": {}},
}

generate_postgresql_server_with_ssl_enforcement(enabled) := {
	"subType": "azure-postgresql-server-db",
	"resource": {"properties": {"sslEnforcement": enabled}},
}

generate_postgresql_server_with_extension(ext) := {
	"subType": "azure-postgresql-server-db",
	"resource": {"extension": ext},
}

generate_postgresql_server_with_infrastructure_encryption(enabled) := {
	"subType": "azure-postgresql-server-db",
	"resource": {"properties": {"infrastructureEncryption": enabled}},
}

generate_flexible_postgresql_server_with_extension(ext) := {
	"subType": "azure-flexible-postgresql-server-db",
	"resource": {"extension": ext},
}

generate_mysql_server_with_ssl_enforcement(enabled) := {
	"subType": "azure-mysql-server-db",
	"resource": {"properties": {"sslEnforcement": enabled}},
}

generate_flexible_mysql_server_with_extension(extension) := {
	"subType": "azure-flexible-mysql-server-db",
	"resource": {"extension": extension},
}

generate_activity_log_alerts_no_alerts := {
	"subType": "azure-activity-log-alert",
	"resource": [],
}

generate_activity_log_alerts(rules) := {
	"subType": "azure-activity-log-alert",
	"resource": rules,
}

generate_activity_log_alert(operation_name, category) := {
	"id": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/activityLogAlerts/providers/microsoft.insights/activityLogAlerts/activityLogAlert",
	"subType": "microsoft.insights/activitylogalerts",
	"kind": "activityLogAlert",
	"name": "activityLogAlert",
	"location": "global",
	"resourceGroup": "activityLogAlerts",
	"subscriptionId": "00000000-0000-0000-0000-000000000000",
	"managedBy": "",
	"properties": {
		"description": "",
		"enabled": true,
		"condition": {"allOf": [
			{
				"equals": category,
				"field": "category",
			},
			{
				"equals": operation_name,
				"field": "operationName",
			},
		]},
		"actions": {"actionGroups": []},
		"scopes": ["/subscriptions/00000000-0000-0000-0000-000000000000"],
	},
}

valid_managed_disk := {
	"id": "/subscriptions/sub-id/resourceGroups/cloudbeat-resource-group-1695893762/providers/Microsoft.Compute/disks/cloudbeatVM_OsDisk_1_e736df07f12142a9a2784ea8de9084ce",
	"resourceGroup": "cloudbeat-resource-group-1695893762",
	"storageAccountType": "Standard_LRS",
}

generate_vm(managed_disk) := generate_vm_full(managed_disk, {})

generate_vm_with_extension(extension) := generate_vm_full({}, extension)

# regal ignore:rule-length
generate_vm_full(managed_disk, extension) := {
	"subType": "azure-vm",
	"resource": {
		"extendedLocation": null,
		"id": "/subscriptions/sub-id/resourceGroups/CLOUDBEAT-RESOURCE-GROUP-1695893762/providers/Microsoft.Compute/virtualMachines/CLOUDBEATVM",
		"identity": {
			"principalId": "8536c470-6db4-45b7-a083-b494b3f8481c",
			"tenantId": "tenant-id",
			"type": "SystemAssigned",
		},
		"kind": "",
		"location": "eastus",
		"managedBy": "",
		"name": "cloudbeatVM",
		"plan": null,
		"extension": extension,
		"properties": {
			"extended": {"instanceView": {
				"computerName": "cloudbeatVM",
				"hyperVGeneration": "V2",
				"osName": "ubuntu",
				"osVersion": "22.04",
				"powerState": {
					"code": "PowerState/running",
					"displayStatus": "VM running",
					"level": "Info",
				},
			}},
			"hardwareProfile": {"vmSize": "Standard_DS2_v2"},
			"networkProfile": {"networkInterfaces": [{
				"id": "/subscriptions/sub-id/resourceGroups/cloudbeat-resource-group-1695893762/providers/Microsoft.Network/networkInterfaces/cloudbeatNic",
				"resourceGroup": "cloudbeat-resource-group-1695893762",
			}]},
			"osProfile": {
				"adminUsername": "cloudbeat",
				"allowExtensionOperations": true,
				"computerName": "cloudbeatVM",
				"linuxConfiguration": {
					"disablePasswordAuthentication": false,
					"enableVMAgentPlatformUpdates": false,
					"patchSettings": {
						"assessmentMode": "ImageDefault",
						"patchMode": "ImageDefault",
					},
					"provisionVMAgent": true,
				},
				"requireGuestProvisionSignal": true,
				"secrets": [],
			},
			"provisioningState": "Succeeded",
			"storageProfile": {
				"dataDisks": [],
				"diskControllerType": "SCSI",
				"imageReference": {
					"exactVersion": "22.04.202309190",
					"offer": "0001-com-ubuntu-server-jammy",
					"publisher": "canonical",
					"sku": "22_04-lts-gen2",
					"version": "latest",
				},
				"osDisk": {
					"caching": "ReadWrite",
					"createOption": "FromImage",
					"deleteOption": "Detach",
					"diskSizeGB": 30,
					"managedDisk": managed_disk,
					"name": "cloudbeatVM_OsDisk_1_e736df07f12142a9a2784ea8de9084ce",
					"osType": "Linux",
				},
			},
			"timeCreated": "2023-09-28T09:36:20.316Z",
			"vmId": "a3842848-355a-42ab-9fb1-488587f301f3",
		},
		"resourceGroup": "cloudbeat-resource-group-1695893762",
		"sku": null,
		"subscriptionId": "sub-id",
		"tags": null,
		"tenantId": "tenant-id",
		"type": "microsoft.compute/virtualmachines",
		"zones": null,
	},
}

generate_insights_components_empty := {
	"subType": "azure-insights-component",
	"resource": [],
}

generate_insights_components(rules) := {
	"subType": "azure-insights-component",
	"resource": rules,
}

generate_insights_component(resource_group, name) := {
	"id": sprintf("/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/%s/providers/microsoft.insights/components/%s", [resource_group, name]),
	"name": name,
	"type": "microsoft.insights/components",
	"location": "eastus",
	"tags": {},
	"kind": "web",
	"etag": "\"00000000-0000-0000-0000-000000000004\"",
	"properties": {
		"ApplicationId": name,
		"AppId": "00000000-0000-0000-0000-000000000003",
		"Application_Type": "web",
		"Flow_Type": "Redfield",
		"Request_Source": "IbizaAIExtension",
		"InstrumentationKey": "00000000-0000-0000-0000-000000000002",
		"ConnectionString": "InstrumentationKey=00000000-0000-0000-0000-000000000002;IngestionEndpoint=https://eastus-8.in.applicationinsights.azure.com/;LiveEndpoint=https://eastus.livediagnostics.monitor.azure.com/",
		"Name": name,
		"CreationDate": "2023-10-29T15:04:34.0549974+00:00",
		"TenantId": "00000000-0000-0000-0000-000000000001",
		"provisioningState": "Succeeded",
		"SamplingPercentage": null,
		"RetentionInDays": 90,
		"WorkspaceResourceId": sprintf("/subscriptions/00000000-0000-0000-0000-000000000001/resourcegroups/%s/providers/Microsoft.OperationalInsights/workspaces/00000000-0000-0000-0000-000000000001-%s-EUS", [resource_group, resource_group]),
		"IngestionMode": "LogAnalytics",
		"publicNetworkAccessForIngestion": "Enabled",
		"publicNetworkAccessForQuery": "Enabled",
		"Ver": "v2",
	},
}

generate_diagnostic_settings_empty := {
	"subType": "azure-diagnostic-settings",
	"resource": [],
}

generate_diagnostic_settings(rules) := {
	"subType": "azure-diagnostic-settings",
	"resource": rules,
}

generate_diagnostic_setting_element(sub_id, resource_group, name, logs) := {
	"id": sprintf("/subscriptions/%s/providers/microsoft.insights/diagnosticSettings/%s", [sub_id, name]),
	"name": name,
	"properties": {
		"logs": logs,
		"metrics": [],
		"workspaceId": sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.OperationalInsights/workspaces/%s-workspace", [sub_id, resource_group, name]),
		"logAnalyticsDestinationType": null,
	},
}

generate_diagnostic_setting_element_logs(flags) := [
	generate_diagnostic_setting_element_log("Administrative", flags.Administrative),
	generate_diagnostic_setting_element_log("Security", flags.Security),
	generate_diagnostic_setting_element_log("Policy", flags.Policy),
	generate_diagnostic_setting_element_log("Alert", flags.Alert),
	generate_diagnostic_setting_element_log("ServiceHealth", false),
	generate_diagnostic_setting_element_log("Recommendation", false),
	generate_diagnostic_setting_element_log("Autoscale", false),
	generate_diagnostic_setting_element_log("ResourceHealth", false),
]

generate_diagnostic_setting_element_log(category, enabled) := {
	"category": category,
	"categoryGroup": null,
	"enabled": enabled,
	"retentionPolicy": {
		"days": 0,
		"enabled": false,
	},
}

generate_key_vault_extension_key(attributes) := {"properties": {"attributes": attributes}}

generate_key_vault_rbac(extension) := {
	"subType": "azure-vault",
	"resource": {
		"properties": {"enableRbacAuthorization": true},
		"extension": extension,
	},
}

generate_key_vault(properties, extension) := {
	"subType": "azure-vault",
	"resource": {
		"properties": properties,
		"extension": extension,
	},
}

generate_security_contacts(resources) := {
	"subType": "azure-security-contacts",
	"resource": resources,
}

generate_single_security_contact(name, properties) := {
	"name": name,
	"properties": properties,
}

generate_security_auto_provisioning_settings(resources) := {
	"subType": "azure-security-auto-provisioning-settings",
	"resource": resources,
}
