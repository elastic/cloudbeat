import pytest

testdata = ['file-system', 'process']
# testdata = ['file-system', 'process', 'kube-api']


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

    assert len(result.body['hits']['hits']) > 0
