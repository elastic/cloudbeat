"""
Module to enable and verify the status of the Entity Store feature in Kibana.

This script:
    - Updates Kibana settings to enable Asset Inventory.
    - Creates the required data view if it doesn't exist.
    - Enables the Entity Store via API.
    - Checks the status of the Entity Store and its engines (host, user, service, generic).
    - Logs progress and errors using loguru.

Requires:
    - Proper authentication and configuration for Kibana.
    - loguru for logging.
    - fleet_api and configuration_fleet modules.
"""

import sys
import time

import configuration_fleet as config_fleet
import requests
from fleet_api.data_view_api import create_security_default_data_view
from fleet_api.entity_store_api import (
    enable_entity_store,
    is_entity_store_fully_started,
)
from fleet_api.kibana_settings import update_kibana_settings
from loguru import logger

elk_config = config_fleet.elk_config
ENTITY_STORE_INIT_TIMEOUT = 60  # seconds

if __name__ == "__main__":
    try:
        # Enable the Asset Inventory feature
        update_kibana_settings(
            cfg=elk_config,
            settings={
                "securitySolution:enableAssetInventory": True,
            },
        )

        # Create data view if it doesn't exist
        logger.info("Creating security default data view for entity store...")
        create_security_default_data_view(cfg=elk_config, name="security-solution")

        enable_entity_store(cfg=elk_config)

        # Check the status of the entity store for up to ENTITY_STORE_INIT_TIMEOUT seconds
        start_time = time.time()
        logger.info("====== Entity Store Status ====")
        while time.time() - start_time < ENTITY_STORE_INIT_TIMEOUT:
            if is_entity_store_fully_started(elk_config):
                logger.info("Entity store is fully started!")
                break
            time.sleep(1)
        else:
            logger.error(f"Entity store did not fully start within {ENTITY_STORE_INIT_TIMEOUT} seconds.")
            sys.exit(1)

    except requests.RequestException as e:
        logger.error(f"An HTTP error occurred while enabling entity store: {e}")
        sys.exit(1)
    except (ValueError, KeyError) as e:
        print(f"A configuration error occurred while enabling entity store: {e}")
        sys.exit(1)
