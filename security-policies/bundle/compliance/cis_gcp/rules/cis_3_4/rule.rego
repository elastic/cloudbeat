package compliance.cis_gcp.rules.cis_3_4

import data.compliance.policy.gcp.dns.ensure_no_sha1 as audit

# Ensure That RSASHA1 Is Not Used for the Key-Signing Key in Cloud DNS DNSSEC.
finding := audit.finding("KEY_SIGNING")
