#!/bin/bash

# ========================================================================
# Script: monitor-scaling.sh
# Description: Monitor HPA and pod scaling in real-time during load tests
# Usage: ./monitor-scaling.sh [namespace]
# ========================================================================

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

NAMESPACE="${1:-default}"

clear

echo -e "${BOLD}${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${BLUE}║        Velure - HPA & Pod Scaling Monitor                 ║${NC}"
echo -e "${BOLD}${BLUE}║        Namespace: ${CYAN}${NAMESPACE}${BLUE}                                      ║${NC}"
echo -e "${BOLD}${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl not found${NC}"
    exit 1
fi

if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}Error: Not connected to Kubernetes cluster${NC}"
    exit 1
fi

echo -e "${GREEN}Connected to:${NC} $(kubectl config current-context)"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop monitoring${NC}"
echo ""

# Check if HPA exists
if ! kubectl get hpa -n $NAMESPACE &> /dev/null; then
    echo -e "${YELLOW}Warning: No HPA resources found in namespace ${NAMESPACE}${NC}"
    echo ""
fi

# Display functions
display_hpa() {
    echo -e "${BOLD}${CYAN}━━━ HorizontalPodAutoscalers ━━━${NC}"
    kubectl get hpa -n $NAMESPACE 2>/dev/null | tail -n +2 | while read -r line; do
        name=$(echo "$line" | awk '{print $1}')
        current=$(echo "$line" | awk '{print $5}')
        min=$(echo "$line" | awk '{print $6}')
        max=$(echo "$line" | awk '{print $7}')

        # Color based on scaling status
        if [ "$current" -gt "$min" ]; then
            echo -e "  ${GREEN}▲${NC} $line"
        else
            echo -e "  ${BLUE}●${NC} $line"
        fi
    done
    echo ""
}

display_pods() {
    echo -e "${BOLD}${CYAN}━━━ Pods Status ━━━${NC}"

    # Group by service
    for service in velure-auth velure-product velure-publish-order velure-process-order velure-ui; do
        pods=$(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=$service --no-headers 2>/dev/null | wc -l | tr -d ' ')

        if [ "$pods" -gt 0 ]; then
            ready=$(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=$service --no-headers 2>/dev/null | grep "Running" | grep "1/1" | wc -l | tr -d ' ')
            echo -e "  ${BLUE}${service}:${NC} ${GREEN}${ready}${NC}/${pods} ready"

            kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=$service --no-headers 2>/dev/null | while read -r line; do
                status=$(echo "$line" | awk '{print $3}')
                if [ "$status" == "Running" ]; then
                    echo -e "    ${GREEN}●${NC} $(echo "$line" | awk '{print $1}') - $(echo "$line" | awk '{print $3}')"
                else
                    echo -e "    ${YELLOW}◉${NC} $(echo "$line" | awk '{print $1}') - $(echo "$line" | awk '{print $3}')"
                fi
            done
        fi
    done
    echo ""
}

display_metrics() {
    echo -e "${BOLD}${CYAN}━━━ Resource Metrics (via metrics-server) ━━━${NC}"

    # Check if metrics-server is available
    if ! kubectl top nodes &> /dev/null; then
        echo -e "  ${YELLOW}⚠ metrics-server not available${NC}"
        echo ""
        return
    fi

    # Show CPU/Memory for pods
    for service in velure-auth velure-product velure-publish-order velure-process-order velure-ui; do
        metrics=$(kubectl top pods -n $NAMESPACE -l app.kubernetes.io/name=$service --no-headers 2>/dev/null)

        if [ -n "$metrics" ]; then
            echo -e "  ${BLUE}${service}:${NC}"
            echo "$metrics" | while read -r line; do
                cpu=$(echo "$line" | awk '{print $2}')
                mem=$(echo "$line" | awk '{print $3}')
                echo -e "    CPU: ${GREEN}${cpu}${NC}  Memory: ${GREEN}${mem}${NC}"
            done
        fi
    done
    echo ""
}

display_events() {
    echo -e "${BOLD}${CYAN}━━━ Recent Scaling Events (last 5 minutes) ━━━${NC}"

    events=$(kubectl get events -n $NAMESPACE --sort-by='.lastTimestamp' 2>/dev/null | grep -i "scaled\|horizontal" | tail -5)

    if [ -n "$events" ]; then
        echo "$events" | while read -r line; do
            echo -e "  ${YELLOW}➤${NC} $(echo "$line" | awk '{print $1, $NF}')"
        done
    else
        echo -e "  ${BLUE}No recent scaling events${NC}"
    fi
    echo ""
}

# Main monitoring loop
while true; do
    clear

    echo -e "${BOLD}${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}${BLUE}║        Velure - HPA & Pod Scaling Monitor                 ║${NC}"
    echo -e "${BOLD}${BLUE}║        $(date '+%Y-%m-%d %H:%M:%S')                              ║${NC}"
    echo -e "${BOLD}${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""

    display_hpa
    display_pods
    display_metrics
    display_events

    echo -e "${YELLOW}Refreshing every 5 seconds... (Ctrl+C to stop)${NC}"

    sleep 5
done
