"""
CIS AWS Elastic Compute Cloud asset inventory verification.
This module verifies presence of retrieved assets
"""

from datetime import datetime, timedelta
from functools import partial

import pytest
from commonlib.utils import get_ES_evaluation, res_identifier
from product.tests.data.aws_asset_inventory import aws_ec2_test_cases as aws_ec2_tc
from product.tests.parameters import Parameters, register_params

from .data.constants import RES_NAME


@pytest.mark.aws_ec2_asset_inventory
def test_aws_ec2_asset_inventory(
    elasticsearch_client,
    cloudbeat_agent,
    rule_tag,
    case_identifier,
    expected,
):

    asset = get_ES_asset()

    assert asset is None, f"GOT ASSET: {asset}"


register_params(
    test_aws_ec2_asset_inventory,
    Parameters(
        # ("rule_tag", "case_identifier", "expected"),
        (),
        [*aws_ec2_tc.test_cases.values()],
        ids=[*aws_ec2_tc.test_cases.keys()],
    ),
)
