package lib.test

import future.keywords.if

assert_pass(finding) if {
	finding.evaluation == "passed"
}

assert_fail(finding) if {
	finding.evaluation == "failed"
}
