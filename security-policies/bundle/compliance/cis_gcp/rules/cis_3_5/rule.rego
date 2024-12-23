package compliance.cis_gcp.rules.cis_3_5

import data.compliance.policy.gcp.dns.ensure_no_sha1 as audit

# Ensure That RSASHA1 Is Not Used for the Zone-Signing Key in Cloud DNS DNSSEC.
finding := audit.finding("ZONE_SIGNING")
