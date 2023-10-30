package lib.test

assert_pass(finding) {
	finding.evaluation == "passed"
}

assert_fail(finding) {
	finding.evaluation == "failed"
}
