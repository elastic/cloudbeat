package lib.test

assert_pass(finding) {
	finding.evaluation == "passed"
}

assert_violation(finding) {
	finding.evaluation == "violation"
}
