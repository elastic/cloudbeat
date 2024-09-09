"""
AWS Asset Inventory Elastic Compute Cloud verification.
This module verifies presence and correctness of retrieved assets
"""

from datetime import datetime, timedelta

import pytest
from commonlib.utils import get_ES_assets
from product.tests.data.aws_asset_inventory import test_cases as aws_tc
from product.tests.parameters import Parameters, register_params


@pytest.mark.asset_inventory
@pytest.mark.asset_inventory_aws
def test_aws_asset_inventory(
    asset_inventory_client,
    category,
    sub_category,
    type_,
    sub_type,
):
    """
    This data driven test verifies assets published by cloudbeat agent.
    """
    assets = get_ES_assets(
        asset_inventory_client,
        timeout=10,
        category=category,
        sub_category=sub_category,
        type_=type_,
        sub_type=sub_type,
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
    test_aws_asset_inventory,
    Parameters(
        ("category", "sub_category", "type_", "sub_type"),
        [*aws_tc.test_cases.values()],
        ids=[*aws_tc.test_cases.keys()],
    ),
)
