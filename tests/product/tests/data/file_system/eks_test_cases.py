"""
This module provides eks file system rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
"""

cis_eks_3_1_1 = [
    ('CIS 3.1.1', 'chmod', '0700', '/var/lib/kubelet/kubeconfig', 'failed'),
    ('CIS 3.1.1', 'chmod', '0644', '/var/lib/kubelet/kubeconfig', 'passed')
]

cis_eks_3_1_2 = [
    ('CIS 3.1.2', 'chown', 'root:daemon', '/var/lib/kubelet/kubeconfig', 'failed'),
    ('CIS 3.1.2', 'chown', 'daemon:root', '/var/lib/kubelet/kubeconfig', 'failed'),
    ('CIS 3.1.2', 'chown', 'daemon:daemon', '/var/lib/kubelet/kubeconfig', 'failed'),
    ('CIS 3.1.2', 'chown', 'root:root', '/var/lib/kubelet/kubeconfig', 'passed')
]

cis_eks_3_1_3 = [
    ('CIS 3.1.3', 'chmod', '0700', '/etc/kubernetes/kubelet/kubelet-config.json', 'failed'),
    ('CIS 3.1.3', 'chmod', '0644', '/etc/kubernetes/kubelet/kubelet-config.json', 'passed')
]

cis_eks_3_1_4 = [
    ('CIS 3.1.4', 'chown', 'root:daemon', '/etc/kubernetes/kubelet/kubelet-config.json', 'failed'),
    ('CIS 3.1.4', 'chown', 'daemon:root', '/etc/kubernetes/kubelet/kubelet-config.json', 'failed'),
    ('CIS 3.1.4', 'chown', 'daemon:daemon', '/etc/kubernetes/kubelet/kubelet-config.json', 'failed'),
    ('CIS 3.1.4', 'chown', 'root:root', '/etc/kubernetes/kubelet/kubelet-config.json', 'passed'),
]

cis_eks_kubeconfig = [
    *cis_eks_3_1_1,
    *cis_eks_3_1_2
]

cis_eks_kubelet_config = [
    *cis_eks_3_1_3,
    *cis_eks_3_1_4
]
