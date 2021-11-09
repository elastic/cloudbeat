package compliance.cis.rules.cis_1_1_2

import data.lib.test

test_violation {
    test.rule_violation(finding) with input as rule_input("root", "user")
    test.rule_violation(finding) with input as rule_input("user", "root")
    test.rule_violation(finding) with input as rule_input("user", "user")
}

test_pass {
    test.rule_pass(finding) with input as rule_input("root", "root")
}

rule_input(uid, gid) = {
  "osquery": {
    "mode": "0644",
    "path": "/hostfs/etc/kubernetes/manifests/kube-apiserver.yaml",
    "uid": uid,
    "filename": "kube-apiserver.yaml",
    "gid": gid
  }
}