from tests.elasticsearch.elastic_wrapper import ElasticWrapper
import configuration
import pytest
from loguru import logger


@pytest.fixture(scope="session", autouse=True)
def cspm_client():
    """
    This function (fixture) instantiate ElasticWrapper.
    @return: ElasticWrapper client with cspm index.
    """
    es_client = ElasticWrapper(configuration.elasticsearch.url,
                               configuration.elasticsearch.basic_auth,
                               configuration.elasticsearch.cspm_index)
    logger.info(f"CSPM client with ElasticSearch url: {configuration.elasticsearch.url}")
    return es_client
