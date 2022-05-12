"""
This module provides common configuration of the test project, and also mapping environment variables

"""
import os
from distutils.util import strtobool
from munch import Munch

# --- Cloudbeat agent environment definition ----------------
agent = Munch()
agent.name = os.getenv('AGENT_NAME', 'cloudbeat')
agent.namespace = os.getenv('AGENT_NAMESPACE', 'kube-system')
agent.findings_timeout = 30

# --- Kubernetes environment definition --------------------
kubernetes = Munch()
kubernetes.is_in_cluster_config = bool(strtobool(os.getenv('KUBERNETES_IN_CLUSTER', 'False')))

# --- Elasticsearch environment definition --------------------------------
elasticsearch = Munch()
elasticsearch.hosts = os.getenv('ES_HOST', "localhost")
elasticsearch.user = os.getenv('ES_USER', 'elastic')
elasticsearch.password = os.getenv('ES_PASSWORD', 'changeme')
elasticsearch.basic_auth = (elasticsearch.user, elasticsearch.password)
elasticsearch.port = os.getenv('ES_PORT', 9200)
elasticsearch.protocol = os.getenv('ES_PROTOCOL', 'http')
elasticsearch.url = f"{elasticsearch.protocol}://{elasticsearch.hosts}:{elasticsearch.port}"
elasticsearch.cis_index = os.getenv('CIS_INDEX', "*cloud_security_posture.findings*")

# --- Docker environment definition
docker = Munch()
docker.base_url = os.getenv('DOCKER_URL', "")
docker.use_docker = bool(strtobool(os.getenv('USE_DOCKER', 'True')))

# Printing all environment keys
for key, value in sorted(os.environ.items()):
    print('{}: {}'.format(key, value))
