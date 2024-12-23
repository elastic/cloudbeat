package compliance.cis_gcp.rules.cis_2_8

import data.compliance.policy.gcp.monitoring.ensure_log_metric_and_alarm_exists as audit

pattern := `resource.type="gce_route"
 AND (protoPayload.methodName:"compute.routes.delete"
 OR protoPayload.methodName:"compute.routes.insert")`

finding := audit.finding(pattern)
