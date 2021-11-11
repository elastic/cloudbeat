package lib.test

rule_pass(finding) {
	finding.evaluation == "passed"
}

rule_violation(finding) {
	finding.evaluation == "violation"
}
