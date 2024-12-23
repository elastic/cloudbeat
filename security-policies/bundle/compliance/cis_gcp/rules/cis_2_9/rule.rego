package compliance.cis_gcp.rules.cis_2_9

import data.compliance.policy.gcp.monitoring.ensure_log_metric_and_alarm_exists as audit

pattern := `resource.type="gce_network"
 AND (protoPayload.methodName:"compute.networks.insert"
 OR protoPayload.methodName:"compute.networks.patch"
 OR protoPayload.methodName:"compute.networks.delete"
 OR protoPayload.methodName:"compute.networks.removePeering"
 OR protoPayload.methodName:"compute.networks.addPeering")`

finding := audit.finding(pattern)
