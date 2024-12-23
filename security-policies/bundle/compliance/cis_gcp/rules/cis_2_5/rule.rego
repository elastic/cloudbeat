package compliance.cis_gcp.rules.cis_2_5

import data.compliance.policy.gcp.monitoring.ensure_log_metric_and_alarm_exists as audit

pattern := `protoPayload.methodName="SetIamPolicy" AND protoPayload.serviceData.policyDelta.auditConfigDeltas:*`

finding := audit.finding(pattern)
