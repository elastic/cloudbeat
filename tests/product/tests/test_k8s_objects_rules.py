"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""
from datetime import datetime
import uuid

import pytest

from loguru import logger
from product.tests.data.k8s_object import k8s_object_rules as k8s_tc
from product.tests.parameters import register_params, Parameters

from commonlib.utils import get_ES_evaluation
from commonlib.framework.reporting import skip_param_case, SkipReportData


@pytest.mark.k8s_object_rules
def test_kube_resource_patch(
    elastic_client,
    test_env,
    rule_tag,
    resource_type,
    resource_body,
    expected,
):
    """
    Test kube resource
    @param test_env: pre step that set-ups a kube resources to test on
    @param rule_tag: rule tag in the CIS benchmark
    @param resource_type: kube resource type, e.g., Pod, ServiceAccount, etc.
    @param resource_body: a dict to represent the relevant properties of the resource
    @param expected: "failed" or "passed"
    """
    # pylint: disable=duplicate-code

    k8s_client, _, agent_config = test_env

    # make sure resource exists
    metadata = resource_body["metadata"]
    relevant_metadata = {k: metadata[k] for k in ("name", "namespace") if k in metadata}
    try:
        resource = k8s_client.get_resource(resource_type=resource_type, **relevant_metadata)
    except TypeError as type_error:
        logger.error(type_error)
        resource = k8s_client.get_resource(
            resource_type=resource_type,
            namespace=agent_config.namespace,
            **relevant_metadata,
        )

    assert resource, f"Resource {resource_type} not found"

    test_resource_id = str(uuid.uuid4())

    labels = metadata.setdefault("labels", {})
    labels["test_resource_id"] = test_resource_id

    # patch resource
    resource = k8s_client.patch_resources(
        resource_type=resource_type,
        body=resource_body,
        **relevant_metadata,
    )
    if resource is None:
        raise ValueError(
            f"Could not patch resource type {resource_type}:" f" {relevant_metadata} with patch {resource_body}",
        )

    def match_resource(ident_resource):
        try:
            eval_resource = ident_resource.resource.raw
            return eval_resource.metadata.labels.test_resource_id == test_resource_id
        except AttributeError:
            return False

    evaluation = get_ES_evaluation(
        elastic_client=elastic_client,
        timeout=agent_config.findings_timeout,
        rule_tag=rule_tag,
        exec_timestamp=datetime.utcnow(),
        resource_identifier=match_resource,
    )

    assert evaluation is not None, f"No evaluation for rule {rule_tag} could be found"
    assert evaluation == expected, f"Rule {rule_tag} verification failed, " f"expected: {expected} actual: {evaluation}"


register_params(
    test_kube_resource_patch,
    Parameters(
        ("rule_tag", "resource_type", "resource_body", "expected"),
        [
            *k8s_tc.cis_5_1_3.values(),
            *k8s_tc.cis_5_1_5.values(),
            *k8s_tc.cis_5_1_6.values(),
            *k8s_tc.cis_5_2_3.values(),
            *k8s_tc.cis_5_2_4.values(),
            *k8s_tc.cis_5_2_5.values(),
            *k8s_tc.cis_5_2_2.values(),
            *k8s_tc.cis_5_2_6.values(),
            *skip_param_case(
                skip_objects=[*k8s_tc.cis_5_2_7.values()],
                data_to_report=SkipReportData(
                    url_title="security-team: #4540",
                    url_link="https://github.com/elastic/security-team/issues/4540",
                    skip_reason="Known issue: incorrect implementation",
                ),
            ),
            *k8s_tc.cis_5_2_8.values(),
        ],
        ids=[
            *k8s_tc.cis_5_1_3.keys(),
            *k8s_tc.cis_5_1_5.keys(),
            *k8s_tc.cis_5_1_6.keys(),
            *k8s_tc.cis_5_2_3.keys(),
            *k8s_tc.cis_5_2_4.keys(),
            *k8s_tc.cis_5_2_5.keys(),
            *k8s_tc.cis_5_2_2.keys(),
            *k8s_tc.cis_5_2_6.keys(),
            *k8s_tc.cis_5_2_7.keys(),
            *k8s_tc.cis_5_2_8.keys(),
            # *k8s_tc.cis_5_2_9.keys(), - TODO: cases are not implemented
            # *k8s_tc.cis_5_2_10.keys() - TODO: cases are not implemented
        ],
    ),
)
