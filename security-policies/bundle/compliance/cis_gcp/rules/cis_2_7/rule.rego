package compliance.cis_gcp.rules.cis_2_7

import data.compliance.policy.gcp.monitoring.ensure_log_metric_and_alarm_exists as audit

pattern := `resource.type="gce_firewall_rule"
 AND (protoPayload.methodName:"compute.firewalls.patch"
 OR protoPayload.methodName:"compute.firewalls.insert"
 OR protoPayload.methodName:"compute.firewalls.delete")`

finding := audit.finding(pattern)
