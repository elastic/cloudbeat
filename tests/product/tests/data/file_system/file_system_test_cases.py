"""
This module provides file system rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
"""

cis_1_1_1 = [
    ('CIS 1.1.1', 'chmod', '0700', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'failed'),
    ('CIS 1.1.1', 'chmod', '0644', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'passed')
]

cis_1_1_2 = [
    ('CIS 1.1.2', 'chown', 'daemon:daemon', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'failed'),
    ('CIS 1.1.2', 'chown', 'root:root', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'passed'),
]

cis_1_1_3 = [
    ('CIS 1.1.3', 'chmod', '0700', '/etc/kubernetes/manifests/kube-controller-manager.yaml', 'failed'),
    ('CIS 1.1.3', 'chmod', '0644', '/etc/kubernetes/manifests/kube-controller-manager.yaml', 'passed'),
]

cis_1_1_4 = [
    ('CIS 1.1.4', 'chown', 'root:daemon', '/etc/kubernetes/manifests/kube-controller-manager.yaml', 'failed'),
    ('CIS 1.1.4', 'chown', 'daemon:root', '/etc/kubernetes/manifests/kube-controller-manager.yaml', 'failed'),
    ('CIS 1.1.4', 'chown', 'daemon:daemon', '/etc/kubernetes/manifests/kube-controller-manager.yaml', 'failed'),
    ('CIS 1.1.4', 'chown', 'root:root', '/etc/kubernetes/manifests/kube-controller-manager.yaml', 'passed'),
]
cis_1_1_5 = [
    ('CIS 1.1.5', 'chmod', '0700', '/etc/kubernetes/manifests/kube-scheduler.yaml', 'failed'),
    ('CIS 1.1.5', 'chmod', '0644', '/etc/kubernetes/manifests/kube-scheduler.yaml', 'passed'),
]

cis_1_1_6 = [
    ('CIS 1.1.6', 'chown', 'root:daemon', '/etc/kubernetes/manifests/kube-scheduler.yaml', 'failed'),
    ('CIS 1.1.6', 'chown', 'root:root', '/etc/kubernetes/manifests/kube-scheduler.yaml', 'passed'),
]

cis_1_1_7 = [
    ('CIS 1.1.7', 'chmod', '0700', '/etc/kubernetes/manifests/etcd.yaml', 'failed'),
    ('CIS 1.1.7', 'chmod', '0644', '/etc/kubernetes/manifests/etcd.yaml', 'passed'),

]

cis_1_1_8 = [
    ('CIS 1.1.8', 'chown', 'root:daemon', '/etc/kubernetes/manifests/etcd.yaml', 'failed'),
    ('CIS 1.1.8', 'chown', 'daemon:root', '/etc/kubernetes/manifests/etcd.yaml', 'failed'),
    ('CIS 1.1.8', 'chown', 'daemon:daemon', '/etc/kubernetes/manifests/etcd.yaml', 'failed'),
    ('CIS 1.1.8', 'chown', 'root:root', '/etc/kubernetes/manifests/etcd.yaml', 'passed'),
]

cis_1_1_11 = [
    ('CIS 1.1.11', 'chmod', '0710', '/var/lib/etcd', 'failed'),
    ('CIS 1.1.11', 'chmod', '0600', '/var/lib/etcd', 'passed'),
]

cis_1_1_12 = [
    ('CIS 1.1.12', 'chown', 'root:root', '/var/lib/etcd', 'failed'),
    ('CIS 1.1.12', 'chown', 'etcd:root', '/var/lib/etcd/', 'failed'),
    ('CIS 1.1.12', 'chown', 'root:etcd', '/var/lib/etcd', 'failed'),
    ('CIS 1.1.12', 'chown', 'root:etcd', '/var/lib/etcd/some_file.txt', 'failed'),
    ('CIS 1.1.12', 'chown', 'etcd:etcd', '/var/lib/etcd', 'passed'),
    ('CIS 1.1.12', 'chown', 'etcd:etcd', '/var/lib/etcd/some_file.txt', 'passed'),
]

cis_1_1_13 = [
    ('CIS 1.1.13', 'chmod', '0700', '/etc/kubernetes/admin.conf', 'failed'),
    ('CIS 1.1.13', 'chmod', '0644', '/etc/kubernetes/admin.conf', 'failed'),
    # todo:
    ('CIS 1.1.13', 'chmod', '0600', '/etc/kubernetes/admin.conf', 'passed'),
]

cis_1_1_14 = [
    ('CIS 1.1.14', 'chown', 'root:daemon', '/etc/kubernetes/admin.conf', 'failed'),
    ('CIS 1.1.14', 'chown', 'daemon:root', '/etc/kubernetes/admin.conf', 'failed'),
    ('CIS 1.1.14', 'chown', 'daemon:daemon', '/etc/kubernetes/admin.conf', 'failed'),
    ('CIS 1.1.14', 'chown', 'root:root', '/etc/kubernetes/admin.conf', 'passed'),
]

cis_1_1_15 = [
    ('CIS 1.1.15', 'chmod', '0700', '/etc/kubernetes/scheduler.conf', 'failed'),
    ('CIS 1.1.15', 'chmod', '0644', '/etc/kubernetes/scheduler.conf', 'passed'),

]

cis_1_1_16 = [
    ('CIS 1.1.16', 'chown', 'root:daemon', '/etc/kubernetes/scheduler.conf', 'failed'),
    ('CIS 1.1.16', 'chown', 'daemon:root', '/etc/kubernetes/scheduler.conf', 'failed'),
    ('CIS 1.1.16', 'chown', 'daemon:daemon', '/etc/kubernetes/scheduler.conf', 'failed'),
    ('CIS 1.1.16', 'chown', 'root:root', '/etc/kubernetes/scheduler.conf', 'passed'),

]

cis_1_1_17 = [
    ('CIS 1.1.17', 'chmod', '0700', '/etc/kubernetes/controller-manager.conf', 'failed'),
    ('CIS 1.1.17', 'chmod', '0644', '/etc/kubernetes/controller-manager.conf', 'passed'),
]

cis_1_1_18 = [
    ('CIS 1.1.18', 'chown', 'root:daemon', '/etc/kubernetes/controller-manager.conf', 'failed'),
    ('CIS 1.1.18', 'chown', 'daemon:root', '/etc/kubernetes/controller-manager.conf', 'failed'),
    ('CIS 1.1.18', 'chown', 'daemon:daemon', '/etc/kubernetes/controller-manager.conf', 'failed'),
    ('CIS 1.1.18', 'chown', 'root:root', '/etc/kubernetes/controller-manager.conf', 'passed'),

]

cis_1_1_19 = [
    ('CIS 1.1.19', 'chown', 'root:daemon', '/etc/kubernetes/pki/', 'failed'),
    ('CIS 1.1.19', 'chown', 'root:root', '/etc/kubernetes/pki/', 'passed'),
    ('CIS 1.1.19', 'chown', 'root:root', '/etc/kubernetes/pki/some_file.txt', 'passed'),
    ('CIS 1.1.19', 'chown', 'daemon:root', '/etc/kubernetes/pki/', 'failed'),
    ('CIS 1.1.19', 'chown', 'daemon:daemon', '/etc/kubernetes/pki/', 'failed'),
    ('CIS 1.1.19', 'chown', 'root:daemon', '/etc/kubernetes/pki/some_file.txt', 'failed'),

]

cis_1_1_20 = [
    ('CIS 1.1.20', 'chmod', '0700', '/etc/kubernetes/pki/apiserver.crt', 'failed'),
    ('CIS 1.1.20', 'chmod', '0666', '/etc/kubernetes/pki/apiserver.crt', 'failed'),
    ('CIS 1.1.20', 'chmod', '0644', '/etc/kubernetes/pki/apiserver.crt', 'passed'),
]

cis_1_1_21 = [
    ('CIS 1.1.21', 'chmod', '0644', '/etc/kubernetes/pki/apiserver.key', 'failed'),
    ('CIS 1.1.21', 'chmod', '0600', '/etc/kubernetes/pki/apiserver.key', 'passed'),
]

cis_4_1_1 = [
    ('CIS 4.1.1', 'chmod', '0700', '/etc/systemd/system/kubelet.service.d/10-kubeadm.conf', 'failed'),
    ('CIS 4.1.1', 'chmod', '0644', '/etc/systemd/system/kubelet.service.d/10-kubeadm.conf', 'passed'),

]

cis_4_1_2 = [
    ('CIS 4.1.2', 'chown', 'root:daemon', '/etc/systemd/system/kubelet.service.d/10-kubeadm.conf', 'failed'),
    ('CIS 4.1.2', 'chown', 'daemon:root', '/etc/systemd/system/kubelet.service.d/10-kubeadm.conf', 'failed'),
    ('CIS 4.1.2', 'chown', 'daemon:daemon', '/etc/systemd/system/kubelet.service.d/10-kubeadm.conf', 'failed'),
    ('CIS 4.1.2', 'chown', 'root:root', '/etc/systemd/system/kubelet.service.d/10-kubeadm.conf', 'passed'),
]

cis_4_1_5 = [
    ('CIS 4.1.5', 'chmod', '0700', '/etc/kubernetes/kubelet.conf', 'failed'),
    ('CIS 4.1.5', 'chmod', '0644', '/etc/kubernetes/kubelet.conf', 'passed'),
]

cis_4_1_6 = [
    ('CIS 4.1.6', 'chown', 'root:daemon', '/etc/kubernetes/kubelet.conf', 'failed'),
    ('CIS 4.1.6', 'chown', 'daemon:root', '/etc/kubernetes/kubelet.conf', 'failed'),
    ('CIS 4.1.6', 'chown', 'daemon:daemon', '/etc/kubernetes/kubelet.conf', 'failed'),
    ('CIS 4.1.6', 'chown', 'root:root', '/etc/kubernetes/kubelet.conf', 'passed'),
]

cis_4_1_6 = [
    ('CIS 4.1.6', 'chown', 'root:daemon', '/etc/kubernetes/kubelet.conf', 'failed'),
    ('CIS 4.1.6', 'chown', 'daemon:root', '/etc/kubernetes/kubelet.conf', 'failed'),
    ('CIS 4.1.6', 'chown', 'daemon:daemon', '/etc/kubernetes/kubelet.conf', 'failed'),
    ('CIS 4.1.6', 'chown', 'root:root', '/etc/kubernetes/kubelet.conf', 'passed'),
]

cis_4_1_9 = [
    ('CIS 4.1.9', 'chmod', '0700', '/var/lib/kubelet/config.yaml', 'failed'),
    ('CIS 4.1.9', 'chmod', '0644', '/var/lib/kubelet/config.yaml', 'passed'),

]

cis_4_1_10 = [
    ('CIS 4.1.10', 'chown', 'root:daemon', '/etc/kubernetes/kubelet.conf', 'failed'),
    ('CIS 4.1.10', 'chown', 'daemon:root', '/etc/kubernetes/kubelet.conf', 'failed'),
    ('CIS 4.1.10', 'chown', 'daemon:daemon', '/etc/kubernetes/kubelet.conf', 'failed'),
    ('CIS 4.1.10', 'chown', 'root:root', '/etc/kubernetes/kubelet.conf', 'passed'),
]
