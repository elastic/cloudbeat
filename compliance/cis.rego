package compliance.cis

import data.compliance.cis.rules

tags := ["CIS", "CIS v1.6.0", "Kubernetes"]

findings[finding] {
    some rule_id
    data.activated_rules[rule_id]
    finding = rules[rule_id].finding
}
