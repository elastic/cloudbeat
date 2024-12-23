package compliance.cis_k8s.rules.cis_1_1_16

import data.compliance.policy.file.ensure_ownership as audit
import future.keywords.if

finding := result if {
	audit.filename_filter("scheduler.conf")
	result := audit.finding("root", "root")
}
