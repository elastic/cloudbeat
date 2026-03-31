"""
Enable Entity Store v2 on Kibana (v2-only).

Uses the same three-step flow as ecp-synthetics-monitors kibana-api.ts:
internal settings (entityStoreEnableV2), install, then maintainers init;
then polls public entity store status until running or timeout.

Requires:
    - configuration_fleet / elk_config with Kibana URL and auth.
"""

import sys
import time

import configuration_fleet as config_fleet
import requests
from fleet_api.entity_store_api import (
    enable_entity_store_v2,
    init_entity_store_v2_maintainers,
    install_entity_store_v2,
    is_entity_store_v2_fully_started,
)
from loguru import logger

elk_config = config_fleet.elk_config
ENTITY_STORE_INIT_TIMEOUT = 180  # seconds

if __name__ == "__main__":
    try:
        enable_entity_store_v2(cfg=elk_config)
        install_entity_store_v2(cfg=elk_config)
        init_entity_store_v2_maintainers(cfg=elk_config)

        start_time = time.time()
        logger.info("====== Entity Store v2 status poll ====")
        while time.time() - start_time < ENTITY_STORE_INIT_TIMEOUT:
            if is_entity_store_v2_fully_started(elk_config):
                logger.info("Entity store is fully started after v2 install.")
                break
            time.sleep(1)
        else:
            logger.error(
                "Entity store did not fully start within {} seconds after v2 install.",
                ENTITY_STORE_INIT_TIMEOUT,
            )
            sys.exit(1)

    except TimeoutError as exc:
        logger.error("Entity Store v2 setup timed out: {}", exc)
        sys.exit(1)
    except requests.RequestException as exc:
        logger.error("HTTP error while enabling entity store v2: {}", exc)
        sys.exit(1)
    except (ValueError, KeyError) as exc:
        logger.error("Configuration error while enabling entity store v2: {}", exc)
        sys.exit(1)
