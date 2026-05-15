# Velure - Terraform Infrastructure

Infrastructure-as-Code for deploying Velure on AWS using EKS.

## 📋 Prerequisites

```bash
# AWS CLI v2
aws --version  # >= 2.0.0
aws configure  # Configure credentials

# Terraform
terraform --version  # >= 1.6.0

# kubectl
kubectl version --client  # >= 1.28.0

# Helm
helm version  # >= 3.0.0
```

## 💰 Cost Estimate

**WARNING**: This is a setup tuned for personal projects, but it still incurs costs.

| Resource | Specification | Monthly cost (us-east-1) |
|----------|---------------|--------------------------|
| EKS Cluster | 1 cluster | $72.00 |
| EC2 Nodes | 2x t3.small (on-demand) | ~$30.00 |
| NAT Gateway | 1x + data transfer | ~$32.00 + transfer |
| RDS Auth | db.t4g.micro (Free Tier) | $0.00 (750h/month) |
| RDS Orders | db.t4g.micro (Free Tier) | $0.00 (750h/month) |
| EBS Volumes | 2x 20GB gp3 (nodes) | ~$3.20 |
| VPC | 1 VPC + subnets | $0.00 |
| CloudWatch Logs | ~5GB/month | ~$2.50 |
| **TOTAL** | | **~$140-150/month** |

### ⚠️ Free Tier (first AWS year)
- RDS: 750h/month of db.t4g.micro (enough for one instance 24/7)
- EBS: 30GB of gp3 storage

### 💡 Cost Reduction Tips
1. **Stop the nodes when not in use**: `kubectl scale deployment --all --replicas=0`
2. **Tear the infra down on weekends**: `terraform destroy`
3. **Use Spot Instances**: switch node_instance_type to spot (~70% savings)
4. **Monitor costs**: AWS Cost Explorer + Budget Alerts

## 🚀 Deploy

### 1. Clone and Configure

```bash
cd terraform/

# Copy the variables example
cp terraform.tfvars.example terraform.tfvars

# Edit variables (especially passwords!)
vim terraform.tfvars
```

### 2. Initialize Terraform

```bash
terraform init
```

### 3. Validate Configuration

```bash
terraform validate
terraform fmt -recursive
```

### 4. Review the Plan

```bash
terraform plan -out=tfplan

# Review carefully:
# - Resources that will be created
# - Estimated costs
# - Security groups
```

### 5. Apply Infrastructure

```bash
terraform apply tfplan

# Wait ~15-20 minutes
# The EKS cluster takes a while to come up
```

### 6. Configure kubectl

```bash
# Get the command to configure kubectl
terraform output -raw kubeconfig_command | bash

# Test connectivity
kubectl get nodes
kubectl get pods -A
```

## 🔧 Post-Deploy

### 1. Install AWS Load Balancer Controller

```bash
# Create the ServiceAccount with IRSA
kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aws-load-balancer-controller
  namespace: kube-system
  annotations:
    eks.amazonaws.com/role-arn: $(terraform output -raw alb_controller_role_arn)
EOF

# Install via Helm
helm repo add eks https://aws.github.io/eks-charts
helm repo update

helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=$(terraform output -raw eks_cluster_name) \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller

# Verify
kubectl get deployment -n kube-system aws-load-balancer-controller
```

### 2. Install Redis (In-Cluster)

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami

helm install redis bitnami/redis \
  --set architecture=standalone \
  --set auth.password="$(openssl rand -base64 32)" \
  --set master.persistence.size=1Gi \
  --set master.resources.requests.memory=256Mi \
  --set master.resources.requests.cpu=100m \
  --set master.resources.limits.memory=512Mi \
  --set master.resources.limits.cpu=200m

# Get the password
export REDIS_PASSWORD=$(kubectl get secret redis -o jsonpath="{.data.redis-password}" | base64 -d)
echo "Redis Password: $REDIS_PASSWORD"
```

### 3. Install RabbitMQ (In-Cluster)

```bash
helm install rabbitmq bitnami/rabbitmq \
  --set auth.username=admin \
  --set auth.password="$(openssl rand -base64 32)" \
  --set persistence.size=2Gi \
  --set resources.requests.memory=256Mi \
  --set resources.requests.cpu=100m \
  --set resources.limits.memory=512Mi \
  --set resources.limits.cpu=200m

# Get the password
export RABBITMQ_PASSWORD=$(kubectl get secret rabbitmq -o jsonpath="{.data.rabbitmq-password}" | base64 -d)
echo "RabbitMQ Password: $RABBITMQ_PASSWORD"
```

### 4. Configure Application Secrets

```bash
# RDS Auth Service
kubectl create secret generic auth-db-secret \
  --from-literal=username=postgres \
  --from-literal=password="$(terraform output -raw rds_auth_password)" \
  --from-literal=host="$(terraform output -raw rds_auth_address)" \
  --from-literal=port=5432 \
  --from-literal=database=velure_auth

# RDS Orders Service
kubectl create secret generic orders-db-secret \
  --from-literal=username=postgres \
  --from-literal=password="$(terraform output -raw rds_orders_password)" \
  --from-literal=host="$(terraform output -raw rds_orders_address)" \
  --from-literal=port=5432 \
  --from-literal=database=velure_orders
```

## 📊 Monitoring

### CloudWatch Logs

```bash
# Check EKS logs
aws logs tail /aws/eks/velure-cluster/cluster --follow

# Check RDS logs
aws logs tail /aws/rds/instance/velure-auth-db/postgresql --follow
```

### Kubernetes

```bash
# Nodes
kubectl top nodes

# Pods
kubectl top pods -A

# Events
kubectl get events -A --sort-by='.lastTimestamp'
```

## 🛠️ Troubleshooting

### Nodes do not join the cluster

```bash
# Check security groups
aws ec2 describe-security-groups \
  --group-ids $(terraform output -raw eks_node_security_group_id)

# Check the node console output
aws ec2 get-console-output --instance-id <instance-id>
```

### RDS is unreachable

```bash
# Test connectivity from a pod
kubectl run -it --rm debug --image=postgres:16-alpine --restart=Never -- \
  psql -h $(terraform output -raw rds_auth_address) -U postgres -d velure_auth

# Check the security group
aws ec2 describe-security-groups \
  --group-ids $(terraform output -raw rds_security_group_id)
```

### ALB is not created

```bash
# Check the controller logs
kubectl logs -n kube-system deployment/aws-load-balancer-controller

# Check the IAM role
aws iam get-role --role-name $(terraform output -raw alb_controller_role_name)
```

## 🗑️ Destroy Infrastructure

```bash
# WARNING: This will delete EVERYTHING, including RDS data!

# 1. Delete LoadBalancers created by the controller
kubectl delete ingress --all -A
kubectl delete service --field-selector spec.type=LoadBalancer -A

# 2. Wait for the ALBs to be deleted (~2 minutes)
aws elbv2 describe-load-balancers --query 'LoadBalancers[].LoadBalancerName'

# 3. Destroy with Terraform
terraform destroy

# Confirm with "yes"
```

## 📁 Structure

```
terraform/
├── main.tf                    # Root module
├── variables.tf               # Input variables
├── outputs.tf                 # Outputs
├── versions.tf                # Provider versions
├── terraform.tfvars.example   # Example configuration
└── modules/
    ├── vpc/                   # VPC, subnets, NAT
    ├── security-groups/       # Security groups
    ├── eks/                   # EKS cluster + nodes
    └── rds/                   # PostgreSQL databases
```

## 🔐 Security Best Practices

- ✅ IMDSv2 enforced on nodes
- ✅ Least-privilege security groups
- ✅ RDS in private subnet
- ✅ Encryption at rest (EBS + RDS)
- ✅ CloudWatch logs enabled
- ✅ IRSA (IAM Roles for Service Accounts)
- ✅ Secrets via Kubernetes Secrets (consider External Secrets Operator)
- ✅ Network policies (to implement)

## 📚 Next Steps

1. **External Secrets Operator**: integrate with AWS Secrets Manager
2. **Network Policies**: isolate pod-to-pod traffic
3. **Pod Security Standards**: enforce PSS restricted
4. **Prometheus + Grafana**: advanced monitoring
5. **ArgoCD**: GitOps for deployments
6. **Cert-Manager**: automatic TLS
7. **Karpenter**: autoscaling more efficient than Cluster Autoscaler

## 🔗 References

- [EKS Best Practices](https://aws.github.io/aws-eks-best-practices/)
- [AWS Load Balancer Controller](https://kubernetes-sigs.github.io/aws-load-balancer-controller/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
