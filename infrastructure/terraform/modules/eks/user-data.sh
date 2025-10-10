#!/bin/bash
set -o xtrace

# Bootstrap do EKS node
/etc/eks/bootstrap.sh ${cluster_name} ${bootstrap_arguments}
