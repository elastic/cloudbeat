"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""
from datetime import datetime, timedelta
import pytest

from commonlib.utils import get_ES_evaluation

from product.tests.data.k8s_object import eks_k8s_object_test_cases as eks_k8s_object_tc
from product.tests.parameters import register_params, Parameters


@pytest.mark.eks_k8s_objects_rules
def test_eks_kube_objects(
    elastic_client,
    cloudbeat_agent,
    rule_tag,
    test_resource_id,
    expected,
):
    """
    This data driven test verifies rules and findings return by cloudbeat agent.
    In order to add new cases @pytest.mark.parameterize section shall be updated.
    Setup and teardown actions are defined in data method.
    This test creates verifies that cloudbeat returns correct finding.
    @param rule_tag: Name of rule to be verified.
    @param test_resource_id: Pod resource id label
    @param expected: Result to be found in finding evaluation field.
    @return: None - Test Pass / Fail result is generated.
    """
    # pylint: disable=duplicate-code

    def identifier(eval_resource):
        try:
            eval_resource = eval_resource.resource.raw
            return eval_resource.metadata.labels.testResourceId == test_resource_id
        except AttributeError:
            return False

    evaluation = get_ES_evaluation(
        elastic_client=elastic_client,
        timeout=cloudbeat_agent.findings_timeout,
        rule_tag=rule_tag,
        exec_timestamp=datetime.utcnow() - timedelta(hours=1),
        resource_identifier=identifier,
    )

    assert evaluation is not None, f"No evaluation for rule {rule_tag} could be found"
    assert evaluation == expected, f"Rule {rule_tag} verification failed," f"expected: {expected}, got: {evaluation}"


# register_params(
#     test_eks_kube_objects,
#     Parameters(
#         ("rule_tag", "test_resource_id", "expected"),
#         [
#             *eks_k8s_object_tc.cis_eks_4_2_1.values(),
#             *eks_k8s_object_tc.cis_eks_4_2_2.values(),
#             *eks_k8s_object_tc.cis_eks_4_2_3.values(),
#             *eks_k8s_object_tc.cis_eks_4_2_4.values(),
#             *eks_k8s_object_tc.cis_eks_4_2_5.values(),
#             *eks_k8s_object_tc.cis_eks_4_2_6.values(),
#             *eks_k8s_object_tc.cis_eks_4_2_7.values(),
#             *skip_param_case(
#                 skip_list=[*eks_k8s_object_tc.cis_eks_4_2_8.values()],
#                 data_to_report=SkipReportData(
#                     url_title="cloudbeat: #500",
#                     url_link="https://github.com/elastic/cloudbeat/issues/500",
#                     skip_reason="Retest after testing configuration will be fixed.",
#                 ),
#             ),
#             *eks_k8s_object_tc.cis_eks_4_2_9.values(),
#         ],
#         ids=[
#             *eks_k8s_object_tc.cis_eks_4_2_1.keys(),
#             *eks_k8s_object_tc.cis_eks_4_2_2.keys(),
#             *eks_k8s_object_tc.cis_eks_4_2_3.keys(),
#             *eks_k8s_object_tc.cis_eks_4_2_4.keys(),
#             *eks_k8s_object_tc.cis_eks_4_2_5.keys(),
#             *eks_k8s_object_tc.cis_eks_4_2_6.keys(),
#             *eks_k8s_object_tc.cis_eks_4_2_7.keys(),
#             *eks_k8s_object_tc.cis_eks_4_2_8.keys(),
#             *eks_k8s_object_tc.cis_eks_4_2_9.keys(),
#         ],
#     ),
# )

register_params(
    test_eks_kube_objects,
    Parameters(
        ("rule_tag", "test_resource_id", "expected"),
        [*eks_k8s_object_tc.cis_eks_k8s_object_cases.values()],
        ids=[*eks_k8s_object_tc.cis_eks_k8s_object_cases.keys()],
    ),
)
