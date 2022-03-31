from elasticsearch import Elasticsearch


class ElasticWrapper:
    """
    Wrapper that uses elasticsearch official package
    """
    def __init__(self, elastic_params):
        self.index = elastic_params.cis_index
        self.es = Elasticsearch(hosts=elastic_params.url,
                                basic_auth=elastic_params.basic_auth)

    def get_index_data(self, index_name: str, query: dict = None):
        result = self.es.search(index=index_name, body=query)
        return result
