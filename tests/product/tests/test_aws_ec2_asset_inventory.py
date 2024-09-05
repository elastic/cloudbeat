"""
CIS AWS Elastic Compute Cloud asset inventory verification.
This module verifies presence of retrieved assets
"""

from datetime import datetime, timedelta
from functools import partial

import pytest
from commonlib.utils import get_ES_assets, res_identifier
from product.tests.data.aws_asset_inventory import aws_ec2_test_cases as aws_ec2_tc
from product.tests.parameters import Parameters, register_params

from .data.constants import RES_NAME


@pytest.mark.asset_inventory
@pytest.mark.asset_inventory_aws
def test_aws_ec2_asset_inventory(
    asset_inventory_client,
):
    assets = get_ES_assets(
        asset_inventory_client,
        timeout=10,
        category="infrastructure",
        sub_category="compute",
        type_="virtual-machine",
        sub_type="ec2-instance",
        exec_timestamp=datetime.utcnow() - timedelta(minutes=30),
    )

    assert assets is not None, "Expected a list of assets, got None"
    assert isinstance(assets, list) and len(assets) > 0, "Expected the list to be non-empty"
    for asset in assets:
        assert asset.cloud, "Expected .cloud section"
        assert asset.cloud.provider == "aws", f'Expected "aws" provider, got {asset.cloud.provider}'
        assert len(asset.asset.id) > 0, "Expected .asset.id list to contain an ID"
        assert len(asset.asset.id[0]) > 0, "Expected the ID to be non-empty"
        assert asset.asset.raw, "Expected the resource under .asset.raw"


register_params(
    test_aws_ec2_asset_inventory,
    Parameters(
        (),
        [*aws_ec2_tc.test_cases.values()],
        ids=[*aws_ec2_tc.test_cases.keys()],
    ),
)
