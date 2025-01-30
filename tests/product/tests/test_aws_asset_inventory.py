"""
AWS Asset Inventory Elastic Compute Cloud verification.
This module verifies presence and correctness of retrieved entities
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
    type_,
):
    """
    This data driven test verifies entities published by cloudbeat agent.
    """
    entities = get_ES_assets(
        asset_inventory_client,
        timeout=10,
        category=category,
        type_=type_,
        exec_timestamp=datetime.utcnow() - timedelta(minutes=30),
    )

    assert entities is not None, "Expected a list of entities, got None"
    assert isinstance(entities, list) and len(entities) > 0, "Expected the list to be non-empty"
    for entity in entities:
        assert entity.cloud, "Expected .cloud section"
        assert entity.cloud.provider == "aws", f'Expected "aws" provider, got {entity.cloud.provider}'
        assert len(entity.entity.id) > 0, "Expected .entity.id list to contain an ID"
        assert len(entity.entity.id[0]) > 0, "Expected the ID to be non-empty"
        assert entity.Attributes, "Expected the resource under .Attributes"


register_params(
    test_aws_asset_inventory,
    Parameters(
        ("category", "type_"),
        [*aws_tc.test_cases.values()],
        ids=[*aws_tc.test_cases.keys()],
    ),
)
