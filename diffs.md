# file

## New

```json
{
  "_index": ".ds-logs-cloud_security_posture.findings-default-2022.09.01-000001",
  "_id": "opaODIMB5gnkGmib__E9",
  "_version": 1,
  "_score": 0,
  "_source": {
    "@timestamp": "2022-09-05T07:30:40.545Z",
    "message": "Rule \"Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated)\": passed",
    "file": {
      "uid": "0",
      "name": "pki",
      "mode": "755",
      "directory": "/hostfs/etc/kubernetes",
      "path": "/hostfs/etc/kubernetes/pki",
      "type": "directory",
      "ctime": "2022-09-05T07:23:20.310Z",
      "gid": "0",
      "accessed": "2022-09-05T07:24:40.385Z",
      "inode": "5524549",
      "group": "root",
      "owner": "root",
      "size": 4096,
      "mtime": "2022-09-05T07:23:20.310Z"
    },
    "resource": {
      "ecsFormat": "file",
      "raw": {
        "group": "root",
        "owner": "root",
        "gid": "0",
        "inode": "5524549",
        "mode": "755",
        "name": "pki",
        "path": "/hostfs/etc/kubernetes/pki",
        "sub_type": "directory",
        "uid": "0"
      },
      "id": "17dc5a20-edf9-5738-951d-5752813db09a",
      "type": "file",
      "sub_type": "directory",
      "name": "/hostfs/etc/kubernetes/pki"
    },
    "ecs": {
      "version": "8.0.0"
    },
    "type": "file",
    "result": {
      "evaluation": "passed",
      "expected": {
        "owner": "root",
        "group": "root"
      },
      "evidence": {
        "group": "root",
        "owner": "root"
      }
    },
    "rule": {
      "section": "Control Plane Node Configuration Files",
      "default_value": "By default, the `/etc/kubernetes/pki/` directory and all of the files and directories contained within it, are set to be owned by the root user.\n",
      "references": "1. [https://kubernetes.io/docs/admin/kube-apiserver/](https://kubernetes.io/docs/admin/kube-apiserver/)\n",
      "tags": [
        "CIS",
        "Kubernetes",
        "CIS 1.1.19",
        "Control Plane Node Configuration Files"
      ],
      "version": "1.0",
      "profile_applicability": "* Level 1 - Master Node\n",
      "description": "Ensure that the Kubernetes PKI directory and file ownership is set to `root:root`.\n",
      "remediation": "Run the below command (based on the file location on your system) on the\ncontrol plane node.\nFor example,\n```\nchown -R root:root /etc/kubernetes/pki/\n```\n",
      "impact": "None\n",
      "id": "780ac02f-e0f5-537c-98ba-354ae5873a81",
      "audit": "Run the below command (based on the file location on your system) on the\ncontrol plane node.\nFor example,\n```\nls -laR /etc/kubernetes/pki/\n```\nVerify that the ownership of all files and directories in this hierarchy is set to `root:root`.\n",
      "benchmark": {
        "id": "cis_k8s",
        "name": "CIS Kubernetes V1.23",
        "version": "v1.0.0"
      },
      "name": "Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated)",
      "rationale": "Kubernetes makes use of a number of certificates as part of its operation. You should set the ownership of the directory containing the PKI information and all files in that directory to maintain their integrity. The directory and files should be owned by root:root.\n"
    },
    "event": {
      "outcome": "success",
      "type": [
        "info"
      ],
      "sequence": 1662363040,
      "created": "2022-09-05T07:30:40.544Z",
      "id": "7ad79339-93e2-44bd-925f-c941a4df069e",
      "kind": "state",
      "category": [
        "configuration"
      ]
    },
    "agent": {
      "type": "cloudbeat",
      "version": "8.5.0",
      "ephemeral_id": "1b86fa91-10e9-4f9b-bc6a-a50c3cf37f8e",
      "id": "41467e74-babc-4580-95f8-60ae97770294",
      "name": "kind-mono-control-plane"
    },
    "host": {
      "os": {
        "kernel": "5.10.104-linuxkit",
        "codename": "bullseye",
        "type": "linux",
        "platform": "debian",
        "version": "11 (bullseye)",
        "family": "debian",
        "name": "Debian GNU/Linux"
      },
      "containerized": false,
      "ip": [
        "10.244.0.1",
        "10.244.0.1",
        "10.244.0.1",
        "172.18.0.2",
        "fc00:f853:ccd:e793::2",
        "fe80::42:acff:fe12:2",
        "172.19.0.2"
      ],
      "name": "kind-mono-control-plane",
      "mac": [
        "02:42:ac:12:00:02",
        "02:42:ac:13:00:02",
        "1e:79:00:2d:a0:f8",
        "92:1f:b6:41:7d:43",
        "fe:76:90:6b:d2:f1"
      ],
      "hostname": "kind-mono-control-plane",
      "architecture": "aarch64"
    },
    "cluster_id": "851b737d-c31e-4c3b-9e76-86d3805b9d6f",
    "resource_id": "17dc5a20-edf9-5738-951d-5752813db09a"
  },
  "fields": {
    "file.mode": [
      "755"
    ],
    "rule.id": [
      "780ac02f-e0f5-537c-98ba-354ae5873a81"
    ],
    "file.path": [
      "/hostfs/etc/kubernetes/pki"
    ],
    "resource.raw.inode": [
      "5524549"
    ],
    "result.evaluation": [
      "passed"
    ],
    "event.category": [
      "configuration"
    ],
    "file.group": [
      "root"
    ],
    "result.evidence.group": [
      "root"
    ],
    "host.hostname": [
      "kind-mono-control-plane"
    ],
    "type": [
      "file"
    ],
    "host.mac": [
      "02:42:ac:12:00:02",
      "02:42:ac:13:00:02",
      "1e:79:00:2d:a0:f8",
      "92:1f:b6:41:7d:43",
      "fe:76:90:6b:d2:f1"
    ],
    "rule.profile_applicability": [
      "* Level 1 - Master Node\n"
    ],
    "resource.raw.owner": [
      "root"
    ],
    "resource.sub_type": [
      "directory"
    ],
    "host.os.version": [
      "11 (bullseye)"
    ],
    "file.mtime": [
      "2022-09-05T07:23:20.310Z"
    ],
    "resource.raw.gid": [
      "0"
    ],
    "host.os.name": [
      "Debian GNU/Linux"
    ],
    "agent.name": [
      "kind-mono-control-plane"
    ],
    "rule.rationale": [
      "Kubernetes makes use of a number of certificates as part of its operation. You should set the ownership of the directory containing the PKI information and all files in that directory to maintain their integrity. The directory and files should be owned by root:root.\n"
    ],
    "host.name": [
      "kind-mono-control-plane"
    ],
    "event.kind": [
      "state"
    ],
    "resource.raw.name": [
      "pki"
    ],
    "event.outcome": [
      "success"
    ],
    "resource.raw.uid": [
      "0"
    ],
    "host.os.type": [
      "linux"
    ],
    "rule.name": [
      "Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated)"
    ],
    "rule.impact": [
      "None\n"
    ],
    "rule.default_value": [
      "By default, the `/etc/kubernetes/pki/` directory and all of the files and directories contained within it, are set to be owned by the root user.\n"
    ],
    "rule.description": [
      "Ensure that the Kubernetes PKI directory and file ownership is set to `root:root`.\n"
    ],
    "resource.type": [
      "file"
    ],
    "data_stream.type": [
      "logs"
    ],
    "rule.references": [
      "1. [https://kubernetes.io/docs/admin/kube-apiserver/](https://kubernetes.io/docs/admin/kube-apiserver/)\n"
    ],
    "host.architecture": [
      "aarch64"
    ],
    "rule.audit": [
      "Run the below command (based on the file location on your system) on the\ncontrol plane node.\nFor example,\n```\nls -laR /etc/kubernetes/pki/\n```\nVerify that the ownership of all files and directories in this hierarchy is set to `root:root`.\n"
    ],
    "file.type": [
      "directory"
    ],
    "agent.id": [
      "41467e74-babc-4580-95f8-60ae97770294"
    ],
    "host.containerized": [
      false
    ],
    "ecs.version": [
      "8.0.0"
    ],
    "event.created": [
      "2022-09-05T07:30:40.544Z"
    ],
    "file.owner": [
      "root"
    ],
    "agent.version": [
      "8.5.0"
    ],
    "host.os.family": [
      "debian"
    ],
    "rule.benchmark.name": [
      "CIS Kubernetes V1.23"
    ],
    "resource.raw.sub_type": [
      "directory"
    ],
    "resource.name": [
      "/hostfs/etc/kubernetes/pki"
    ],
    "file.gid": [
      "0"
    ],
    "file.uid": [
      "0"
    ],
    "resource.ecsFormat": [
      "file"
    ],
    "rule.tags": [
      "CIS",
      "Kubernetes",
      "CIS 1.1.19",
      "Control Plane Node Configuration Files"
    ],
    "event.sequence": [
      1662363010
    ],
    "host.ip": [
      "10.244.0.1",
      "10.244.0.1",
      "10.244.0.1",
      "172.18.0.2",
      "fc00:f853:ccd:e793::2",
      "fe80::42:acff:fe12:2",
      "172.19.0.2"
    ],
    "cluster_id": [
      "851b737d-c31e-4c3b-9e76-86d3805b9d6f"
    ],
    "result.expected.owner": [
      "root"
    ],
    "agent.type": [
      "cloudbeat"
    ],
    "host.os.kernel": [
      "5.10.104-linuxkit"
    ],
    "file.accessed": [
      "2022-09-05T07:24:40.385Z"
    ],
    "file.inode": [
      "5524549"
    ],
    "result.evidence.owner": [
      "root"
    ],
    "file.directory": [
      "/hostfs/etc/kubernetes"
    ],
    "resource.raw.mode": [
      "755"
    ],
    "file.name": [
      "pki"
    ],
    "host.os.codename": [
      "bullseye"
    ],
    "file.size": [
      4096
    ],
    "rule.remediation": [
      "Run the below command (based on the file location on your system) on the\ncontrol plane node.\nFor example,\n```\nchown -R root:root /etc/kubernetes/pki/\n```\n"
    ],
    "message": [
      "Rule \"Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated)\": passed"
    ],
    "rule.version": [
      "1.0"
    ],
    "file.ctime": [
      "2022-09-05T07:23:20.310Z"
    ],
    "resource.id": [
      "17dc5a20-edf9-5738-951d-5752813db09a"
    ],
    "rule.section": [
      "Control Plane Node Configuration Files"
    ],
    "resource.raw.group": [
      "root"
    ],
    "rule.benchmark.id": [
      "cis_k8s"
    ],
    "result.expected.group": [
      "root"
    ],
    "@timestamp": [
      "2022-09-05T07:30:40.545Z"
    ],
    "host.os.platform": [
      "debian"
    ],
    "event.type": [
      "info"
    ],
    "resource_id": [
      "17dc5a20-edf9-5738-951d-5752813db09a"
    ],
    "agent.ephemeral_id": [
      "1b86fa91-10e9-4f9b-bc6a-a50c3cf37f8e"
    ],
    "resource.raw.path": [
      "/hostfs/etc/kubernetes/pki"
    ],
    "event.id": [
      "7ad79339-93e2-44bd-925f-c941a4df069e"
    ],
    "rule.benchmark.version": [
      "v1.0.0"
    ]
  }
}
```

## Old

```json
{
  "_index": ".ds-logs-cloud_security_posture.findings-default-2022.09.01-000001",
  "_id": "EZeWDIMB5gnkGmibMw-E",
  "_version": 1,
  "_score": 0,
  "_source": {
    "@timestamp": "2022-09-05T07:38:32.697Z",
    "message": "Rule \"Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated)\": passed",
    "host": {
      "containerized": false,
      "ip": [
        "10.244.0.1",
        "10.244.0.1",
        "10.244.0.1",
        "172.18.0.2",
        "fc00:f853:ccd:e793::2",
        "fe80::42:acff:fe12:2",
        "172.19.0.2"
      ],
      "name": "kind-mono-control-plane",
      "mac": [
        "02:42:ac:12:00:02",
        "02:42:ac:13:00:02",
        "1e:79:00:2d:a0:f8",
        "92:1f:b6:41:7d:43",
        "fe:76:90:6b:d2:f1"
      ],
      "hostname": "kind-mono-control-plane",
      "architecture": "aarch64",
      "os": {
        "kernel": "5.10.104-linuxkit",
        "codename": "bullseye",
        "type": "linux",
        "platform": "debian",
        "version": "11 (bullseye)",
        "family": "debian",
        "name": "Debian GNU/Linux"
      }
    },
    "agent": {
      "type": "cloudbeat",
      "version": "8.5.0",
      "ephemeral_id": "b7a60101-e450-4085-9514-fed2b48eeaa9",
      "id": "ac066b7d-887d-4f0f-a562-5c6e2cc1ce07",
      "name": "kind-mono-control-plane"
    },
    "cluster_id": "851b737d-c31e-4c3b-9e76-86d3805b9d6f",
    "event": {
      "kind": "state",
      "sequence": 1662363512,
      "outcome": "success",
      "type": [
        "info"
      ],
      "category": [
        "configuration"
      ],
      "created": "2022-09-05T07:38:32.697191047Z",
      "id": "7462a937-3449-4f51-a8d4-830ae2e00c56"
    },
    "resource": {
      "id": "17dc5a20-edf9-5738-951d-5752813db09a",
      "type": "file",
      "sub_type": "directory",
      "name": "/hostfs/etc/kubernetes/pki",
      "ecsFormat": "file",
      "raw": {
        "inode": "5524549",
        "mode": "755",
        "owner": "root",
        "path": "/hostfs/etc/kubernetes/pki",
        "sub_type": "directory",
        "gid": "0",
        "group": "root",
        "name": "pki",
        "uid": "0"
      }
    },
    "result": {
      "evaluation": "passed",
      "expected": {
        "group": "root",
        "owner": "root"
      },
      "evidence": {
        "group": "root",
        "owner": "root"
      }
    },
    "rule": {
      "remediation": "Run the below command (based on the file location on your system) on the\ncontrol plane node.\nFor example,\n```\nchown -R root:root /etc/kubernetes/pki/\n```\n",
      "impact": "None\n",
      "tags": [
        "CIS",
        "Kubernetes",
        "CIS 1.1.19",
        "Control Plane Node Configuration Files"
      ],
      "id": "780ac02f-e0f5-537c-98ba-354ae5873a81",
      "audit": "Run the below command (based on the file location on your system) on the\ncontrol plane node.\nFor example,\n```\nls -laR /etc/kubernetes/pki/\n```\nVerify that the ownership of all files and directories in this hierarchy is set to `root:root`.\n",
      "default_value": "By default, the `/etc/kubernetes/pki/` directory and all of the files and directories contained within it, are set to be owned by the root user.\n",
      "name": "Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated)",
      "description": "Ensure that the Kubernetes PKI directory and file ownership is set to `root:root`.\n",
      "version": "1.0",
      "rationale": "Kubernetes makes use of a number of certificates as part of its operation. You should set the ownership of the directory containing the PKI information and all files in that directory to maintain their integrity. The directory and files should be owned by root:root.\n",
      "section": "Control Plane Node Configuration Files",
      "benchmark": {
        "id": "cis_k8s",
        "name": "CIS Kubernetes V1.23",
        "version": "v1.0.0"
      },
      "profile_applicability": "* Level 1 - Master Node\n",
      "references": "1. [https://kubernetes.io/docs/admin/kube-apiserver/](https://kubernetes.io/docs/admin/kube-apiserver/)\n"
    },
    "resource_id": "17dc5a20-edf9-5738-951d-5752813db09a",
    "type": "file",
    "file": {
      "accessed": "2022-09-05T07:24:40.385827011Z",
      "mode": "755",
      "group": "root",
      "directory": "/hostfs/etc/kubernetes",
      "mtime": "2022-09-05T07:23:20.310278004Z",
      "uid": "0",
      "owner": "root",
      "inode": "5524549",
      "ctime": "2022-09-05T07:23:20.310278004Z",
      "size": 4096,
      "type": "directory",
      "name": "pki",
      "gid": "0",
      "path": "/hostfs/etc/kubernetes/pki"
    },
    "ecs": {
      "version": "8.0.0"
    }
  },
  "fields": {
    "file.mode": [
      "755"
    ],
    "rule.id": [
      "780ac02f-e0f5-537c-98ba-354ae5873a81"
    ],
    "file.path": [
      "/hostfs/etc/kubernetes/pki"
    ],
    "resource.raw.inode": [
      "5524549"
    ],
    "result.evaluation": [
      "passed"
    ],
    "event.category": [
      "configuration"
    ],
    "file.group": [
      "root"
    ],
    "result.evidence.group": [
      "root"
    ],
    "host.hostname": [
      "kind-mono-control-plane"
    ],
    "type": [
      "file"
    ],
    "host.mac": [
      "02:42:ac:12:00:02",
      "02:42:ac:13:00:02",
      "1e:79:00:2d:a0:f8",
      "92:1f:b6:41:7d:43",
      "fe:76:90:6b:d2:f1"
    ],
    "rule.profile_applicability": [
      "* Level 1 - Master Node\n"
    ],
    "resource.raw.owner": [
      "root"
    ],
    "resource.sub_type": [
      "directory"
    ],
    "host.os.version": [
      "11 (bullseye)"
    ],
    "file.mtime": [
      "2022-09-05T07:23:20.310278004Z"
    ],
    "resource.raw.gid": [
      "0"
    ],
    "host.os.name": [
      "Debian GNU/Linux"
    ],
    "agent.name": [
      "kind-mono-control-plane"
    ],
    "rule.rationale": [
      "Kubernetes makes use of a number of certificates as part of its operation. You should set the ownership of the directory containing the PKI information and all files in that directory to maintain their integrity. The directory and files should be owned by root:root.\n"
    ],
    "host.name": [
      "kind-mono-control-plane"
    ],
    "event.kind": [
      "state"
    ],
    "resource.raw.name": [
      "pki"
    ],
    "event.outcome": [
      "success"
    ],
    "resource.raw.uid": [
      "0"
    ],
    "host.os.type": [
      "linux"
    ],
    "rule.name": [
      "Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated)"
    ],
    "rule.impact": [
      "None\n"
    ],
    "rule.default_value": [
      "By default, the `/etc/kubernetes/pki/` directory and all of the files and directories contained within it, are set to be owned by the root user.\n"
    ],
    "rule.description": [
      "Ensure that the Kubernetes PKI directory and file ownership is set to `root:root`.\n"
    ],
    "resource.type": [
      "file"
    ],
    "data_stream.type": [
      "logs"
    ],
    "rule.references": [
      "1. [https://kubernetes.io/docs/admin/kube-apiserver/](https://kubernetes.io/docs/admin/kube-apiserver/)\n"
    ],
    "host.architecture": [
      "aarch64"
    ],
    "rule.audit": [
      "Run the below command (based on the file location on your system) on the\ncontrol plane node.\nFor example,\n```\nls -laR /etc/kubernetes/pki/\n```\nVerify that the ownership of all files and directories in this hierarchy is set to `root:root`.\n"
    ],
    "file.type": [
      "directory"
    ],
    "agent.id": [
      "ac066b7d-887d-4f0f-a562-5c6e2cc1ce07"
    ],
    "host.containerized": [
      false
    ],
    "ecs.version": [
      "8.0.0"
    ],
    "event.created": [
      "2022-09-05T07:38:32.697191047Z"
    ],
    "file.owner": [
      "root"
    ],
    "agent.version": [
      "8.5.0"
    ],
    "host.os.family": [
      "debian"
    ],
    "rule.benchmark.name": [
      "CIS Kubernetes V1.23"
    ],
    "resource.raw.sub_type": [
      "directory"
    ],
    "resource.name": [
      "/hostfs/etc/kubernetes/pki"
    ],
    "file.gid": [
      "0"
    ],
    "file.uid": [
      "0"
    ],
    "resource.ecsFormat": [
      "file"
    ],
    "rule.tags": [
      "CIS",
      "Kubernetes",
      "CIS 1.1.19",
      "Control Plane Node Configuration Files"
    ],
    "event.sequence": [
      1662363520
    ],
    "host.ip": [
      "10.244.0.1",
      "10.244.0.1",
      "10.244.0.1",
      "172.18.0.2",
      "fc00:f853:ccd:e793::2",
      "fe80::42:acff:fe12:2",
      "172.19.0.2"
    ],
    "cluster_id": [
      "851b737d-c31e-4c3b-9e76-86d3805b9d6f"
    ],
    "result.expected.owner": [
      "root"
    ],
    "agent.type": [
      "cloudbeat"
    ],
    "host.os.kernel": [
      "5.10.104-linuxkit"
    ],
    "file.accessed": [
      "2022-09-05T07:24:40.385827011Z"
    ],
    "file.inode": [
      "5524549"
    ],
    "result.evidence.owner": [
      "root"
    ],
    "file.directory": [
      "/hostfs/etc/kubernetes"
    ],
    "resource.raw.mode": [
      "755"
    ],
    "file.name": [
      "pki"
    ],
    "host.os.codename": [
      "bullseye"
    ],
    "file.size": [
      4096
    ],
    "rule.remediation": [
      "Run the below command (based on the file location on your system) on the\ncontrol plane node.\nFor example,\n```\nchown -R root:root /etc/kubernetes/pki/\n```\n"
    ],
    "message": [
      "Rule \"Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated)\": passed"
    ],
    "rule.version": [
      "1.0"
    ],
    "file.ctime": [
      "2022-09-05T07:23:20.310278004Z"
    ],
    "resource.id": [
      "17dc5a20-edf9-5738-951d-5752813db09a"
    ],
    "rule.section": [
      "Control Plane Node Configuration Files"
    ],
    "resource.raw.group": [
      "root"
    ],
    "rule.benchmark.id": [
      "cis_k8s"
    ],
    "result.expected.group": [
      "root"
    ],
    "@timestamp": [
      "2022-09-05T07:38:32.697Z"
    ],
    "host.os.platform": [
      "debian"
    ],
    "event.type": [
      "info"
    ],
    "resource_id": [
      "17dc5a20-edf9-5738-951d-5752813db09a"
    ],
    "agent.ephemeral_id": [
      "b7a60101-e450-4085-9514-fed2b48eeaa9"
    ],
    "resource.raw.path": [
      "/hostfs/etc/kubernetes/pki"
    ],
    "event.id": [
      "7462a937-3449-4f51-a8d4-830ae2e00c56"
    ],
    "rule.benchmark.version": [
      "v1.0.0"
    ]
  }
}
```

# process

## New

```json
{
  "_index": ".ds-logs-cloud_security_posture.findings-default-2022.09.01-000001",
  "_id": "x5aMDIMB5gnkGmibP-f6",
  "_version": 1,
  "_score": 0,
  "_source": {
    "@timestamp": "2022-09-05T07:27:41.112Z",
    "result": {
      "evaluation": "failed",
      "evidence": {
        "process_args": {
          "--cgroup-root": "/kubelet",
          "--provider-id": "kind://docker/kind-mono/kind-mono-control-plane",
          "--container-runtime-endpoint": "unix:///run/containerd/containerd.sock",
          "--fail-swap-on": "false",
          "--node-ip": "172.18.0.2",
          "--node-labels": "",
          "--/usr/bin/kubelet": "",
          "--container-runtime": "remote",
          "--bootstrap-kubeconfig": "/etc/kubernetes/bootstrap-kubelet.conf",
          "--config": "/var/lib/kubelet/config.yaml",
          "--kubeconfig": "/etc/kubernetes/kubelet.conf",
          "--pod-infra-container-image": "k8s.gcr.io/pause:3.7"
        }
      }
    },
    "rule": {
      "version": "1.0",
      "benchmark": {
        "id": "cis_k8s",
        "name": "CIS Kubernetes V1.23",
        "version": "v1.0.0"
      },
      "audit": "Run the following command on each node:\n```\nps -ef | grep kubelet\n```\nVerify that the `--tls-cert-file` and `--tls-private-key-file` arguments exist and they are set as\nappropriate.\nIf these arguments are not present, check that there is a Kubelet config specified by `--config`\nand that it contains appropriate settings for tlsCertFile and tlsPrivateKeyFile.\n",
      "remediation": "If using a Kubelet config file, edit the file to set tlsCertFile to the location of the certificate\nfile to use to identify this Kubelet, and tlsPrivateKeyFile to the location of the\ncorresponding private key file.\nIf using command line arguments, edit the kubelet service file\n`/etc/systemd/system/kubelet.service.d/10-kubeadm.conf` on each worker node and set\nthe below parameters in KUBELET_CERTIFICATE_ARGS variable.\n--tls-cert-file=<path/to/tls-certificate-file> --tls-private-key-file=<path/to/tls-key-file>\nBased on your system, restart the kubelet service. For example:\n```\nsystemctl daemon-reload\nsystemctl restart kubelet.service\n```\n",
      "tags": [
        "CIS",
        "Kubernetes",
        "CIS 4.2.10",
        "Kubelet"
      ],
      "rationale": "The connections from the apiserver to the kubelet are used for fetching logs for pods,\nattaching (through kubectl) to running pods, and using the kubelet’s port-forwarding\nfunctionality. These connections terminate at the kubelet’s HTTPS endpoint. By default, the\napiserver does not verify the kubelet’s serving certificate, which makes the connection\nsubject to man-in-the-middle attacks, and unsafe to run over untrusted and/or public\nnetworks.\n",
      "section": "Kubelet",
      "name": "Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual)",
      "profile_applicability": "* Level 1 - Worker Node\n",
      "id": "dc91f4c4-4f0e-59ba-a0e1-96e996736787,",
      "description": "Setup TLS connection on the Kubelets.\n"
    },
    "event": {
      "kind": "state",
      "category": [
        "configuration"
      ],
      "outcome": "success",
      "type": [
        "info"
      ],
      "sequence": 1662362860,
      "created": "2022-09-05T07:27:41.112Z",
      "id": "a558b781-6ae0-4bad-9e4f-df24e3efbf7a"
    },
    "resource_id": "584d3461-8fab-53ca-8ef6-04e26d0c2162",
    "ecs": {
      "version": "8.0.0"
    },
    "host": {
      "name": "kind-mono-control-plane",
      "containerized": false,
      "ip": [
        "10.244.0.1",
        "10.244.0.1",
        "10.244.0.1",
        "172.18.0.2",
        "fc00:f853:ccd:e793::2",
        "fe80::42:acff:fe12:2",
        "172.19.0.2"
      ],
      "mac": [
        "02:42:ac:12:00:02",
        "02:42:ac:13:00:02",
        "1e:79:00:2d:a0:f8",
        "92:1f:b6:41:7d:43",
        "fe:76:90:6b:d2:f1"
      ],
      "hostname": "kind-mono-control-plane",
      "architecture": "aarch64",
      "os": {
        "family": "debian",
        "name": "Debian GNU/Linux",
        "kernel": "5.10.104-linuxkit",
        "codename": "bullseye",
        "type": "linux",
        "platform": "debian",
        "version": "11 (bullseye)"
      }
    },
    "agent": {
      "id": "41467e74-babc-4580-95f8-60ae97770294",
      "name": "kind-mono-control-plane",
      "type": "cloudbeat",
      "version": "8.5.0",
      "ephemeral_id": "1b86fa91-10e9-4f9b-bc6a-a50c3cf37f8e"
    },
    "cluster_id": "851b737d-c31e-4c3b-9e76-86d3805b9d6f",
    "message": "Rule \"Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual)\": failed",
    "process": {
      "uptime": 249,
      "parent": {
        "PID": 1,
        "CommandLine": "",
        "Args": null,
        "Name": "",
        "PGID": 0,
        "Title": "",
        "ThreadName": "",
        "Uptime": 0,
        "WorkingDirectory": "",
        "EntityID": "",
        "ArgsCount": 0,
        "Start": "0001-01-01T00:00:00Z",
        "Parent": null,
        "Executable": "",
        "ThreadID": 0,
        "ExitCode": 0,
        "End": "0001-01-01T00:00:00Z"
      },
      "name": "kubelet",
      "pid": 705,
      "command_line": "/usr/bin/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock --node-ip=172.18.0.2 --node-labels= --pod-infra-container-image=k8s.gcr.io/pause:3.7 --provider-id=kind://docker/kind-mono/kind-mono-control-plane --fail-swap-on=false --cgroup-root=/kubelet",
      "args_count": 12,
      "args": [
        "/usr/bin/kubelet",
        "--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf",
        "--kubeconfig=/etc/kubernetes/kubelet.conf",
        "--config=/var/lib/kubelet/config.yaml",
        "--container-runtime=remote",
        "--container-runtime-endpoint=unix:///run/containerd/containerd.sock",
        "--node-ip=172.18.0.2",
        "--node-labels=",
        "--pod-infra-container-image=k8s.gcr.io/pause:3.7",
        "--provider-id=kind://docker/kind-mono/kind-mono-control-plane",
        "--fail-swap-on=false",
        "--cgroup-root=/kubelet"
      ],
      "title": "kubelet",
      "start": "2022-09-05T07:23:30.946Z",
      "pgid": 705
    },
    "resource": {
      "raw": {
        "stat": {
          "EffectiveUID": "",
          "EffectiveGID": "",
          "RealGID": "",
          "TotalSize": "2021400000",
          "RealUID": "",
          "Parent": "1",
          "Threads": "18",
          "Nice": "0",
          "SavedUID": "",
          "StartTime": "338857",
          "SavedGID": "",
          "State": "S",
          "Group": "705",
          "Name": "kubelet",
          "UserTime": "370",
          "SystemTime": "516",
          "ResidentSize": "76820000"
        },
        "command": "/usr/bin/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock --node-ip=172.18.0.2 --node-labels= --pod-infra-container-image=k8s.gcr.io/pause:3.7 --provider-id=kind://docker/kind-mono/kind-mono-control-plane --fail-swap-on=false --cgroup-root=/kubelet",
        "external_data": {
          "config": {
            "nodeStatusReportFrequency": "0s",
            "staticPodPath": "/etc/kubernetes/manifests",
            "shutdownGracePeriodCriticalPods": "0s",
            "evictionHard": {
              "nodefs.available": "0%",
              "nodefs.inodesFree": "0%",
              "imagefs.available": "0%"
            },
            "clusterDNS": [
              "10.96.0.10"
            ],
            "authorization": {
              "mode": "Webhook",
              "webhook": {
                "cacheUnauthorizedTTL": "0s",
                "cacheAuthorizedTTL": "0s"
              }
            },
            "fileCheckFrequency": "0s",
            "logging": {
              "flushFrequency": "0",
              "options": {
                "json": {
                  "infoBufferSize": "0"
                }
              },
              "verbosity": "0"
            },
            "shutdownGracePeriod": "0s",
            "cgroupRoot": "/kubelet",
            "evictionPressureTransitionPeriod": "0s",
            "rotateCertificates": true,
            "authentication": {
              "anonymous": {
                "enabled": false
              },
              "webhook": {
                "cacheTTL": "0s",
                "enabled": true
              },
              "x509": {
                "clientCAFile": "/etc/kubernetes/pki/ca.crt"
              }
            },
            "cpuManagerReconcilePeriod": "0s",
            "syncFrequency": "0s",
            "failSwapOn": false,
            "imageMinimumGCAge": "0s",
            "clusterDomain": "cluster.local",
            "imageGCHighThresholdPercent": "100",
            "healthzBindAddress": "127.0.0.1",
            "nodeStatusUpdateFrequency": "0s",
            "streamingConnectionIdleTimeout": "0s",
            "cgroupDriver": "systemd",
            "memorySwap": {},
            "healthzPort": "10248",
            "runtimeRequestTimeout": "0s",
            "httpCheckFrequency": "0s",
            "kind": "KubeletConfiguration",
            "volumeStatsAggPeriod": "0s",
            "apiVersion": "kubelet.config.k8s.io/v1beta1"
          }
        },
        "pid": "705"
      },
      "id": "584d3461-8fab-53ca-8ef6-04e26d0c2162",
      "type": "process",
      "sub_type": "process",
      "name": "kubelet",
      "ecsFormat": "process"
    },
    "type": "process"
  },
  "fields": {
    "rule.id": [
      "dc91f4c4-4f0e-59ba-a0e1-96e996736787,"
    ],
    "resource.raw.external_data.config.imageGCHighThresholdPercent": [
      100
    ],
    "event.category": [
      "configuration"
    ],
    "resource.raw.external_data.config.authorization.webhook.cacheAuthorizedTTL": [
      "0s"
    ],
    "host.hostname": [
      "kind-mono-control-plane"
    ],
    "type": [
      "process"
    ],
    "host.mac": [
      "02:42:ac:12:00:02",
      "02:42:ac:13:00:02",
      "1e:79:00:2d:a0:f8",
      "92:1f:b6:41:7d:43",
      "fe:76:90:6b:d2:f1"
    ],
    "resource.raw.stat.EffectiveUID": [
      ""
    ],
    "process.parent.Executable": [
      ""
    ],
    "rule.profile_applicability": [
      "* Level 1 - Worker Node\n"
    ],
    "result.evidence.process_args.--fail-swap-on": [
      "false"
    ],
    "resource.raw.external_data.config.imageMinimumGCAge": [
      "0s"
    ],
    "resource.sub_type": [
      "process"
    ],
    "resource.raw.external_data.config.authentication.anonymous.enabled": [
      false
    ],
    "host.os.version": [
      "11 (bullseye)"
    ],
    "resource.raw.stat.UserTime": [
      "370"
    ],
    "agent.name": [
      "kind-mono-control-plane"
    ],
    "rule.rationale": [
      "The connections from the apiserver to the kubelet are used for fetching logs for pods,\nattaching (through kubectl) to running pods, and using the kubelet’s port-forwarding\nfunctionality. These connections terminate at the kubelet’s HTTPS endpoint. By default, the\napiserver does not verify the kubelet’s serving certificate, which makes the connection\nsubject to man-in-the-middle attacks, and unsafe to run over untrusted and/or public\nnetworks.\n"
    ],
    "resource.raw.stat.RealGID": [
      ""
    ],
    "resource.raw.external_data.config.logging.flushFrequency": [
      0
    ],
    "resource.raw.external_data.config.staticPodPath": [
      "/etc/kubernetes/manifests"
    ],
    "event.outcome": [
      "success"
    ],
    "resource.raw.external_data.config.kind": [
      "KubeletConfiguration"
    ],
    "host.os.type": [
      "linux"
    ],
    "resource.raw.external_data.config.syncFrequency": [
      "0s"
    ],
    "resource.raw.external_data.config.clusterDomain": [
      "cluster.local"
    ],
    "resource.raw.external_data.config.nodeStatusReportFrequency": [
      "0s"
    ],
    "resource.raw.external_data.config.logging.options.json.infoBufferSize": [
      "0"
    ],
    "resource.raw.external_data.config.authorization.mode": [
      "Webhook"
    ],
    "process.parent.PID": [
      1
    ],
    "resource.raw.external_data.config.volumeStatsAggPeriod": [
      "0s"
    ],
    "host.architecture": [
      "aarch64"
    ],
    "agent.id": [
      "41467e74-babc-4580-95f8-60ae97770294"
    ],
    "host.containerized": [
      false
    ],
    "resource.raw.external_data.config.runtimeRequestTimeout": [
      "0s"
    ],
    "resource.raw.external_data.config.httpCheckFrequency": [
      "0s"
    ],
    "process.parent.ArgsCount": [
      0
    ],
    "resource.raw.stat.Group": [
      "705"
    ],
    "process.parent.Name": [
      ""
    ],
    "resource.raw.stat.StartTime": [
      "338857"
    ],
    "result.evidence.process_args.--bootstrap-kubeconfig": [
      "/etc/kubernetes/bootstrap-kubelet.conf"
    ],
    "resource.raw.external_data.config.evictionPressureTransitionPeriod": [
      "0s"
    ],
    "process.parent.Start": [
      "0001-01-01T00:00:00Z"
    ],
    "rule.tags": [
      "CIS",
      "Kubernetes",
      "CIS 4.2.10",
      "Kubelet"
    ],
    "host.ip": [
      "10.244.0.1",
      "10.244.0.1",
      "10.244.0.1",
      "172.18.0.2",
      "fc00:f853:ccd:e793::2",
      "fe80::42:acff:fe12:2",
      "172.19.0.2"
    ],
    "agent.type": [
      "cloudbeat"
    ],
    "process.parent.WorkingDirectory": [
      ""
    ],
    "resource.raw.external_data.config.nodeStatusUpdateFrequency": [
      "0s"
    ],
    "result.evidence.process_args.--node-labels": [
      ""
    ],
    "resource.raw.stat.TotalSize": [
      "2021400000"
    ],
    "process.uptime": [
      249
    ],
    "process.parent.PGID": [
      0
    ],
    "resource.raw.external_data.config.healthzBindAddress": [
      "127.0.0.1"
    ],
    "process.parent.Uptime": [
      0
    ],
    "resource.raw.external_data.config.apiVersion": [
      "kubelet.config.k8s.io/v1beta1"
    ],
    "result.evidence.process_args.--container-runtime-endpoint": [
      "unix:///run/containerd/containerd.sock"
    ],
    "process.parent.ThreadID": [
      0
    ],
    "resource.raw.stat.SavedUID": [
      ""
    ],
    "result.evidence.process_args.--/usr/bin/kubelet": [
      ""
    ],
    "host.os.codename": [
      "bullseye"
    ],
    "resource.raw.external_data.config.failSwapOn": [
      false
    ],
    "process.parent.ThreadName": [
      ""
    ],
    "process.pgid": [
      705
    ],
    "resource.raw.pid": [
      "705"
    ],
    "resource.id": [
      "584d3461-8fab-53ca-8ef6-04e26d0c2162"
    ],
    "rule.section": [
      "Kubelet"
    ],
    "process.parent.Title": [
      ""
    ],
    "result.evidence.process_args.--container-runtime": [
      "remote"
    ],
    "@timestamp": [
      "2022-09-05T07:27:41.112Z"
    ],
    "host.os.platform": [
      "debian"
    ],
    "result.evidence.process_args.--node-ip": [
      "172.18.0.2"
    ],
    "resource.raw.external_data.config.authentication.x509.clientCAFile": [
      "/etc/kubernetes/pki/ca.crt"
    ],
    "resource.raw.command": [
      "/usr/bin/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock --node-ip=172.18.0.2 --node-labels= --pod-infra-container-image=k8s.gcr.io/pause:3.7 --provider-id=kind://docker/kind-mono/kind-mono-control-plane --fail-swap-on=false --cgroup-root=/kubelet"
    ],
    "resource_id": [
      "584d3461-8fab-53ca-8ef6-04e26d0c2162"
    ],
    "agent.ephemeral_id": [
      "1b86fa91-10e9-4f9b-bc6a-a50c3cf37f8e"
    ],
    "result.evidence.process_args.--pod-infra-container-image": [
      "k8s.gcr.io/pause:3.7"
    ],
    "event.id": [
      "a558b781-6ae0-4bad-9e4f-df24e3efbf7a"
    ],
    "result.evidence.process_args.--kubeconfig": [
      "/etc/kubernetes/kubelet.conf"
    ],
    "process.parent.End": [
      "0001-01-01T00:00:00Z"
    ],
    "rule.benchmark.version": [
      "v1.0.0"
    ],
    "result.evaluation": [
      "failed"
    ],
    "resource.raw.external_data.config.rotateCertificates": [
      true
    ],
    "resource.raw.stat.State": [
      "S"
    ],
    "result.evidence.process_args.--cgroup-root": [
      "/kubelet"
    ],
    "process.pid": [
      705
    ],
    "resource.raw.external_data.config.healthzPort": [
      10248
    ],
    "process.parent.CommandLine": [
      ""
    ],
    "resource.raw.external_data.config.cgroupRoot": [
      "/kubelet"
    ],
    "resource.raw.stat.Threads": [
      "18"
    ],
    "resource.raw.stat.RealUID": [
      ""
    ],
    "host.os.name": [
      "Debian GNU/Linux"
    ],
    "process.parent.ExitCode": [
      0
    ],
    "resource.raw.external_data.config.shutdownGracePeriodCriticalPods": [
      "0s"
    ],
    "host.name": [
      "kind-mono-control-plane"
    ],
    "resource.raw.external_data.config.evictionHard.nodefs.inodesFree": [
      "0%"
    ],
    "resource.raw.stat.EffectiveGID": [
      ""
    ],
    "event.kind": [
      "state"
    ],
    "rule.name": [
      "Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual)"
    ],
    "resource.raw.stat.SystemTime": [
      "516"
    ],
    "rule.description": [
      "Setup TLS connection on the Kubelets.\n"
    ],
    "resource.type": [
      "process"
    ],
    "data_stream.type": [
      "logs"
    ],
    "process.name": [
      "kubelet"
    ],
    "resource.raw.external_data.config.fileCheckFrequency": [
      "0s"
    ],
    "rule.audit": [
      "Run the following command on each node:\n```\nps -ef | grep kubelet\n```\nVerify that the `--tls-cert-file` and `--tls-private-key-file` arguments exist and they are set as\nappropriate.\nIf these arguments are not present, check that there is a Kubelet config specified by `--config`\nand that it contains appropriate settings for tlsCertFile and tlsPrivateKeyFile.\n"
    ],
    "ecs.version": [
      "8.0.0"
    ],
    "resource.raw.stat.Name": [
      "kubelet"
    ],
    "event.created": [
      "2022-09-05T07:27:41.112Z"
    ],
    "resource.raw.external_data.config.evictionHard.nodefs.available": [
      "0%"
    ],
    "agent.version": [
      "8.5.0"
    ],
    "resource.raw.external_data.config.logging.verbosity": [
      0
    ],
    "host.os.family": [
      "debian"
    ],
    "process.title": [
      "kubelet"
    ],
    "process.start": [
      "2022-09-05T07:23:30.946Z"
    ],
    "rule.benchmark.name": [
      "CIS Kubernetes V1.23"
    ],
    "result.evidence.process_args.--config": [
      "/var/lib/kubelet/config.yaml"
    ],
    "resource.raw.external_data.config.streamingConnectionIdleTimeout": [
      "0s"
    ],
    "resource.name": [
      "kubelet"
    ],
    "resource.raw.external_data.config.authentication.webhook.enabled": [
      true
    ],
    "resource.ecsFormat": [
      "process"
    ],
    "resource.raw.external_data.config.authentication.webhook.cacheTTL": [
      "0s"
    ],
    "event.sequence": [
      1662362880
    ],
    "cluster_id": [
      "851b737d-c31e-4c3b-9e76-86d3805b9d6f"
    ],
    "host.os.kernel": [
      "5.10.104-linuxkit"
    ],
    "resource.raw.external_data.config.shutdownGracePeriod": [
      "0s"
    ],
    "resource.raw.stat.Nice": [
      "0"
    ],
    "resource.raw.external_data.config.authorization.webhook.cacheUnauthorizedTTL": [
      "0s"
    ],
    "resource.raw.stat.Parent": [
      "1"
    ],
    "resource.raw.external_data.config.cgroupDriver": [
      "systemd"
    ],
    "process.args_count": [
      12
    ],
    "process.args": [
      "/usr/bin/kubelet",
      "--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf",
      "--kubeconfig=/etc/kubernetes/kubelet.conf",
      "--config=/var/lib/kubelet/config.yaml",
      "--container-runtime=remote",
      "--container-runtime-endpoint=unix:///run/containerd/containerd.sock",
      "--node-ip=172.18.0.2",
      "--node-labels=",
      "--pod-infra-container-image=k8s.gcr.io/pause:3.7",
      "--provider-id=kind://docker/kind-mono/kind-mono-control-plane",
      "--fail-swap-on=false",
      "--cgroup-root=/kubelet"
    ],
    "rule.remediation": [
      "If using a Kubelet config file, edit the file to set tlsCertFile to the location of the certificate\nfile to use to identify this Kubelet, and tlsPrivateKeyFile to the location of the\ncorresponding private key file.\nIf using command line arguments, edit the kubelet service file\n`/etc/systemd/system/kubelet.service.d/10-kubeadm.conf` on each worker node and set\nthe below parameters in KUBELET_CERTIFICATE_ARGS variable.\n--tls-cert-file=<path/to/tls-certificate-file> --tls-private-key-file=<path/to/tls-key-file>\nBased on your system, restart the kubelet service. For example:\n```\nsystemctl daemon-reload\nsystemctl restart kubelet.service\n```\n"
    ],
    "message": [
      "Rule \"Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual)\": failed"
    ],
    "resource.raw.stat.ResidentSize": [
      "76820000"
    ],
    "rule.version": [
      "1.0"
    ],
    "process.parent.EntityID": [
      ""
    ],
    "resource.raw.external_data.config.evictionHard.imagefs.available": [
      "0%"
    ],
    "resource.raw.external_data.config.clusterDNS": [
      "10.96.0.10"
    ],
    "result.evidence.process_args.--provider-id": [
      "kind://docker/kind-mono/kind-mono-control-plane"
    ],
    "rule.benchmark.id": [
      "cis_k8s"
    ],
    "resource.raw.stat.SavedGID": [
      ""
    ],
    "event.type": [
      "info"
    ],
    "process.command_line": [
      "/usr/bin/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock --node-ip=172.18.0.2 --node-labels= --pod-infra-container-image=k8s.gcr.io/pause:3.7 --provider-id=kind://docker/kind-mono/kind-mono-control-plane --fail-swap-on=false --cgroup-root=/kubelet"
    ],
    "resource.raw.external_data.config.cpuManagerReconcilePeriod": [
      "0s"
    ]
  }
}
```

## Old

```json
{
  "_index": ".ds-logs-cloud_security_posture.findings-default-2022.09.01-000001",
  "_id": "KJeWDIMB5gnkGmibqBLi",
  "_version": 1,
  "_score": 0,
  "_source": {
    "@timestamp": "2022-09-05T07:39:03.401Z",
    "message": "Rule \"Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual)\": failed",
    "process": {
      "start": "2022-09-05T07:23:31.195892255Z",
      "uptime": 931,
      "pgid": 705,
      "args": [
        "/usr/bin/kubelet",
        "--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf",
        "--kubeconfig=/etc/kubernetes/kubelet.conf",
        "--config=/var/lib/kubelet/config.yaml",
        "--container-runtime=remote",
        "--container-runtime-endpoint=unix:///run/containerd/containerd.sock",
        "--node-ip=172.18.0.2",
        "--node-labels=",
        "--pod-infra-container-image=k8s.gcr.io/pause:3.7",
        "--provider-id=kind://docker/kind-mono/kind-mono-control-plane",
        "--fail-swap-on=false",
        "--cgroup-root=/kubelet"
      ],
      "args_count": 12,
      "title": "kubelet",
      "parent": {
        "pid": 1,
        "start": "0001-01-01T00:00:00Z"
      },
      "pid": 705,
      "name": "kubelet",
      "command_line": "/usr/bin/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock --node-ip=172.18.0.2 --node-labels= --pod-infra-container-image=k8s.gcr.io/pause:3.7 --provider-id=kind://docker/kind-mono/kind-mono-control-plane --fail-swap-on=false --cgroup-root=/kubelet"
    },
    "event": {
      "category": [
        "configuration"
      ],
      "created": "2022-09-05T07:39:03.401055006Z",
      "id": "ba5d285f-7ce3-4aa3-9a2a-5f1ab30c7ace",
      "kind": "state",
      "sequence": 1662363542,
      "outcome": "success",
      "type": [
        "info"
      ]
    },
    "resource_id": "584d3461-8fab-53ca-8ef6-04e26d0c2162",
    "type": "process",
    "host": {
      "name": "kind-mono-control-plane",
      "containerized": false,
      "ip": [
        "10.244.0.1",
        "10.244.0.1",
        "10.244.0.1",
        "172.18.0.2",
        "fc00:f853:ccd:e793::2",
        "fe80::42:acff:fe12:2",
        "172.19.0.2"
      ],
      "mac": [
        "02:42:ac:12:00:02",
        "02:42:ac:13:00:02",
        "1e:79:00:2d:a0:f8",
        "92:1f:b6:41:7d:43",
        "fe:76:90:6b:d2:f1"
      ],
      "hostname": "kind-mono-control-plane",
      "architecture": "aarch64",
      "os": {
        "kernel": "5.10.104-linuxkit",
        "codename": "bullseye",
        "type": "linux",
        "platform": "debian",
        "version": "11 (bullseye)",
        "family": "debian",
        "name": "Debian GNU/Linux"
      }
    },
    "agent": {
      "ephemeral_id": "b7a60101-e450-4085-9514-fed2b48eeaa9",
      "id": "ac066b7d-887d-4f0f-a562-5c6e2cc1ce07",
      "name": "kind-mono-control-plane",
      "type": "cloudbeat",
      "version": "8.5.0"
    },
    "cluster_id": "851b737d-c31e-4c3b-9e76-86d3805b9d6f",
    "rule": {
      "tags": [
        "CIS",
        "Kubernetes",
        "CIS 4.2.10",
        "Kubelet"
      ],
      "benchmark": {
        "name": "CIS Kubernetes V1.23",
        "version": "v1.0.0",
        "id": "cis_k8s"
      },
      "profile_applicability": "* Level 1 - Worker Node\n",
      "audit": "Run the following command on each node:\n```\nps -ef | grep kubelet\n```\nVerify that the `--tls-cert-file` and `--tls-private-key-file` arguments exist and they are set as\nappropriate.\nIf these arguments are not present, check that there is a Kubelet config specified by `--config`\nand that it contains appropriate settings for tlsCertFile and tlsPrivateKeyFile.\n",
      "impact": "",
      "section": "Kubelet",
      "id": "dc91f4c4-4f0e-59ba-a0e1-96e996736787,",
      "name": "Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual)",
      "remediation": "If using a Kubelet config file, edit the file to set tlsCertFile to the location of the certificate\nfile to use to identify this Kubelet, and tlsPrivateKeyFile to the location of the\ncorresponding private key file.\nIf using command line arguments, edit the kubelet service file\n`/etc/systemd/system/kubelet.service.d/10-kubeadm.conf` on each worker node and set\nthe below parameters in KUBELET_CERTIFICATE_ARGS variable.\n--tls-cert-file=<path/to/tls-certificate-file> --tls-private-key-file=<path/to/tls-key-file>\nBased on your system, restart the kubelet service. For example:\n```\nsystemctl daemon-reload\nsystemctl restart kubelet.service\n```\n",
      "default_value": "",
      "references": "",
      "version": "1.0",
      "description": "Setup TLS connection on the Kubelets.\n",
      "rationale": "The connections from the apiserver to the kubelet are used for fetching logs for pods,\nattaching (through kubectl) to running pods, and using the kubelet’s port-forwarding\nfunctionality. These connections terminate at the kubelet’s HTTPS endpoint. By default, the\napiserver does not verify the kubelet’s serving certificate, which makes the connection\nsubject to man-in-the-middle attacks, and unsafe to run over untrusted and/or public\nnetworks.\n"
    },
    "resource": {
      "id": "584d3461-8fab-53ca-8ef6-04e26d0c2162",
      "type": "process",
      "sub_type": "process",
      "name": "kubelet",
      "ecsFormat": "process",
      "raw": {
        "command": "/usr/bin/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock --node-ip=172.18.0.2 --node-labels= --pod-infra-container-image=k8s.gcr.io/pause:3.7 --provider-id=kind://docker/kind-mono/kind-mono-control-plane --fail-swap-on=false --cgroup-root=/kubelet",
        "external_data": {
          "config": {
            "clusterDNS": [
              "10.96.0.10"
            ],
            "cpuManagerReconcilePeriod": "0s",
            "apiVersion": "kubelet.config.k8s.io/v1beta1",
            "authorization": {
              "mode": "Webhook",
              "webhook": {
                "cacheAuthorizedTTL": "0s",
                "cacheUnauthorizedTTL": "0s"
              }
            },
            "failSwapOn": false,
            "fileCheckFrequency": "0s",
            "shutdownGracePeriodCriticalPods": "0s",
            "httpCheckFrequency": "0s",
            "imageGCHighThresholdPercent": 100,
            "imageMinimumGCAge": "0s",
            "shutdownGracePeriod": "0s",
            "syncFrequency": "0s",
            "healthzPort": 10248,
            "staticPodPath": "/etc/kubernetes/manifests",
            "streamingConnectionIdleTimeout": "0s",
            "authentication": {
              "anonymous": {
                "enabled": false
              },
              "webhook": {
                "cacheTTL": "0s",
                "enabled": true
              },
              "x509": {
                "clientCAFile": "/etc/kubernetes/pki/ca.crt"
              }
            },
            "cgroupRoot": "/kubelet",
            "evictionHard": {
              "imagefs.available": "0%",
              "nodefs.available": "0%",
              "nodefs.inodesFree": "0%"
            },
            "evictionPressureTransitionPeriod": "0s",
            "healthzBindAddress": "127.0.0.1",
            "memorySwap": {},
            "cgroupDriver": "systemd",
            "kind": "KubeletConfiguration",
            "logging": {
              "flushFrequency": 0,
              "options": {
                "json": {
                  "infoBufferSize": "0"
                }
              },
              "verbosity": 0
            },
            "rotateCertificates": true,
            "runtimeRequestTimeout": "0s",
            "clusterDomain": "cluster.local",
            "nodeStatusReportFrequency": "0s",
            "nodeStatusUpdateFrequency": "0s",
            "volumeStatsAggPeriod": "0s"
          }
        },
        "pid": "705",
        "stat": {
          "EffectiveUID": "",
          "Group": "705",
          "StartTime": "338857",
          "Threads": "19",
          "Name": "kubelet",
          "Parent": "1",
          "ResidentSize": "78320000",
          "State": "S",
          "EffectiveGID": "",
          "Nice": "0",
          "RealUID": "",
          "SavedGID": "",
          "SystemTime": "1863",
          "RealGID": "",
          "SavedUID": "",
          "TotalSize": "2095192000",
          "UserTime": "1359"
        }
      }
    },
    "result": {
      "evaluation": "failed",
      "expected": null,
      "evidence": {
        "process_args": {
          "--cgroup-root": "/kubelet",
          "--fail-swap-on": "false",
          "--node-ip": "172.18.0.2",
          "--provider-id": "kind://docker/kind-mono/kind-mono-control-plane",
          "--/usr/bin/kubelet": "",
          "--bootstrap-kubeconfig": "/etc/kubernetes/bootstrap-kubelet.conf",
          "--container-runtime-endpoint": "unix:///run/containerd/containerd.sock",
          "--kubeconfig": "/etc/kubernetes/kubelet.conf",
          "--node-labels": "",
          "--pod-infra-container-image": "k8s.gcr.io/pause:3.7",
          "--config": "/var/lib/kubelet/config.yaml",
          "--container-runtime": "remote"
        }
      }
    },
    "ecs": {
      "version": "8.0.0"
    }
  },
  "fields": {
    "rule.id": [
      "dc91f4c4-4f0e-59ba-a0e1-96e996736787,"
    ],
    "resource.raw.external_data.config.imageGCHighThresholdPercent": [
      100
    ],
    "event.category": [
      "configuration"
    ],
    "resource.raw.external_data.config.authorization.webhook.cacheAuthorizedTTL": [
      "0s"
    ],
    "process.parent.pid": [
      1
    ],
    "host.hostname": [
      "kind-mono-control-plane"
    ],
    "type": [
      "process"
    ],
    "host.mac": [
      "02:42:ac:12:00:02",
      "02:42:ac:13:00:02",
      "1e:79:00:2d:a0:f8",
      "92:1f:b6:41:7d:43",
      "fe:76:90:6b:d2:f1"
    ],
    "resource.raw.stat.EffectiveUID": [
      ""
    ],
    "rule.profile_applicability": [
      "* Level 1 - Worker Node\n"
    ],
    "result.evidence.process_args.--fail-swap-on": [
      "false"
    ],
    "resource.raw.external_data.config.imageMinimumGCAge": [
      "0s"
    ],
    "resource.sub_type": [
      "process"
    ],
    "resource.raw.external_data.config.authentication.anonymous.enabled": [
      false
    ],
    "host.os.version": [
      "11 (bullseye)"
    ],
    "resource.raw.stat.UserTime": [
      "1359"
    ],
    "agent.name": [
      "kind-mono-control-plane"
    ],
    "rule.rationale": [
      "The connections from the apiserver to the kubelet are used for fetching logs for pods,\nattaching (through kubectl) to running pods, and using the kubelet’s port-forwarding\nfunctionality. These connections terminate at the kubelet’s HTTPS endpoint. By default, the\napiserver does not verify the kubelet’s serving certificate, which makes the connection\nsubject to man-in-the-middle attacks, and unsafe to run over untrusted and/or public\nnetworks.\n"
    ],
    "resource.raw.stat.RealGID": [
      ""
    ],
    "resource.raw.external_data.config.logging.flushFrequency": [
      0
    ],
    "resource.raw.external_data.config.staticPodPath": [
      "/etc/kubernetes/manifests"
    ],
    "event.outcome": [
      "success"
    ],
    "resource.raw.external_data.config.kind": [
      "KubeletConfiguration"
    ],
    "host.os.type": [
      "linux"
    ],
    "process.parent.start": [
      "0001-01-01T00:00:00Z"
    ],
    "resource.raw.external_data.config.syncFrequency": [
      "0s"
    ],
    "resource.raw.external_data.config.clusterDomain": [
      "cluster.local"
    ],
    "resource.raw.external_data.config.nodeStatusReportFrequency": [
      "0s"
    ],
    "resource.raw.external_data.config.logging.options.json.infoBufferSize": [
      "0"
    ],
    "resource.raw.external_data.config.authorization.mode": [
      "Webhook"
    ],
    "resource.raw.external_data.config.volumeStatsAggPeriod": [
      "0s"
    ],
    "host.architecture": [
      "aarch64"
    ],
    "agent.id": [
      "ac066b7d-887d-4f0f-a562-5c6e2cc1ce07"
    ],
    "host.containerized": [
      false
    ],
    "resource.raw.external_data.config.runtimeRequestTimeout": [
      "0s"
    ],
    "resource.raw.external_data.config.httpCheckFrequency": [
      "0s"
    ],
    "resource.raw.stat.Group": [
      "705"
    ],
    "resource.raw.stat.StartTime": [
      "338857"
    ],
    "result.evidence.process_args.--bootstrap-kubeconfig": [
      "/etc/kubernetes/bootstrap-kubelet.conf"
    ],
    "resource.raw.external_data.config.evictionPressureTransitionPeriod": [
      "0s"
    ],
    "rule.tags": [
      "CIS",
      "Kubernetes",
      "CIS 4.2.10",
      "Kubelet"
    ],
    "host.ip": [
      "10.244.0.1",
      "10.244.0.1",
      "10.244.0.1",
      "172.18.0.2",
      "fc00:f853:ccd:e793::2",
      "fe80::42:acff:fe12:2",
      "172.19.0.2"
    ],
    "agent.type": [
      "cloudbeat"
    ],
    "resource.raw.external_data.config.nodeStatusUpdateFrequency": [
      "0s"
    ],
    "result.evidence.process_args.--node-labels": [
      ""
    ],
    "resource.raw.stat.TotalSize": [
      "2095192000"
    ],
    "process.uptime": [
      931
    ],
    "resource.raw.external_data.config.healthzBindAddress": [
      "127.0.0.1"
    ],
    "resource.raw.external_data.config.apiVersion": [
      "kubelet.config.k8s.io/v1beta1"
    ],
    "result.evidence.process_args.--container-runtime-endpoint": [
      "unix:///run/containerd/containerd.sock"
    ],
    "resource.raw.stat.SavedUID": [
      ""
    ],
    "result.evidence.process_args.--/usr/bin/kubelet": [
      ""
    ],
    "host.os.codename": [
      "bullseye"
    ],
    "resource.raw.external_data.config.failSwapOn": [
      false
    ],
    "process.pgid": [
      705
    ],
    "resource.raw.pid": [
      "705"
    ],
    "resource.id": [
      "584d3461-8fab-53ca-8ef6-04e26d0c2162"
    ],
    "rule.section": [
      "Kubelet"
    ],
    "result.evidence.process_args.--container-runtime": [
      "remote"
    ],
    "@timestamp": [
      "2022-09-05T07:39:03.401Z"
    ],
    "host.os.platform": [
      "debian"
    ],
    "result.evidence.process_args.--node-ip": [
      "172.18.0.2"
    ],
    "resource.raw.external_data.config.authentication.x509.clientCAFile": [
      "/etc/kubernetes/pki/ca.crt"
    ],
    "resource.raw.command": [
      "/usr/bin/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock --node-ip=172.18.0.2 --node-labels= --pod-infra-container-image=k8s.gcr.io/pause:3.7 --provider-id=kind://docker/kind-mono/kind-mono-control-plane --fail-swap-on=false --cgroup-root=/kubelet"
    ],
    "resource_id": [
      "584d3461-8fab-53ca-8ef6-04e26d0c2162"
    ],
    "agent.ephemeral_id": [
      "b7a60101-e450-4085-9514-fed2b48eeaa9"
    ],
    "result.evidence.process_args.--pod-infra-container-image": [
      "k8s.gcr.io/pause:3.7"
    ],
    "event.id": [
      "ba5d285f-7ce3-4aa3-9a2a-5f1ab30c7ace"
    ],
    "result.evidence.process_args.--kubeconfig": [
      "/etc/kubernetes/kubelet.conf"
    ],
    "rule.benchmark.version": [
      "v1.0.0"
    ],
    "result.evaluation": [
      "failed"
    ],
    "resource.raw.external_data.config.rotateCertificates": [
      true
    ],
    "resource.raw.stat.State": [
      "S"
    ],
    "result.evidence.process_args.--cgroup-root": [
      "/kubelet"
    ],
    "process.pid": [
      705
    ],
    "resource.raw.external_data.config.healthzPort": [
      10248
    ],
    "resource.raw.external_data.config.cgroupRoot": [
      "/kubelet"
    ],
    "resource.raw.stat.Threads": [
      "19"
    ],
    "resource.raw.stat.RealUID": [
      ""
    ],
    "host.os.name": [
      "Debian GNU/Linux"
    ],
    "resource.raw.external_data.config.shutdownGracePeriodCriticalPods": [
      "0s"
    ],
    "host.name": [
      "kind-mono-control-plane"
    ],
    "resource.raw.external_data.config.evictionHard.nodefs.inodesFree": [
      "0%"
    ],
    "resource.raw.stat.EffectiveGID": [
      ""
    ],
    "event.kind": [
      "state"
    ],
    "rule.name": [
      "Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual)"
    ],
    "rule.impact": [
      ""
    ],
    "resource.raw.stat.SystemTime": [
      "1863"
    ],
    "rule.default_value": [
      ""
    ],
    "rule.description": [
      "Setup TLS connection on the Kubelets.\n"
    ],
    "resource.type": [
      "process"
    ],
    "data_stream.type": [
      "logs"
    ],
    "rule.references": [
      ""
    ],
    "process.name": [
      "kubelet"
    ],
    "resource.raw.external_data.config.fileCheckFrequency": [
      "0s"
    ],
    "rule.audit": [
      "Run the following command on each node:\n```\nps -ef | grep kubelet\n```\nVerify that the `--tls-cert-file` and `--tls-private-key-file` arguments exist and they are set as\nappropriate.\nIf these arguments are not present, check that there is a Kubelet config specified by `--config`\nand that it contains appropriate settings for tlsCertFile and tlsPrivateKeyFile.\n"
    ],
    "ecs.version": [
      "8.0.0"
    ],
    "resource.raw.stat.Name": [
      "kubelet"
    ],
    "event.created": [
      "2022-09-05T07:39:03.401055006Z"
    ],
    "resource.raw.external_data.config.evictionHard.nodefs.available": [
      "0%"
    ],
    "agent.version": [
      "8.5.0"
    ],
    "resource.raw.external_data.config.logging.verbosity": [
      0
    ],
    "host.os.family": [
      "debian"
    ],
    "process.title": [
      "kubelet"
    ],
    "process.start": [
      "2022-09-05T07:23:31.195892255Z"
    ],
    "rule.benchmark.name": [
      "CIS Kubernetes V1.23"
    ],
    "result.evidence.process_args.--config": [
      "/var/lib/kubelet/config.yaml"
    ],
    "resource.raw.external_data.config.streamingConnectionIdleTimeout": [
      "0s"
    ],
    "resource.name": [
      "kubelet"
    ],
    "resource.raw.external_data.config.authentication.webhook.enabled": [
      true
    ],
    "resource.ecsFormat": [
      "process"
    ],
    "resource.raw.external_data.config.authentication.webhook.cacheTTL": [
      "0s"
    ],
    "event.sequence": [
      1662363520
    ],
    "cluster_id": [
      "851b737d-c31e-4c3b-9e76-86d3805b9d6f"
    ],
    "host.os.kernel": [
      "5.10.104-linuxkit"
    ],
    "resource.raw.external_data.config.shutdownGracePeriod": [
      "0s"
    ],
    "resource.raw.stat.Nice": [
      "0"
    ],
    "resource.raw.external_data.config.authorization.webhook.cacheUnauthorizedTTL": [
      "0s"
    ],
    "resource.raw.stat.Parent": [
      "1"
    ],
    "resource.raw.external_data.config.cgroupDriver": [
      "systemd"
    ],
    "process.args_count": [
      12
    ],
    "process.args": [
      "/usr/bin/kubelet",
      "--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf",
      "--kubeconfig=/etc/kubernetes/kubelet.conf",
      "--config=/var/lib/kubelet/config.yaml",
      "--container-runtime=remote",
      "--container-runtime-endpoint=unix:///run/containerd/containerd.sock",
      "--node-ip=172.18.0.2",
      "--node-labels=",
      "--pod-infra-container-image=k8s.gcr.io/pause:3.7",
      "--provider-id=kind://docker/kind-mono/kind-mono-control-plane",
      "--fail-swap-on=false",
      "--cgroup-root=/kubelet"
    ],
    "rule.remediation": [
      "If using a Kubelet config file, edit the file to set tlsCertFile to the location of the certificate\nfile to use to identify this Kubelet, and tlsPrivateKeyFile to the location of the\ncorresponding private key file.\nIf using command line arguments, edit the kubelet service file\n`/etc/systemd/system/kubelet.service.d/10-kubeadm.conf` on each worker node and set\nthe below parameters in KUBELET_CERTIFICATE_ARGS variable.\n--tls-cert-file=<path/to/tls-certificate-file> --tls-private-key-file=<path/to/tls-key-file>\nBased on your system, restart the kubelet service. For example:\n```\nsystemctl daemon-reload\nsystemctl restart kubelet.service\n```\n"
    ],
    "message": [
      "Rule \"Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual)\": failed"
    ],
    "resource.raw.stat.ResidentSize": [
      "78320000"
    ],
    "rule.version": [
      "1.0"
    ],
    "resource.raw.external_data.config.evictionHard.imagefs.available": [
      "0%"
    ],
    "resource.raw.external_data.config.clusterDNS": [
      "10.96.0.10"
    ],
    "result.evidence.process_args.--provider-id": [
      "kind://docker/kind-mono/kind-mono-control-plane"
    ],
    "rule.benchmark.id": [
      "cis_k8s"
    ],
    "resource.raw.stat.SavedGID": [
      ""
    ],
    "event.type": [
      "info"
    ],
    "process.command_line": [
      "/usr/bin/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock --node-ip=172.18.0.2 --node-labels= --pod-infra-container-image=k8s.gcr.io/pause:3.7 --provider-id=kind://docker/kind-mono/kind-mono-control-plane --fail-swap-on=false --cgroup-root=/kubelet"
    ],
    "resource.raw.external_data.config.cpuManagerReconcilePeriod": [
      "0s"
    ]
  }
}
```

# k8s_object 

## New

```json
{
  "_index": ".ds-logs-cloud_security_posture.findings-default-2022.09.01-000001",
  "_id": "5ZaKDIMB5gnkGmib4OEh",
  "_version": 1,
  "_score": 0,
  "_source": {
    "@timestamp": "2022-09-05T07:26:11.102Z",
    "type": "k8s_object",
    "result": {
      "evidence": {
        "serviceAccount": [],
        "serviceAccounts": []
      },
      "evaluation": "passed"
    },
    "resource": {
      "name": "daemon-set-controller",
      "raw": {
        "apiVersion": "v1",
        "kind": "ServiceAccount",
        "metadata": {
          "creationTimestamp": "2022-09-05T07:23:31Z",
          "name": "daemon-set-controller",
          "namespace": "kube-system",
          "resourceVersion": "263",
          "uid": "0f51286d-a204-4d69-b6c9-d07881c92350"
        }
      },
      "id": "0f51286d-a204-4d69-b6c9-d07881c92350",
      "type": "k8s_object",
      "sub_type": "ServiceAccount"
    },
    "resource_id": "0f51286d-a204-4d69-b6c9-d07881c92350",
    "rule": {
      "tags": [
        "CIS",
        "Kubernetes",
        "CIS 5.1.5",
        "RBAC and Service Accounts"
      ],
      "name": "Ensure that default service accounts are not actively used. (Manual)",
      "description": "The `default` service account should not be used to ensure that rights granted to applications can be more easily audited and reviewed.\n",
      "audit": "For each namespace in the cluster, review the rights assigned to the default service account and ensure that it has no roles or cluster roles bound to it apart from the defaults. Additionally ensure that the `automountServiceAccountToken: false` setting is in place for each default service account.\n",
      "version": "1.0",
      "profile_applicability": "* Level 1 - Master Node\n",
      "impact": "All workloads which require access to the Kubernetes API will require an explicit service account to be created.\n",
      "references": "1. [https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)\n",
      "remediation": "Create explicit service accounts wherever a Kubernetes workload requires\nspecific access\nto the Kubernetes API server.\nModify the configuration of each default service account to include this value\n```\nautomountServiceAccountToken: false\n```\n",
      "benchmark": {
        "name": "CIS Kubernetes V1.23",
        "version": "v1.0.0",
        "id": "cis_k8s"
      },
      "id": "2b399496-f79d-5533-8a86-4ea00b95e3bd",
      "rationale": "Kubernetes provides a `default` service account which is used by cluster workloads where no specific service account is assigned to the pod. Where access to the Kubernetes API from a pod is required, a specific service account should be created for that pod, and rights granted to that service account. The default service account should be configured such that it does not provide a service account token and does not have any explicit rights assignments.\n",
      "default_value": "By default the `default` service account allows for its service account token\nto be mounted\nin pods in its namespace.\n",
      "section": "RBAC and Service Accounts"
    },
    "message": "Rule \"Ensure that default service accounts are not actively used. (Manual)\": passed",
    "event": {
      "type": [
        "info"
      ],
      "sequence": 1662362770,
      "created": "2022-09-05T07:26:11.102Z",
      "id": "6b203e10-fe58-4eff-9158-cb9c3e20a3eb",
      "kind": "state",
      "category": [
        "configuration"
      ],
      "outcome": "success"
    },
    "ecs": {
      "version": "8.0.0"
    },
    "host": {
      "hostname": "kind-mono-control-plane",
      "architecture": "aarch64",
      "os": {
        "codename": "bullseye",
        "type": "linux",
        "platform": "debian",
        "version": "11 (bullseye)",
        "family": "debian",
        "name": "Debian GNU/Linux",
        "kernel": "5.10.104-linuxkit"
      },
      "containerized": false,
      "name": "kind-mono-control-plane",
      "ip": [
        "10.244.0.1",
        "10.244.0.1",
        "10.244.0.1",
        "172.18.0.2",
        "fc00:f853:ccd:e793::2",
        "fe80::42:acff:fe12:2",
        "172.19.0.2"
      ],
      "mac": [
        "02:42:ac:12:00:02",
        "02:42:ac:13:00:02",
        "1e:79:00:2d:a0:f8",
        "92:1f:b6:41:7d:43",
        "fe:76:90:6b:d2:f1"
      ]
    },
    "agent": {
      "version": "8.5.0",
      "ephemeral_id": "1b86fa91-10e9-4f9b-bc6a-a50c3cf37f8e",
      "id": "41467e74-babc-4580-95f8-60ae97770294",
      "name": "kind-mono-control-plane",
      "type": "cloudbeat"
    },
    "cluster_id": "851b737d-c31e-4c3b-9e76-86d3805b9d6f"
  },
  "fields": {
    "rule.id": [
      "2b399496-f79d-5533-8a86-4ea00b95e3bd"
    ],
    "result.evaluation": [
      "passed"
    ],
    "event.category": [
      "configuration"
    ],
    "host.hostname": [
      "kind-mono-control-plane"
    ],
    "type": [
      "k8s_object"
    ],
    "host.mac": [
      "02:42:ac:12:00:02",
      "02:42:ac:13:00:02",
      "1e:79:00:2d:a0:f8",
      "92:1f:b6:41:7d:43",
      "fe:76:90:6b:d2:f1"
    ],
    "resource.raw.metadata.namespace": [
      "kube-system"
    ],
    "rule.profile_applicability": [
      "* Level 1 - Master Node\n"
    ],
    "resource.sub_type": [
      "ServiceAccount"
    ],
    "resource.raw.kind": [
      "ServiceAccount"
    ],
    "host.os.version": [
      "11 (bullseye)"
    ],
    "host.os.name": [
      "Debian GNU/Linux"
    ],
    "agent.name": [
      "kind-mono-control-plane"
    ],
    "rule.rationale": [
      "Kubernetes provides a `default` service account which is used by cluster workloads where no specific service account is assigned to the pod. Where access to the Kubernetes API from a pod is required, a specific service account should be created for that pod, and rights granted to that service account. The default service account should be configured such that it does not provide a service account token and does not have any explicit rights assignments.\n"
    ],
    "host.name": [
      "kind-mono-control-plane"
    ],
    "event.kind": [
      "state"
    ],
    "event.outcome": [
      "success"
    ],
    "host.os.type": [
      "linux"
    ],
    "rule.name": [
      "Ensure that default service accounts are not actively used. (Manual)"
    ],
    "rule.impact": [
      "All workloads which require access to the Kubernetes API will require an explicit service account to be created.\n"
    ],
    "rule.default_value": [
      "By default the `default` service account allows for its service account token\nto be mounted\nin pods in its namespace.\n"
    ],
    "rule.description": [
      "The `default` service account should not be used to ensure that rights granted to applications can be more easily audited and reviewed.\n"
    ],
    "resource.type": [
      "k8s_object"
    ],
    "data_stream.type": [
      "logs"
    ],
    "rule.references": [
      "1. [https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)\n"
    ],
    "host.architecture": [
      "aarch64"
    ],
    "rule.audit": [
      "For each namespace in the cluster, review the rights assigned to the default service account and ensure that it has no roles or cluster roles bound to it apart from the defaults. Additionally ensure that the `automountServiceAccountToken: false` setting is in place for each default service account.\n"
    ],
    "agent.id": [
      "41467e74-babc-4580-95f8-60ae97770294"
    ],
    "host.containerized": [
      false
    ],
    "ecs.version": [
      "8.0.0"
    ],
    "event.created": [
      "2022-09-05T07:26:11.102Z"
    ],
    "agent.version": [
      "8.5.0"
    ],
    "host.os.family": [
      "debian"
    ],
    "rule.benchmark.name": [
      "CIS Kubernetes V1.23"
    ],
    "resource.raw.metadata.name": [
      "daemon-set-controller"
    ],
    "resource.raw.metadata.uid": [
      "0f51286d-a204-4d69-b6c9-d07881c92350"
    ],
    "resource.name": [
      "daemon-set-controller"
    ],
    "rule.tags": [
      "CIS",
      "Kubernetes",
      "CIS 5.1.5",
      "RBAC and Service Accounts"
    ],
    "event.sequence": [
      1662362750
    ],
    "host.ip": [
      "10.244.0.1",
      "10.244.0.1",
      "10.244.0.1",
      "172.18.0.2",
      "fc00:f853:ccd:e793::2",
      "fe80::42:acff:fe12:2",
      "172.19.0.2"
    ],
    "cluster_id": [
      "851b737d-c31e-4c3b-9e76-86d3805b9d6f"
    ],
    "agent.type": [
      "cloudbeat"
    ],
    "resource.raw.metadata.creationTimestamp": [
      "2022-09-05T07:23:31Z"
    ],
    "host.os.kernel": [
      "5.10.104-linuxkit"
    ],
    "resource.raw.apiVersion": [
      "v1"
    ],
    "resource.raw.metadata.resourceVersion": [
      "263"
    ],
    "host.os.codename": [
      "bullseye"
    ],
    "rule.remediation": [
      "Create explicit service accounts wherever a Kubernetes workload requires\nspecific access\nto the Kubernetes API server.\nModify the configuration of each default service account to include this value\n```\nautomountServiceAccountToken: false\n```\n"
    ],
    "message": [
      "Rule \"Ensure that default service accounts are not actively used. (Manual)\": passed"
    ],
    "rule.version": [
      "1.0"
    ],
    "resource.id": [
      "0f51286d-a204-4d69-b6c9-d07881c92350"
    ],
    "rule.section": [
      "RBAC and Service Accounts"
    ],
    "rule.benchmark.id": [
      "cis_k8s"
    ],
    "@timestamp": [
      "2022-09-05T07:26:11.102Z"
    ],
    "host.os.platform": [
      "debian"
    ],
    "event.type": [
      "info"
    ],
    "resource_id": [
      "0f51286d-a204-4d69-b6c9-d07881c92350"
    ],
    "agent.ephemeral_id": [
      "1b86fa91-10e9-4f9b-bc6a-a50c3cf37f8e"
    ],
    "event.id": [
      "6b203e10-fe58-4eff-9158-cb9c3e20a3eb"
    ],
    "rule.benchmark.version": [
      "v1.0.0"
    ]
  }
}

```

## Old

```json
{
  "_index": ".ds-logs-cloud_security_posture.findings-default-2022.09.01-000001",
  "_id": "I5eSDIMB5gnkGmibigEA",
  "_version": 1,
  "_score": 0,
  "_source": {
    "@timestamp": "2022-09-05T07:34:33.227Z",
    "event": {
      "type": [
        "info"
      ],
      "category": [
        "configuration"
      ],
      "created": "2022-09-05T07:34:33.226887258Z",
      "id": "d452ca92-cda0-4151-a494-60f60e9226bd",
      "kind": "state",
      "sequence": 1662363272,
      "outcome": "success"
    },
    "cluster_id": "851b737d-c31e-4c3b-9e76-86d3805b9d6f",
    "rule": {
      "name": "Ensure that default service accounts are not actively used. (Manual)",
      "rationale": "Kubernetes provides a `default` service account which is used by cluster workloads where no specific service account is assigned to the pod. Where access to the Kubernetes API from a pod is required, a specific service account should be created for that pod, and rights granted to that service account. The default service account should be configured such that it does not provide a service account token and does not have any explicit rights assignments.\n",
      "audit": "For each namespace in the cluster, review the rights assigned to the default service account and ensure that it has no roles or cluster roles bound to it apart from the defaults. Additionally ensure that the `automountServiceAccountToken: false` setting is in place for each default service account.\n",
      "remediation": "Create explicit service accounts wherever a Kubernetes workload requires\nspecific access\nto the Kubernetes API server.\nModify the configuration of each default service account to include this value\n```\nautomountServiceAccountToken: false\n```\n",
      "tags": [
        "CIS",
        "Kubernetes",
        "CIS 5.1.5",
        "RBAC and Service Accounts"
      ],
      "benchmark": {
        "id": "cis_k8s",
        "name": "CIS Kubernetes V1.23",
        "version": "v1.0.0"
      },
      "id": "2b399496-f79d-5533-8a86-4ea00b95e3bd",
      "profile_applicability": "* Level 1 - Master Node\n",
      "version": "1.0",
      "description": "The `default` service account should not be used to ensure that rights granted to applications can be more easily audited and reviewed.\n",
      "impact": "All workloads which require access to the Kubernetes API will require an explicit service account to be created.\n",
      "default_value": "By default the `default` service account allows for its service account token\nto be mounted\nin pods in its namespace.\n",
      "references": "1. [https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)\n",
      "section": "RBAC and Service Accounts"
    },
    "result": {
      "evaluation": "passed",
      "expected": null,
      "evidence": {
        "serviceAccount": [],
        "serviceAccounts": []
      }
    },
    "message": "Rule \"Ensure that default service accounts are not actively used. (Manual)\": passed",
    "resource": {
      "type": "k8s_object",
      "sub_type": "ServiceAccount",
      "name": "daemon-set-controller",
      "raw": {
        "metadata": {
          "resourceVersion": "263",
          "uid": "0f51286d-a204-4d69-b6c9-d07881c92350",
          "creationTimestamp": "2022-09-05T07:23:31Z",
          "name": "daemon-set-controller",
          "namespace": "kube-system"
        },
        "apiVersion": "v1",
        "kind": "ServiceAccount"
      },
      "id": "0f51286d-a204-4d69-b6c9-d07881c92350"
    },
    "resource_id": "0f51286d-a204-4d69-b6c9-d07881c92350",
    "agent": {
      "ephemeral_id": "b7a60101-e450-4085-9514-fed2b48eeaa9",
      "id": "ac066b7d-887d-4f0f-a562-5c6e2cc1ce07",
      "name": "kind-mono-control-plane",
      "type": "cloudbeat",
      "version": "8.5.0"
    },
    "ecs": {
      "version": "8.0.0"
    },
    "host": {
      "hostname": "kind-mono-control-plane",
      "architecture": "aarch64",
      "os": {
        "kernel": "5.10.104-linuxkit",
        "codename": "bullseye",
        "type": "linux",
        "platform": "debian",
        "version": "11 (bullseye)",
        "family": "debian",
        "name": "Debian GNU/Linux"
      },
      "containerized": false,
      "ip": [
        "10.244.0.1",
        "10.244.0.1",
        "10.244.0.1",
        "172.18.0.2",
        "fc00:f853:ccd:e793::2",
        "fe80::42:acff:fe12:2",
        "172.19.0.2"
      ],
      "mac": [
        "02:42:ac:12:00:02",
        "02:42:ac:13:00:02",
        "1e:79:00:2d:a0:f8",
        "92:1f:b6:41:7d:43",
        "fe:76:90:6b:d2:f1"
      ],
      "name": "kind-mono-control-plane"
    },
    "type": "k8s_object"
  },
  "fields": {
    "rule.id": [
      "2b399496-f79d-5533-8a86-4ea00b95e3bd"
    ],
    "result.evaluation": [
      "passed"
    ],
    "event.category": [
      "configuration"
    ],
    "host.hostname": [
      "kind-mono-control-plane"
    ],
    "type": [
      "k8s_object"
    ],
    "host.mac": [
      "02:42:ac:12:00:02",
      "02:42:ac:13:00:02",
      "1e:79:00:2d:a0:f8",
      "92:1f:b6:41:7d:43",
      "fe:76:90:6b:d2:f1"
    ],
    "resource.raw.metadata.namespace": [
      "kube-system"
    ],
    "rule.profile_applicability": [
      "* Level 1 - Master Node\n"
    ],
    "resource.sub_type": [
      "ServiceAccount"
    ],
    "resource.raw.kind": [
      "ServiceAccount"
    ],
    "host.os.version": [
      "11 (bullseye)"
    ],
    "host.os.name": [
      "Debian GNU/Linux"
    ],
    "agent.name": [
      "kind-mono-control-plane"
    ],
    "rule.rationale": [
      "Kubernetes provides a `default` service account which is used by cluster workloads where no specific service account is assigned to the pod. Where access to the Kubernetes API from a pod is required, a specific service account should be created for that pod, and rights granted to that service account. The default service account should be configured such that it does not provide a service account token and does not have any explicit rights assignments.\n"
    ],
    "host.name": [
      "kind-mono-control-plane"
    ],
    "event.kind": [
      "state"
    ],
    "event.outcome": [
      "success"
    ],
    "host.os.type": [
      "linux"
    ],
    "rule.name": [
      "Ensure that default service accounts are not actively used. (Manual)"
    ],
    "rule.impact": [
      "All workloads which require access to the Kubernetes API will require an explicit service account to be created.\n"
    ],
    "rule.default_value": [
      "By default the `default` service account allows for its service account token\nto be mounted\nin pods in its namespace.\n"
    ],
    "rule.description": [
      "The `default` service account should not be used to ensure that rights granted to applications can be more easily audited and reviewed.\n"
    ],
    "resource.type": [
      "k8s_object"
    ],
    "data_stream.type": [
      "logs"
    ],
    "rule.references": [
      "1. [https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)\n"
    ],
    "host.architecture": [
      "aarch64"
    ],
    "rule.audit": [
      "For each namespace in the cluster, review the rights assigned to the default service account and ensure that it has no roles or cluster roles bound to it apart from the defaults. Additionally ensure that the `automountServiceAccountToken: false` setting is in place for each default service account.\n"
    ],
    "agent.id": [
      "ac066b7d-887d-4f0f-a562-5c6e2cc1ce07"
    ],
    "host.containerized": [
      false
    ],
    "ecs.version": [
      "8.0.0"
    ],
    "event.created": [
      "2022-09-05T07:34:33.226887258Z"
    ],
    "agent.version": [
      "8.5.0"
    ],
    "host.os.family": [
      "debian"
    ],
    "rule.benchmark.name": [
      "CIS Kubernetes V1.23"
    ],
    "resource.raw.metadata.name": [
      "daemon-set-controller"
    ],
    "resource.raw.metadata.uid": [
      "0f51286d-a204-4d69-b6c9-d07881c92350"
    ],
    "resource.name": [
      "daemon-set-controller"
    ],
    "rule.tags": [
      "CIS",
      "Kubernetes",
      "CIS 5.1.5",
      "RBAC and Service Accounts"
    ],
    "event.sequence": [
      1662363260
    ],
    "host.ip": [
      "10.244.0.1",
      "10.244.0.1",
      "10.244.0.1",
      "172.18.0.2",
      "fc00:f853:ccd:e793::2",
      "fe80::42:acff:fe12:2",
      "172.19.0.2"
    ],
    "cluster_id": [
      "851b737d-c31e-4c3b-9e76-86d3805b9d6f"
    ],
    "agent.type": [
      "cloudbeat"
    ],
    "resource.raw.metadata.creationTimestamp": [
      "2022-09-05T07:23:31Z"
    ],
    "host.os.kernel": [
      "5.10.104-linuxkit"
    ],
    "resource.raw.apiVersion": [
      "v1"
    ],
    "resource.raw.metadata.resourceVersion": [
      "263"
    ],
    "host.os.codename": [
      "bullseye"
    ],
    "rule.remediation": [
      "Create explicit service accounts wherever a Kubernetes workload requires\nspecific access\nto the Kubernetes API server.\nModify the configuration of each default service account to include this value\n```\nautomountServiceAccountToken: false\n```\n"
    ],
    "message": [
      "Rule \"Ensure that default service accounts are not actively used. (Manual)\": passed"
    ],
    "rule.version": [
      "1.0"
    ],
    "resource.id": [
      "0f51286d-a204-4d69-b6c9-d07881c92350"
    ],
    "rule.section": [
      "RBAC and Service Accounts"
    ],
    "rule.benchmark.id": [
      "cis_k8s"
    ],
    "@timestamp": [
      "2022-09-05T07:34:33.227Z"
    ],
    "host.os.platform": [
      "debian"
    ],
    "event.type": [
      "info"
    ],
    "resource_id": [
      "0f51286d-a204-4d69-b6c9-d07881c92350"
    ],
    "agent.ephemeral_id": [
      "b7a60101-e450-4085-9514-fed2b48eeaa9"
    ],
    "event.id": [
      "d452ca92-cda0-4151-a494-60f60e9226bd"
    ],
    "rule.benchmark.version": [
      "v1.0.0"
    ]
  }
}
```