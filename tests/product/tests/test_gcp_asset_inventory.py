"""
GCP Asset Inventory Elastic Compute Cloud verification.
This module verifies presence and correctness of retrieved assets
"""

from datetime import datetime, timedelta

import pytest
from commonlib.utils import get_ES_assets
from product.tests.data.gcp_asset_inventory import test_cases as gcp_tc
from product.tests.parameters import Parameters, register_params


@pytest.mark.asset_inventory
@pytest.mark.asset_inventory_gcp
def test_gcp_asset_inventory(
    asset_inventory_client,
    category,
    type_,
):
    """
    This data driven test verifies assets published by cloudbeat agent.
    """
    # pylint: disable=duplicate-code
    assets = get_ES_assets(
        asset_inventory_client,
        timeout=10,
        category=category,
        type_=type_,
        exec_timestamp=datetime.utcnow() - timedelta(minutes=30),
    )

    assert assets is not None, "Expected a list of assets, got None"
    assert isinstance(assets, list) and len(assets) > 0, "Expected the list to be non-empty"
    for asset in assets:
        assert asset.cloud, "Expected .cloud section"
        assert asset.cloud.provider == "gcp", f'Expected "gcp" provider, got {asset.cloud.provider}'
        assert len(asset.asset.id) > 0, "Expected .asset.id list to contain an ID"
        assert len(asset.asset.id[0]) > 0, "Expected the ID to be non-empty"
        assert asset.asset.raw, "Expected the resource under .asset.raw"


register_params(
    test_gcp_asset_inventory,
    Parameters(
        ("category", "type_"),
        [*gcp_tc.test_cases.values()],
        ids=[*gcp_tc.test_cases.keys()],
    ),
)
