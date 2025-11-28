#!/bin/bash

# Monitor Kubernetes resources during load testing
# Run this script in a separate terminal while the load test is running

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Refresh interval in seconds
REFRESH_INTERVAL=${1:-5}

echo -e "${BLUE}╔═══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           Velure K8s Load Test Monitor                       ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${CYAN}Refresh interval: ${REFRESH_INTERVAL}s${NC}"
echo -e "${CYAN}Press Ctrl+C to stop monitoring${NC}"
echo ""
sleep 2

while true; do
  # Clear screen
  clear

  # Header
  echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
  echo -e "${BLUE}  Velure K8s Monitoring Dashboard - $(date '+%Y-%m-%d %H:%M:%S')${NC}"
  echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
  echo ""

  # Pod Status
  echo -e "${GREEN}📦 Pod Status:${NC}"
  kubectl get pods 2>/dev/null | grep -E "NAME|auth-service|product-service|publish-order|process-order|mongodb|postgres|rabbitmq" || echo "  No pods found"
  echo ""

  # Resource Usage
  echo -e "${GREEN}💻 Resource Usage:${NC}"
  if kubectl top pods &>/dev/null; then
    kubectl top pods 2>/dev/null | grep -E "NAME|auth-service|product-service|publish-order|process-order|mongodb|postgres|rabbitmq" || echo "  No metrics available"
  else
    echo -e "  ${YELLOW}⚠️  Metrics server not available${NC}"
    echo "  Install with: kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml"
  fi
  echo ""

  # HPA Status (if exists)
  HPA_COUNT=$(kubectl get hpa 2>/dev/null | wc -l)
  if [ "$HPA_COUNT" -gt 1 ]; then
    echo -e "${GREEN}⚖️  Horizontal Pod Autoscaler:${NC}"
    kubectl get hpa 2>/dev/null || echo "  No HPA configured"
    echo ""
  fi

  # Service Endpoints
  echo -e "${GREEN}🌐 Service Endpoints:${NC}"
  kubectl get svc 2>/dev/null | grep -E "NAME|auth-service|product-service|publish-order|process-order" || echo "  No services found"
  echo ""

  # Recent Events
  echo -e "${GREEN}📋 Recent Events (Last 5):${NC}"
  kubectl get events --sort-by='.lastTimestamp' 2>/dev/null | tail -6 | head -5 || echo "  No recent events"
  echo ""

  # Pod Restarts
  RESTART_COUNT=$(kubectl get pods 2>/dev/null | grep -E "auth-service|product-service|publish-order|process-order" | awk '{sum += $4} END {print sum}')
  if [ -n "$RESTART_COUNT" ] && [ "$RESTART_COUNT" -gt 0 ]; then
    echo -e "${RED}⚠️  WARNING: Total pod restarts: $RESTART_COUNT${NC}"
    kubectl get pods 2>/dev/null | grep -E "auth-service|product-service|publish-order|process-order" | awk '$4 > 0 {print "  - " $1 ": " $4 " restarts"}'
    echo ""
  fi

  # Failed Pods
  FAILED_PODS=$(kubectl get pods 2>/dev/null | grep -v "Running\|Completed\|NAME" | wc -l)
  if [ "$FAILED_PODS" -gt 0 ]; then
    echo -e "${RED}❌ Failed/Pending Pods:${NC}"
    kubectl get pods 2>/dev/null | grep -v "Running\|Completed\|NAME" || true
    echo ""
  fi

  # Database Connections (if we can access)
  POSTGRES_POD=$(kubectl get pods 2>/dev/null | grep postgres | grep Running | head -1 | awk '{print $1}')
  if [ -n "$POSTGRES_POD" ]; then
    echo -e "${GREEN}🗄️  PostgreSQL Connections:${NC}"
    CONN_COUNT=$(kubectl exec "$POSTGRES_POD" -- psql -U velure_user -d postgres -t -c "SELECT count(*) FROM pg_stat_activity WHERE datname IS NOT NULL;" 2>/dev/null | tr -d ' ' || echo "N/A")
    echo "  Active connections: $CONN_COUNT"
    echo ""
  fi

  # RabbitMQ Queue Depth (if we can access)
  RABBITMQ_POD=$(kubectl get pods 2>/dev/null | grep rabbitmq | grep Running | head -1 | awk '{print $1}')
  if [ -n "$RABBITMQ_POD" ]; then
    echo -e "${GREEN}🐰 RabbitMQ Status:${NC}"
    kubectl exec "$RABBITMQ_POD" -- rabbitmqctl list_queues 2>/dev/null | head -5 || echo "  Unable to fetch queue stats"
    echo ""
  fi

  # Footer
  echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
  echo -e "${CYAN}Next refresh in ${REFRESH_INTERVAL}s... (Ctrl+C to stop)${NC}"

  sleep "$REFRESH_INTERVAL"
done
