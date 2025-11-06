#!/bin/bash
set -o xtrace

# Bootstrap EKS node
/etc/eks/bootstrap.sh ${cluster_name} \
  --b64-cluster-ca ${cluster_ca} \
  --apiserver-endpoint ${cluster_endpoint} \
  --dns-cluster-ip ${cluster_service_cidr} \
  --use-max-pods false

# Enable IMDSv2
yum install -y amazon-ssm-agent
systemctl enable amazon-ssm-agent
systemctl start amazon-ssm-agent
