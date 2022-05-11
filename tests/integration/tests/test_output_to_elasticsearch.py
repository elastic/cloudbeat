import pytest
import json
import allure

testdata = ['file', 'process', 'k8s_object']


@pytest.mark.integration
@pytest.mark.ci_cloudbeat
@pytest.mark.parametrize("match_type", testdata)
def test_elastic_index_exists(elastic_client, match_type):
    """
    This test verifies that findings of all types are sending to elasticsearch
    :param elastic_client: Elastic API client
    :param match_type: Findings type for matching
    :return:
    """
    file_system_query = {
        "size": 1,
        "query": {
            "match": {
                "type": match_type
            }
        },
        "sort": [{
            "@timestamp": {
                "order": "desc"
            }
        }]
    }
    result = elastic_client.get_index_data(index_name=elastic_client.index, query=file_system_query)
    allure.attach(json.dumps(result['hits']['hits'][0]['_source'], indent=4, sort_keys=True),
                  match_type,
                  attachment_type=allure.attachment_type.JSON)
    assert len(result.body['hits']['hits']) > 0, f"The findings of type {match_type} not found"
