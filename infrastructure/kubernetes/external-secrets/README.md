# External Secrets for Velure

This directory contains ExternalSecret resources that sync secrets from AWS Secrets Manager to Kubernetes.

## Prerequisites

1. **Install External Secrets Operator**
   ```bash
   helm repo add external-secrets https://charts.external-secrets.io
   helm install external-secrets external-secrets/external-secrets \
     -n external-secrets --create-namespace
   ```

2. **Create IAM Role for Service Account (IRSA)**
   The Terraform module already creates the necessary IAM roles. You need to create the service account:
   ```bash
   kubectl create serviceaccount external-secrets-sa -n external-secrets
   kubectl annotate serviceaccount external-secrets-sa -n external-secrets \
     eks.amazonaws.com/role-arn=arn:aws:iam::<ACCOUNT_ID>:role/velure-external-secrets-role
   ```

## Deployment

1. **Apply ClusterSecretStore**
   ```bash
   kubectl apply -f cluster-secret-store.yaml
   ```

2. **Create namespaces (if not exists)**
   ```bash
   kubectl create ns authentication
   kubectl create ns order
   kubectl create ns frontend
   ```

3. **Apply ExternalSecrets**
   ```bash
   kubectl apply -f auth-service-secrets.yaml
   kubectl apply -f product-service-secrets.yaml
   kubectl apply -f publish-order-secrets.yaml
   kubectl apply -f process-order-secrets.yaml
   ```

4. **Verify secrets are synced**
   ```bash
   kubectl get externalsecrets -A
   kubectl get secrets -n authentication
   kubectl get secrets -n order
   ```

## Secret Structure in AWS Secrets Manager

The following secrets must exist in AWS Secrets Manager (created by Terraform):

- `velure/production/jwt` - JWT signing secrets
- `velure/production/rds-auth` - Auth database credentials
- `velure/production/rds-orders` - Orders database credentials
- `velure/production/rabbitmq` - RabbitMQ credentials
- `velure/production/mongodb` - MongoDB Atlas connection string
- `velure/production/redis` - Redis connection details

## Troubleshooting

Check ExternalSecret status:
```bash
kubectl describe externalsecret <name> -n <namespace>
```

Check operator logs:
```bash
kubectl logs -n external-secrets -l app.kubernetes.io/name=external-secrets
```
