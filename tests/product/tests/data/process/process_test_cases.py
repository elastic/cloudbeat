"""
This module contains process test cases definition.
Each rule is list of tuples
Rule test case is defined as tuple of data
"""
from commonlib.framework.reporting import skip_param_case, SkipReportData

cis_1_2_4 = [(
    'CIS 1.2.4',
    {
        "set": {
            "--kubelet-https": "false",
        },
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.4',
        {
            "set": {
                "--kubelet-https": "true",
            },
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
),
    (
        'CIS 1.2.4',
        {
            "unset": [
                "--kubelet-https"
            ]
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
)]

cis_2_1 = [(
    'CIS 2.1',
    {
        "set": {
            "--cert-file": "/etc/kubernetes/pki/etcd/server.crt",
            "--key-file": "/etc/kubernetes/pki/etcd/server.key"
        }
    },
    '/etc/kubernetes/manifests/etcd.yaml',
    'passed'
)]

cis_2_2 = [(
    'CIS 2.2',
    {
        "unset": [
            "--client-cert-auth"
        ]
    },
    '/etc/kubernetes/manifests/etcd.yaml',
    'failed'
),
    (
        'CIS 2.2',
        {
            "set": {
                "--client-cert-auth": "false"
            }
        },
        '/etc/kubernetes/manifests/etcd.yaml',
        'failed'
),
    (
        'CIS 2.2',
        {
            "set": {
                "--client-cert-auth": "true"
            }
        },
        '/etc/kubernetes/manifests/etcd.yaml',
        'passed'
)]

cis_2_3 = [(
    'CIS 2.3',
    {
        "set": {
            "--auto-tls": "false"
        }
    },
    '/etc/kubernetes/manifests/etcd.yaml',
    'passed'
),
    (
        'CIS 2.3',
        {
            "set": {
                "--auto-tls": "true"
            }
        },
        '/etc/kubernetes/manifests/etcd.yaml',
        'failed'
),
    (
        'CIS 2.3',
        {
            "unset": [
                "--auto-tls"
            ]
        },
        '/etc/kubernetes/manifests/etcd.yaml',
        'passed'
)]

cis_2_4 = [(
    'CIS 2.4',
    {
        "set": {
            "--peer-cert-file": "/etc/kubernetes/pki/etcd/peer.crt",
            "--peer-key-file": "/etc/kubernetes/pki/etcd/peer.key"
        }
    },
    '/etc/kubernetes/manifests/etcd.yaml',
    'passed'
)]

cis_2_5 = [(
    'CIS 2.5',
    {
        "unset": [
            "--peer-client-cert-auth"
        ]
    },
    '/etc/kubernetes/manifests/etcd.yaml',
    'failed'
),
    (
        'CIS 2.5',
        {
            "set": {
                "--peer-client-cert-auth": "false"
            }
        },
        '/etc/kubernetes/manifests/etcd.yaml',
        'failed'
),
    (
        'CIS 2.5',
        {
            "set": {
                "--peer-client-cert-auth": "true"
            }
        },
        '/etc/kubernetes/manifests/etcd.yaml',
        'passed'
)]

cis_2_6 = [
    (
        'CIS 2.6',
        {
            "set": {
                "--peer-auto-tls": "false"
            }
        },
        '/etc/kubernetes/manifests/etcd.yaml',
        'passed'
    ),
    (
        'CIS 2.6',
        {
            "set": {
                "--peer-auto-tls": "true"
            }
        },
        '/etc/kubernetes/manifests/etcd.yaml',
        'failed'
    ),
    (
        'CIS 2.6',
        {
            "unset": [
                "--peer-auto-tls"
            ]
        },
        '/etc/kubernetes/manifests/etcd.yaml',
        'passed'
    )]

cis_1_4_1 = [(
    'CIS 1.4.1',
    {
        "set": {
            "--profiling": "true"
        }
    },
    '/etc/kubernetes/manifests/kube-scheduler.yaml',
    'failed'
),
    (
        'CIS 1.4.1',
        {
            "unset": [
                "--profiling"
            ]
        },
        '/etc/kubernetes/manifests/kube-scheduler.yaml',
        'failed'
),
    (
        'CIS 1.4.1',
        {
            "set": {
                "--profiling": "false"
            }
        },
        '/etc/kubernetes/manifests/kube-scheduler.yaml',
        'passed'
)]

cis_1_4_2 = [(
    'CIS 1.4.2',
    {
        "set": {
            "--bind-address": "0.0.0.0"
        }
    },
    '/etc/kubernetes/manifests/kube-scheduler.yaml',
    'failed'
),
    (
        'CIS 1.4.2',
        {
            "unset": [
                "--bind-address"
            ]
        },
        '/etc/kubernetes/manifests/kube-scheduler.yaml',
        'failed'
),
    (
        'CIS 1.4.2',
        {
            "set": {
                "--bind-address": "127.0.0.1"
            }
        },
        '/etc/kubernetes/manifests/kube-scheduler.yaml',
        'passed'
)]

cis_1_3_2 = [(
    'CIS 1.3.2',
    {
        "set": {
            "--profiling": "true"
        }
    },
    '/etc/kubernetes/manifests/kube-controller-manager.yaml',
    'failed'
),
    (
        'CIS 1.3.2',
        {
            "unset": [
                "--profiling"
            ]
        },
        '/etc/kubernetes/manifests/kube-controller-manager.yaml',
        'failed'
),
    (
        'CIS 1.3.2',
        {
            "set": {
                "--profiling": "false"
            }
        },
        '/etc/kubernetes/manifests/kube-controller-manager.yaml',
        'passed'
)]

cis_1_3_3 = [(
    'CIS 1.3.3',
    {
        "set": {
            "--use-service-account-credentials": "false"
        }
    },
    '/etc/kubernetes/manifests/kube-controller-manager.yaml',
    'failed'
),
    (
        'CIS 1.3.3',
        {
            "unset": [
                "--use-service-account-credentials"
            ]
        },
        '/etc/kubernetes/manifests/kube-controller-manager.yaml',
        'failed'
),
    (
        'CIS 1.3.3',
        {
            "set": {
                "--use-service-account-credentials": "true"
            }
        },
        '/etc/kubernetes/manifests/kube-controller-manager.yaml',
        'passed'
)]

cis_1_3_4 = [(
    'CIS 1.3.4',
    {
        "unset": [
            "--use-service-account-credentials"
        ]
    },
    '/etc/kubernetes/manifests/kube-controller-manager.yaml',
    'passed'
)]

cis_1_3_5 = [(
    'CIS 1.3.5',
    {
        "unset": [
            "--root-ca-file"
        ]
    },
    '/etc/kubernetes/manifests/kube-controller-manager.yaml',
    'failed'
)]

cis_1_3_6 = [(
    'CIS 1.3.6',
    {
        "set": {
            "--feature-gates": "RotateKubeletServerCertificate=false"
        }
    },
    '/etc/kubernetes/manifests/kube-controller-manager.yaml',
    'failed'
),
    (
        'CIS 1.3.6',
        {
            "unset": [
                "--feature-gates"
            ]
        },
        '/etc/kubernetes/manifests/kube-controller-manager.yaml',
        'failed'
),
    (
        'CIS 1.3.6',
        {
            "set": {
                "--feature-gates": "RotateKubeletServerCertificate=true"
            }
        },
        '/etc/kubernetes/manifests/kube-controller-manager.yaml',
        'passed'
)]

cis_1_3_7 = [(
    'CIS 1.3.7',
    {
        "set": {
            "--bind-address": "0.0.0.0"
        }
    },
    '/etc/kubernetes/manifests/kube-controller-manager.yaml',
    'failed'
),
    (
        'CIS 1.3.7',
        {
            "unset": [
                "--bind-address"
            ]
        },
        '/etc/kubernetes/manifests/kube-controller-manager.yaml',
        'failed'
),
    (
        'CIS 1.3.7',
        {
            "set": {
                "--bind-address": "127.0.0.1"
            }
        },
        '/etc/kubernetes/manifests/kube-controller-manager.yaml',
        'passed'
)]

cis_1_2_2 = [(
    'CIS 1.2.2',
    {
        "unset": [
            "--token-auth-file"
        ]
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'passed'
)]

cis_1_2_3 = [(
    'CIS 1.2.3',
    {
        "unset": [
            "--DenyServiceExternalIPs"
        ]
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'passed'
)]

cis_1_2_5 = [(
    'CIS 1.2.5',
    {
        "set": {
            "--kubelet-client-certificate": "/etc/kubernetes/pki/apiserver-kubelet-client.crt ",
            "--kubelet-client-key": "/etc/kubernetes/pki/apiserver-kubelet-client.key"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'passed'
)]

cis_1_2_6 = [(
    'CIS 1.2.6',
    {
        "unset": [
            "--kubelet-certificate-authority"
        ]
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
)]

cis_1_2_7 = [(
    'CIS 1.2.7',
    {
        "set": {
            "--authorization-mode": "AlwaysAllow"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.7',
        {
            "unset": [
                "--authorization-mode"
            ]
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
),
    (
        'CIS 1.2.7',
        {
            "set": {
                "--authorization-mode": "Node,RBAC"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
)]

cis_1_2_8 = [(
    'CIS 1.2.8',
    {
        "unset": [
            "--authorization-mode"
        ]
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.8',
        {
            "set": {
                "--authorization-mode": "Node,RBAC"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
)]

cis_1_2_9 = [(
    'CIS 1.2.9',
    {
        "set": {
            "--authorization-mode": "Node"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.9',
        {
            "unset": [
                "--authorization-mode"
            ]
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'failed'
),
    (
        'CIS 1.2.9',
        {
            "set": {
                "--authorization-mode": "Node,RBAC"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
)]

cis_1_2_10 = [(
    'CIS 1.2.10',
    {
        "unset": [
            "--enable-admission-plugins"
        ]
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.10',
        {
            "set": {
                "--enable-admission-plugins": "EventRateLimit"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
)]

cis_1_2_11 = [(
    'CIS 1.2.11',
    {
        "set": {
            "--enable-admission-plugins": "AlwaysAdmit"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.11',
        {
            "unset": [
                "--enable-admission-plugins"
            ]
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
),
    (
        'CIS 1.2.11',
        {
            "set": {
                "--enable-admission-plugins": "NodeRestriction"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
)]

cis_1_2_12 = [(
    'CIS 1.2.12',
    {
        "unset": [
            "--enable-admission-plugins"
        ]
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.12',
        {
            "set": {
                "--enable-admission-plugins": "AlwaysPullImages"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
)]

cis_1_2_13 = [(
    'CIS 1.2.13',
    {
        "set": {
            "--enable-admission-plugins": "AlwaysDeny"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.13',
        {
            "set": {
                "--enable-admission-plugins": "SecurityContextDeny"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
),
    (
        'CIS 1.2.13',
        {
            "set": {
                "--enable-admission-plugins": "PodSecurityPolicy"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
)]

cis_1_2_14 = [(
    'CIS 1.2.14',
    {
        "set": {
            "--disable-admission-plugins": "ServiceAccount"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.14',
        {
            "unset": [
                "--disable-admission-plugins"
            ]
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'failed'
)]

cis_1_2_15 = [(
    'CIS 1.2.15',
    {
        "set": {
            "--disable-admission-plugins": "NamespaceLifecycle"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.15',
        {
            "unset": [
                "--disable-admission-plugins"
            ]
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'failed'
)]

cis_1_2_16 = [(
    'CIS 1.2.16',
    {
        "unset": [
            "--enable-admission-plugins"
        ]
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.16',
        {
            "set": {
                "--enable-admission-plugins": "NodeRestriction"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
)]

cis_1_2_17 = [(
    'CIS 1.2.17',
    {
        "unset": [
            "--secure-port"
        ]
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'passed'
),
    (
        'CIS 1.2.17',
        {
            "set": {
                "--secure-port": "260492"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'failed'
),
    (
        'CIS 1.2.17',
        {
            "set": {
                "--secure-port": "6443"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
)]

cis_1_2_18 = [(
    'CIS 1.2.18',
    {
        "set": {
            "--profiling": "true"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.18',
        {
            "set": {
                "--profiling": "false"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
),
    (
        'CIS 1.2.18',
        {
            "unset": [
                "--profiling"
            ]
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'failed'
)]

cis_1_2_19 = [(
    'CIS 1.2.19',
    {
        "unset": [
            "--audit-log-path"
        ]
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
)]

cis_1_2_20 = [(
    'CIS 1.2.20',
    {
        "set": {
            "--audit-log-maxage": "260492"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'passed'
),
    (
        'CIS 1.2.20',
        {
            "set": {
                "--audit-log-maxage": "30"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
),
    (
        'CIS 1.2.20',
        {
            "unset": [
                "--audit-log-maxage"
            ]
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'failed'
)]

cis_1_2_21 = [(
    'CIS 1.2.21',
    {
        "set": {
            "--audit-log-maxbackup": "-1"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.21',
        {
            "set": {
                "--audit-log-maxbackup": "10"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
),
    (
        'CIS 1.2.21',
        {
            "unset": [
                "--audit-log-maxbackup"
            ]
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'failed'
)]

cis_1_2_22 = [(
    'CIS 1.2.22',
    {
        "set": {
            "--audit-log-maxsize": "-1"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.22',
        {
            "set": {
                "--audit-log-maxsize": "100"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
),
    (
        'CIS 1.2.22',
        {
            "unset": [
                "--audit-log-maxsize"
            ]
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'failed'
)]

cis_1_2_23 = [(
    'CIS 1.2.23',
    {
        "set": {
            "--request-timeout": "-1s"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.23',
        {
            "set": {
                "--request-timeout": "300s"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
),
    (
        'CIS 1.2.23',
        {
            "unset": [
                "--request-timeout"
            ]
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
)]

cis_1_2_24 = [(
    'CIS 1.2.24',
    {
        "set": {
            "--service-account-lookup": "false"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1.2.24',
        {
            "set": {
                "--service-account-lookup": "true"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
),
    (
        'CIS 1.2.24',
        {
            "unset": [
                "--service-account-lookup"
            ]
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
)]

cis_1_2_25 = [(
    'CIS 1.2.25',
    {
        "set": {
            "--service-account-key-file": "/etc/kubernetes/pki/sa.pub"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'passed'
)]

cis_1_2_26 = [(
    'CIS 1.2.26',
    {
        "set": {
            "--etcd-certfile": "/etc/kubernetes/pki/apiserver-etcd-client.crt",
            "--etcd-keyfile": "/etc/kubernetes/pki/apiserver-etcd-client.key"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'passed'
)]

cis_1_2_27 = [(
    'CIS 1.2.27',
    {
        "set": {
            "--tls-cert-file": "/etc/kubernetes/pki/apiserver.crt",
            "--tls-private-key-file": "/etc/kubernetes/pki/apiserver.key"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'passed'
)]

cis_1_2_28 = [(
    'CIS 1.2.28',
    {
        "set": {
            "--client-ca-file": "/etc/kubernetes/pki/ca.crt"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'passed'
)]

cis_1_2_29 = [(
    'CIS 1.2.29',
    {
        "set": {
            "--etcd-cafile": "/etc/kubernetes/pki/etcd/ca.crt"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'passed'
)]

cis_1_2_32 = [(
    'CIS 1_2_32',
    {
        "set": {
            "--tls-cipher-suites": "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_DUMMY"
        }
    },
    '/etc/kubernetes/manifests/kube-apiserver.yaml',
    'failed'
),
    (
        'CIS 1_2_32',
        {
            "set": {
                "--tls-cipher-suites": "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
),
    (
        'CIS 1_2_32',
        {
            "set": {
                "--tls-cipher-suites":
                    "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
            }
        },
        '/etc/kubernetes/manifests/kube-apiserver.yaml',
        'passed'
)]

cis_4_2_1 = [(
    'CIS 4.2.1',
    {
        "set": {
            "authentication": {
                "anonymous": {
                    "enabled": True
                }
            }
        },
    },
    '/var/lib/kubelet/config.yaml',
    'failed'
),
    (
        'CIS 4.2.1',
        {
            "set": {
                "authentication": {
                    "anonymous": {
                        "enabled": False
                    }
                }
            }
        },
        '/var/lib/kubelet/config.yaml',
        'passed'
)]

cis_4_2_2 = [(
    'CIS 4.2.2',
    {
        "set": {
            "authorization": {
                "mode": "AlwaysAllow"
            }
        }
    },
    '/var/lib/kubelet/config.yaml',
    'failed'
),
    (
        'CIS 4.2.2',
        {
            "set": {
                "authorization": {
                    "mode": "Webhook"
                }
            }
        },
        '/var/lib/kubelet/config.yaml',
        'passed'
)]

cis_4_2_3 = [(
    'CIS 4.2.3',
    {
        "unset": ["authentication.x509.clientCAFile"]
    },
    '/var/lib/kubelet/config.yaml',
    'failed'
)]

cis_4_2_4 = [(
    'CIS 4.2.4',
    {
        "set": {
            "readOnlyPort": 26492
        }
    },
    '/var/lib/kubelet/config.yaml',
    'failed'
),
    (
        'CIS 4.2.4',
        {
            "set": {
                "readOnlyPort": 0
            }
        },
        '/var/lib/kubelet/config.yaml',
        'passed'
)]

cis_4_2_5 = [(
    'CIS 4.2.5',
    {
        "set": {
            "streamingConnectionIdleTimeout": 0
        }
    },
    '/var/lib/kubelet/config.yaml',
    'failed'
),
    (
        'CIS 4.2.5',
        {
            "set": {
                "streamingConnectionIdleTimeout": "26492s"
            }
        },
        '/var/lib/kubelet/config.yaml',
        'passed'
)]

cis_4_2_6 = [(
    'CIS 4.2.6',
    {
        "set": {
            "protectKernelDefaults": False
        }
    },
    '/var/lib/kubelet/config.yaml',
    'failed'
),
    (
        'CIS 4.2.6',
        {
            "set": {
                "protectKernelDefaults": True
            }
        },
        '/var/lib/kubelet/config.yaml',
        'passed'
)]

cis_4_2_7 = [(
    'CIS 4.2.7',
    {
        "set": {
            "makeIPTablesUtilChains": False
        }
    },
    '/var/lib/kubelet/config.yaml',
    'failed'
),
    (
        'CIS 4.2.7',
        {
            "set": {
                "makeIPTablesUtilChains": True
            }
        },
        '/var/lib/kubelet/config.yaml',
        'passed'
)]

cis_4_2_9 = [(
    'CIS 4.2.9',
    {
        "set": {
            "eventRecordQPS": 4
        }
    },
    '/var/lib/kubelet/config.yaml',
    'failed'
),
    (
        'CIS 4.2.9',
        {
            "set": {
                "eventRecordQPS": 0
            }
        },
        '/var/lib/kubelet/config.yaml',
        'passed'
)]

cis_4_2_10 = [(
    'CIS 4.2.10',
    {
        "set": {
            "tlsCertFile": "",
            "tlsPrivateKeyFile": ""
        }
    },
    '/var/lib/kubelet/config.yaml',
    'passed'
)]

cis_4_2_11 = [(
    'CIS 4.2.11',
    {
        "set": {
            "rotateCertificates": False
        }
    },
    '/var/lib/kubelet/config.yaml',
    'failed'
),
    (
        'CIS 4.2.11',
        {
            "set": {
                "rotateCertificates": True
            }
        },
        '/var/lib/kubelet/config.yaml',
        'passed'
)]

cis_4_2_12 = [
    # TODO test case should fail instead of pass
    # (
    #     'CIS 4.2.12',
    #     {
    #         "set": {
    #             "serverTLSBootstrap": False,
    #             "featureGates": {
    #                 "RotateKubeletServerCertificate": False
    #             }
    #         }
    #     },
    #     '/var/lib/kubelet/config.yaml',
    #     'failed'
    # ),
    (
        'CIS 4.2.12',
        {
            "set": {
                "serverTLSBootstrap": False,
                "featureGates": {
                    "RotateKubeletServerCertificate": True
                }
            }
        },
        '/var/lib/kubelet/config.yaml',
        'passed'
    ),
    (
        'CIS 4.2.12',
        {
            "set": {
                "serverTLSBootstrap": True,
                "featureGates": {
                    "RotateKubeletServerCertificate": False
                }
            }
        },
        '/var/lib/kubelet/config.yaml',
        'passed'
    )]

cis_4_2_13 = [(
    'CIS 4.2.13',
    {
        "set": {
            "TLSCipherSuites": ["TLS_ECDHE_ECDSA_WITH_AES_128_GCM_DUMMY"]
        }
    },
    '/var/lib/kubelet/config.yaml',
    'failed'
),
    (
        'CIS 4.2.13',
        {
            "set": {
                "TLSCipherSuites": [
                    "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
                ]
            }
        },
        '/var/lib/kubelet/config.yaml',
        'passed'
),
    (
        'CIS 4.2.13',
        {
            "set": {
                "TLSCipherSuites": [
                    "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
                    "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
                    "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
                    "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
                    "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
                    "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
                    "TLS_RSA_WITH_AES_256_GCM_SHA384",
                    "TLS_RSA_WITH_AES_128_GCM_SHA256"
                ]
            }
        },
        '/var/lib/kubelet/config.yaml',
        'passed'
)]

etcd_rules = [
    *cis_2_1,
    *cis_2_2,
    *cis_2_3,
    *cis_2_4,
    *cis_2_5,
    *cis_2_6,
]

api_server_rules = [
    *cis_1_2_2,
    *skip_param_case(skip_list=[*cis_1_2_3,
                                *cis_1_2_4,
                                *cis_1_2_5
                                ],
                     data_to_report=SkipReportData(
                         skip_reason="This case fails and breaks cluster")
                     ),
    *cis_1_2_6,
    *cis_1_2_7,
    *cis_1_2_8,
    *skip_param_case(skip_list=[*cis_1_2_9,
                                *cis_1_2_10
                                ],
                     data_to_report=SkipReportData(
                         skip_reason="This case fails and breaks cluster")
                     ),
    *cis_1_2_11,
    *cis_1_2_12,
    *cis_1_2_13,
    *cis_1_2_14,
    *cis_1_2_15,
    *cis_1_2_16,
    *skip_param_case(skip_list=[*cis_1_2_17],
                     data_to_report=SkipReportData(
                         skip_reason="This case fails and breaks cluster")
                     ),
    *cis_1_2_18,
    *cis_1_2_19,
    *cis_1_2_20,
    *cis_1_2_21,
    *cis_1_2_22,
    *skip_param_case(skip_list=[*cis_1_2_23],
                     data_to_report=SkipReportData(
                         skip_reason="This case fails and breaks cluster")
                     ),
    *cis_1_2_24,
    *cis_1_2_25,
    *cis_1_2_26,
    *cis_1_2_27,
    *cis_1_2_28,
    *cis_1_2_29,
    *skip_param_case(skip_list=[*cis_1_2_32],
                     data_to_report=SkipReportData(
                         skip_reason="This case fails and breaks cluster")
                     )
]

controller_manager_rules = [
    *cis_1_3_2,
    *cis_1_3_3,
    *cis_1_3_4,
    *cis_1_3_5,
    *cis_1_3_6,
    *cis_1_3_7,
]

scheduler_rules = [
    *cis_1_4_1,
    *cis_1_4_2,
]

kubelet_rules = [
    *skip_param_case(skip_list=[*cis_4_2_1,
                                *cis_4_2_2,
                                *cis_4_2_3,
                                *cis_4_2_4,
                                *cis_4_2_5,
                                *cis_4_2_6,
                                *cis_4_2_7,
                                *cis_4_2_9,
                                ],
                     data_to_report=SkipReportData(
                         skip_reason="Dangling tests"
                     )),
    # *cis_4_2_8, # TODO setting is not configurable via the Kubelet config file.
    *cis_4_2_10,
    *cis_4_2_11,
    *cis_4_2_12,  # TODO first test case should fail instead of pass
    *cis_4_2_13,
]
