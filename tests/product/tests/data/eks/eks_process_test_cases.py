"""
This module provides eks kubelet rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
"""

from commonlib.framework.reporting import SkipReportData, skip_param_case
from configuration import eks

from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS
from ..eks_test_case import EksTestCase

config_1_node_1 = eks.config_1_node_1
config_1_node_2 = eks.config_1_node_2
config_2_node_1 = eks.config_2_node_1
config_2_node_2 = eks.config_2_node_2

cis_eks_3_2_1_pass = EksTestCase(
    rule_tag="CIS 3.2.1",
    node_hostname=config_1_node_1,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_2_1_fail = EksTestCase(
    rule_tag="CIS 3.2.1",
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_1_config_1 = {
    "3.2.1 Kubelet config Authentication.anonymous.enabled==false eval passed": cis_eks_3_2_1_pass,
    "3.2.1 Kubelet config Authentication.anonymous.enabled==false eval failed": cis_eks_3_2_1_fail,
}

cis_eks_3_2_2_pass = EksTestCase(
    rule_tag="CIS 3.2.2",
    node_hostname=config_1_node_1,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_2_2_fail = EksTestCase(
    rule_tag="CIS 3.2.2",
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_2_config_1 = {
    "3.2.2 Kubelet config Authentication.webhook.enabled==true eval passed": cis_eks_3_2_2_pass,
    "3.2.2 Kubelet config Authentication.webhook.enabled==false eval failed": cis_eks_3_2_2_fail,
}

cis_eks_3_2_3_pass = EksTestCase(
    rule_tag="CIS 3.2.3",
    node_hostname=config_1_node_1,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_2_3_fail = EksTestCase(
    rule_tag="CIS 3.2.3",
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_3_config_1 = {
    "3.2.3 Kubelet config x509.clientCAFile exists - eval passed": cis_eks_3_2_3_pass,
    "3.2.3 Kubelet config x509.clientCAFile does not exist eval failed": cis_eks_3_2_3_fail,
}

cis_eks_3_2_4_fail_1 = EksTestCase(
    rule_tag="CIS 3.2.4",
    node_hostname=config_1_node_1,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_4_fail_2 = EksTestCase(
    rule_tag="CIS 3.2.4",
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_4_pass_1 = EksTestCase(
    rule_tag="CIS 3.2.4",
    node_hostname=config_2_node_1,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_2_4_pass_2 = EksTestCase(
    rule_tag="CIS 3.2.4",
    node_hostname=config_2_node_2,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_4_config_1 = {
    "3.2.4 Kubelet args --read-only-port=26492 eval failed": cis_eks_3_2_4_fail_1,
    "3.2.4 Kubelet config readOnlyPort=26492 eval failed": cis_eks_3_2_4_fail_2,
}

cis_eks_3_2_4_config_2 = {
    "3.2.4 Kubelet args --read-only-port=0 eval passed": cis_eks_3_2_4_pass_1,
}

cis_eks_3_2_4_config_2_skip = {
    "3.2.4 Kubelet config readOnlyPort=26492, --read-only-port=0 eval failed": cis_eks_3_2_4_pass_2,
}

cis_eks_3_2_5_fail_1 = EksTestCase(
    rule_tag="CIS 3.2.5",
    node_hostname=config_1_node_1,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_5_fail_2 = EksTestCase(
    rule_tag="CIS 3.2.5",
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_5_pass = EksTestCase(
    rule_tag="CIS 3.2.5",
    node_hostname=config_2_node_1,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_2_5_fail_3 = EksTestCase(
    rule_tag="CIS 3.2.5",
    node_hostname=config_2_node_2,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_5_config_1 = {
    "3.2.5 Kubelet args --streaming-connection-idle-timeout=0 eval failed": cis_eks_3_2_5_fail_1,
    "3.2.5 Kubelet config streamingConnectionIdleTimeout=0s eval failed": cis_eks_3_2_5_fail_2,
}

cis_eks_3_2_5_config_1_skip = {
    "3.2.5 Kubelet config streamingConnectionIdleTimeout=0s eval failed": cis_eks_3_2_5_fail_2,
}

cis_eks_3_2_5_config_2 = {
    "3.2.5 Kubelet args --streaming-connection-idle-timeout=26492s eval passed": cis_eks_3_2_5_pass,
}

cis_eks_3_2_5_config_2_skip = {
    "3.2.5 Kubelet config streamConnection=26492s, arg=0s eval failed": cis_eks_3_2_5_fail_3,
}

cis_eks_3_2_6_fail = EksTestCase(
    rule_tag="CIS 3.2.6",
    node_hostname=config_1_node_1,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_6_pass = EksTestCase(
    rule_tag="CIS 3.2.6",
    node_hostname=config_1_node_2,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_2_6_config_1 = {
    "3.2.6 Kubelet config protectKernelDefaults==false eval failed": cis_eks_3_2_6_fail,
    "3.2.6 Kubelet config protectKernelDefaults default value eval passed": cis_eks_3_2_6_pass,
}

cis_eks_3_2_7_fail_1 = EksTestCase(
    rule_tag="CIS 3.2.7",
    node_hostname=config_1_node_1,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_7_fail_2 = EksTestCase(
    rule_tag="CIS 3.2.7",
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_7_config_1 = {
    "3.2.7 Kubelet config makeIPTablesUtilChains==FALSE eval failed": cis_eks_3_2_7_fail_1,
    "3.2.7 Kubelet args --make-iptables-util-chains==FALSE eval failed": cis_eks_3_2_7_fail_2,
}

cis_eks_3_2_7_pass_1 = EksTestCase(
    rule_tag="CIS 3.2.7",
    node_hostname=config_2_node_1,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_2_7_pass_2 = EksTestCase(
    rule_tag="CIS 3.2.7",
    node_hostname=config_2_node_2,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_2_7_config_2 = {
    "3.2.7 Kubelet config makeIPTablesUtilChains default values eval passed": cis_eks_3_2_7_pass_1,
}

cis_eks_3_2_7_config_2_skip = {
    "3.2.7 Kubelet args over config values --make-iptables-util-chains==TRUE eval passed": cis_eks_3_2_7_pass_2,
}

cis_eks_3_2_8_pass = EksTestCase(
    rule_tag="CIS 3.2.8",
    node_hostname=config_1_node_1,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_2_8_fail = EksTestCase(
    rule_tag="CIS 3.2.8",
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_8_config_1 = {
    "3.2.8 Kubelet args --hostname-override default eval passed": cis_eks_3_2_8_pass,
    "3.2.8 Kubelet args --hostname-override exists eval failed": cis_eks_3_2_8_fail,
}

cis_eks_3_2_9_pass_1 = EksTestCase(
    rule_tag="CIS 3.2.9",
    node_hostname=config_1_node_1,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_2_9_pass_2 = EksTestCase(
    rule_tag="CIS 3.2.9",
    node_hostname=config_1_node_2,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_2_9_fail = EksTestCase(
    rule_tag="CIS 3.2.9",
    node_hostname=config_2_node_2,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_9_config_1 = {
    "3.2.9 Kubelet config eventRecordQPS==2 eval passed": cis_eks_3_2_9_pass_1,
    "3.2.9 Kubelet args --event-qps==5 eval passed": cis_eks_3_2_9_pass_2,
}

cis_eks_3_2_9_config_2 = {
    "3.2.9 Kubelet config eventRecordQPS==5, --event-qps==0 eval failed": cis_eks_3_2_9_fail,
}

cis_eks_3_2_10_fail = EksTestCase(
    rule_tag="CIS 3.2.10",
    node_hostname=config_1_node_1,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_10_pass = EksTestCase(
    rule_tag="CIS 3.2.10",
    node_hostname=config_1_node_2,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_2_10_config_1 = {
    "3.2.10 Kubelet config rotateCertificates==false, eval failed": cis_eks_3_2_10_fail,
    "3.2.10 Kubelet config rotateCertificates not exists, eval passed": cis_eks_3_2_10_pass,
}

cis_eks_3_2_11_fail = EksTestCase(
    rule_tag="CIS 3.2.11",
    node_hostname=config_1_node_1,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_2_11_pass = EksTestCase(
    rule_tag="CIS 3.2.11",
    node_hostname=config_1_node_2,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_2_11_config_1 = {
    "3.2.11 Kubelet config featureGates.RotateKubeletServerCertificate=={}, eval failed": cis_eks_3_2_11_fail,
    "3.2.11 Kubelet config featureGates.RotateKubeletServerCertificate default, eval passed": cis_eks_3_2_11_pass,
}

eks_process_config_1 = {
    **cis_eks_3_2_1_config_1,
    **cis_eks_3_2_2_config_1,
    **cis_eks_3_2_3_config_1,
    **cis_eks_3_2_4_config_1,
    **cis_eks_3_2_5_config_1,
    **skip_param_case(
        cis_eks_3_2_5_config_1_skip,
        data_to_report=SkipReportData(
            skip_reason=(
                "When streamingConnectionIdleTimeout or "
                "--streaming-connection-idle-timeout equals 0 evaluation is passed "
            ),
            url_title="cloudbeat: #632",
            url_link="https://github.com/elastic/cloudbeat/issues/632",
        ),
    ),
    **cis_eks_3_2_6_config_1,
    **cis_eks_3_2_7_config_1,
    **cis_eks_3_2_8_config_1,
    **skip_param_case(
        skip_objects=cis_eks_3_2_9_config_1,
        data_to_report=SkipReportData(
            skip_reason="Rule 3.2.9 - unclear CIS definition and implementation",
            url_title="security-team: #4947",
            url_link="https://github.com/elastic/security-team/issues/4947",
        ),
    ),
    **cis_eks_3_2_10_config_1,
    **cis_eks_3_2_11_config_1,
}

eks_process_config_2 = {
    **cis_eks_3_2_4_config_2,
    **skip_param_case(
        cis_eks_3_2_4_config_2_skip,
        data_to_report=SkipReportData(
            skip_reason="Rule 3.2.4 - When multiple args provided cloudbeat evaluates results incorrectly.",
            url_title="cloudbeat: #719",
            url_link="https://github.com/elastic/cloudbeat/issues/719",
        ),
    ),
    **cis_eks_3_2_5_config_2,
    **skip_param_case(
        cis_eks_3_2_5_config_2_skip,
        data_to_report=SkipReportData(
            skip_reason=(
                "When streamingConnectionIdleTimeout or "
                "--streaming-connection-idle-timeout equals 0 evaluation is passed"
            ),
            url_title="cloudbeat: #632",
            url_link="https://github.com/elastic/cloudbeat/issues/632",
        ),
    ),
    **cis_eks_3_2_7_config_2,
    **skip_param_case(
        cis_eks_3_2_7_config_2_skip,
        data_to_report=SkipReportData(
            skip_reason="Rule 3.2.7 - When multiple args provided cloudbeat evaluates results incorrectly.",
            url_title="cloudbeat: #719",
            url_link="https://github.com/elastic/cloudbeat/issues/719",
        ),
    ),
    **skip_param_case(
        cis_eks_3_2_9_config_2,
        data_to_report=SkipReportData(
            skip_reason="Rule 3.2.9 - unclear CIS definition and implementation",
            url_title="security-team: #4947",
            url_link="https://github.com/elastic/security-team/issues/4947",
        ),
    ),
}

cis_eks_all = {
    "test-eks-config-1": eks_process_config_1,
    "test-eks-config-2": eks_process_config_2,
}

cis_eks_kubelet_cases = cis_eks_all.get(eks.current_config, {})
