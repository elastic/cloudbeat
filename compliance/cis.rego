package compliance.cis

import data.compliance.cis.rules

tags := ["CIS", "CIS v1.6.0", "Kubernetes"]

# CIS 1.1.1
findings[finding] {
    data.activated_rules.cis_1_1_1
    finding = rules.cis_1_1_1.finding
}

# CIS 1.1.2
findings[finding] {
    data.activated_rules.cis_1_1_2
    finding = rules.cis_1_1_2.finding
}