from dataclasses import dataclass, astuple


@dataclass
class KubeTestCase:
    """
    Represent a test case for Kube API resources
    """
    rule_tag: str
    resource_type: str
    resource_body: dict
    expected: str

    def __iter__(self):
        return iter(astuple(self))

    def __len__(self):
        return len(astuple(self))
