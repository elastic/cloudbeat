"""
This module is a wrapper of ElasticSearch low-level client
"""

from elasticsearch import Elasticsearch


class ElasticWrapper:
    """
    Wrapper that uses elasticsearch official package
    """

    def __init__(self, url: str, basic_auth: tuple, index: str, use_ssl: bool = True):
        self.index = index
        verify_certs = use_ssl
        ssl_show_warn = use_ssl
        self.es_client = Elasticsearch(
            hosts=url,
            basic_auth=basic_auth,
            retry_on_timeout=True,
            verify_certs=verify_certs,
            ssl_show_warn=ssl_show_warn,
        )

    def get_index_data(
        self,
        query: dict,
        sort: list,
        size: int = 1,
    ) -> dict:
        """
        This method retrieves data from specified index
        @param query: Query to be applied on index
        @param size: The number of hits to return.
        @param sort: Sorting order
        @return: Result dictionary
        """
        result = self.es_client.search(
            index=self.index,
            query=query,
            size=size,
            sort=sort,
        )
        return result

    @staticmethod
    def get_total_value(data: dict) -> int:
        """
        This method retrieves total hits value
        @param data: Data dictionary from elasticsearch
        @return: Total Value integer
        """
        ret_value = data.get("hits", {}).get("total", {}).get("value", 0)
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
            ret_value = data["hits"]["hits"][index]["_source"]
        except IndexError:
            return {}
        return ret_value

    @staticmethod
    def get_doc_hits(data: dict) -> dict:
        """
        This method parses data and retrieves hits dictionary
        @param data: Data dictionary
        @return: Hits dictionary
        """
        ret_value = data["hits"]["hits"]
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
                    {"term": term},
                    {"range": {"@timestamp": {"gte": "now-5m"}}},
                ],
            },
        }

        sort = [{"@timestamp": {"order": "desc"}}]

        return query, sort

    @staticmethod
    def build_es_must_match_query(must_query_list: list[dict], time_range: str):
        """
        Build an Elasticsearch 'must' query with the given query list.

        Args:
            must_query_list (list[dict]): List of queries.
            time_range (str): Time range for filtering the query.

        Returns:
            tuple: Tuple containing the Elasticsearch query and sorting order.

        """
        query = {
            "bool": {
                "must": must_query_list,
                "filter": [{"range": {"@timestamp": {"gte": time_range}}}],
            },
        }

        sort = [{"@timestamp": {"order": "desc"}}]

        return query, sort
