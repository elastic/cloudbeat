"""
This module is a wrapper of ElasticSearch low-level client
"""

from elasticsearch import Elasticsearch


class ElasticWrapper:
    """
    Wrapper that uses elasticsearch official package
    """

    def __init__(self, elastic_params):
        self.index = elastic_params.cis_index
        self.es_client = Elasticsearch(hosts=elastic_params.url,
                                       basic_auth=elastic_params.basic_auth)

    def get_index_data(self, index_name: str,
                       query: dict,
                       sort: list,
                       size: int = 1) -> dict:
        """
        This method retrieves data from specified index
        @param index_name: Name of index the data should be received from
        @param query: Query to be applied on index
        @param size: The number of hits to return.
        @param sort: Sorting order
        @return: Result dictionary
        """
        result = self.es_client.search(index=index_name,
                                       query=query,
                                       size=size,
                                       sort=sort)
        return result

    @staticmethod
    def get_total_value(data: dict) -> int:
        """
        This method retrieves total hits value
        @param data: Data dictionary from elasticsearch
        @return: Total Value integer
        """
        ret_value = data.get('hits', {}) \
            .get('total', {}) \
            .get('value', 0)
        return ret_value

    @staticmethod
    def get_doc_source(data: dict, index: int = 0) -> dict:
        """
        This method parses result, and retrieves _source section dictionary
        @param data: Data dictionary
        @param index: Document index in result dictionary
        @return: Source dictionary
        """
        try:
            ret_value = data['hits']['hits'][index]['_source']
        except IndexError as ex:
            print(ex)
            return {}
        return ret_value

    @staticmethod
    def get_doc_hits(data: dict) -> dict:
        """
        This method parses data and retrieves hits dictionary
        @param data: Data dictionary
        @return: Hits dictionary
        """
        ret_value = data['hits']['hits']
        return ret_value

    @staticmethod
    def build_es_query(term: dict) -> (dict, list):
        """
        This method builds a ES query based on the provided param
        @param term: search term to be matched against
        @return: ES query and sorting order
        """
        query = {
            "bool": {
                "filter": [
                    {
                        "term": term
                    },
                    {
                        "range": {
                            "@timestamp": {
                                "gte": "now-30s"
                            }
                        }
                    }
                ]
            }
        }

        sort = [{
            "@timestamp": {
                "order": "desc"
            }
        }]

        return query, sort
