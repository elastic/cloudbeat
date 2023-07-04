"""
Define the kspm_client fixture.
"""
import configuration
import pytest
from loguru import logger
from tests.elasticsearch.elastic_wrapper import ElasticWrapper


@pytest.fixture(scope="session", autouse=True)
def kspm_client():
    """
    This function (fixture) instantiate ElasticWrapper.
    @return: ElasticWrapper client with kspm index.
    """
    es_client = ElasticWrapper(configuration.elasticsearch.url,
                               configuration.elasticsearch.basic_auth,
                               configuration.elasticsearch.kspm_index)
    logger.info(f"KSPM client with ElasticSearch url: {configuration.elasticsearch.url}")
    return es_client
