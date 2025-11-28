#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔═══════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║        PgBouncer Installation for Velure             ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════╝${NC}"
echo ""

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl not found. Please install kubectl first.${NC}"
    exit 1
fi

# Check if connected to cluster
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}❌ Not connected to Kubernetes cluster. Please configure kubectl first.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ kubectl connected to cluster${NC}"
kubectl cluster-info | head -1
echo ""

# Prompt for RDS credentials
echo -e "${YELLOW}📝 Enter RDS PostgreSQL credentials:${NC}"
echo ""

read -p "RDS Host (e.g., velure-production-auth.cw9gu66melkv.us-east-1.rds.amazonaws.com): " DB_HOST
read -p "Database User [postgres]: " DB_USER
DB_USER=${DB_USER:-postgres}
read -sp "Database Password: " DB_PASSWORD
echo ""
echo ""

if [ -z "$DB_HOST" ] || [ -z "$DB_PASSWORD" ]; then
    echo -e "${RED}❌ RDS Host and Password are required${NC}"
    exit 1
fi

echo -e "${BLUE}🔧 Installing PgBouncer...${NC}"
echo ""

# Create namespace
echo -e "${BLUE}1️⃣  Creating namespace velure-db...${NC}"
kubectl apply -f "${SCRIPT_DIR}/namespace.yaml"
echo ""

# Create secret
echo -e "${BLUE}2️⃣  Creating secret with RDS credentials...${NC}"
kubectl create secret generic pgbouncer-secret \
  --from-literal=db-host="${DB_HOST}" \
  --from-literal=db-user="${DB_USER}" \
  --from-literal=db-password="${DB_PASSWORD}" \
  --namespace velure-db \
  --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✅ Secret created${NC}"
echo ""

# Apply ConfigMap
echo -e "${BLUE}3️⃣  Applying PgBouncer configuration...${NC}"
kubectl apply -f "${SCRIPT_DIR}/configmap.yaml"
echo -e "${GREEN}✅ ConfigMap applied${NC}"
echo ""

# Apply Deployment
echo -e "${BLUE}4️⃣  Deploying PgBouncer pods...${NC}"
kubectl apply -f "${SCRIPT_DIR}/deployment.yaml"
echo -e "${GREEN}✅ Deployment created${NC}"
echo ""

# Apply Service
echo -e "${BLUE}5️⃣  Creating PgBouncer service...${NC}"
kubectl apply -f "${SCRIPT_DIR}/service.yaml" 2>/dev/null || true
echo -e "${GREEN}✅ Service created${NC}"
echo ""

# Wait for pods to be ready
echo -e "${BLUE}6️⃣  Waiting for PgBouncer pods to be ready...${NC}"
kubectl wait --for=condition=ready pod \
  -l app=pgbouncer \
  -n velure-db \
  --timeout=120s

echo -e "${GREEN}✅ PgBouncer pods are ready${NC}"
echo ""

# Show pod status
echo -e "${BLUE}📊 PgBouncer Status:${NC}"
kubectl get pods -n velure-db -l app=pgbouncer
echo ""

# Show service
echo -e "${BLUE}🌐 PgBouncer Service:${NC}"
kubectl get svc -n velure-db -l app=pgbouncer
echo ""

# Connection string
PGBOUNCER_HOST="pgbouncer.velure-db.svc.cluster.local"
echo -e "${GREEN}╔═══════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║           ✅ PgBouncer Installed Successfully!        ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}📝 Next Steps:${NC}"
echo ""
echo -e "1. Update your application deployments to use PgBouncer:"
echo -e "   ${BLUE}POSTGRES_HOST=${PGBOUNCER_HOST}${NC}"
echo ""
echo -e "2. Redeploy your services:"
echo -e "   ${BLUE}helm upgrade velure-auth-service ./infrastructure/kubernetes/charts/velure-auth-service${NC}"
echo -e "   ${BLUE}helm upgrade velure-publish-order-service ./infrastructure/kubernetes/charts/velure-publish-order-service${NC}"
echo ""
echo -e "3. Test connection:"
echo -e "   ${BLUE}kubectl run -it --rm debug --image=postgres:15 --restart=Never -- \\${NC}"
echo -e "   ${BLUE}  psql -h ${PGBOUNCER_HOST} -U ${DB_USER} -d velure_auth${NC}"
echo ""
echo -e "4. Monitor PgBouncer:"
echo -e "   ${BLUE}kubectl logs -f deployment/pgbouncer -n velure-db${NC}"
echo ""
echo -e "${YELLOW}📚 Documentation:${NC} ${SCRIPT_DIR}/README.md"
echo ""
