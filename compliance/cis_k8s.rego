package compliance.cis_k8s

import data.compliance.cis.rules

default_tags := ["CIS", "CIS v1.6.0", "Kubernetes"]

findings[finding] {
    some rule_id
    data.activated_rules.cis_k8s[rule_id]
    finding = rules[rule_id].finding
}
