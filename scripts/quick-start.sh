#!/usr/bin/env bash
set -euo pipefail

# ===========================================================================================
# Quick Start Script
# ===========================================================================================
# Script rápido para deploy completo em uma única execução
# ===========================================================================================

cat << "EOF"
╦  ╦┌─┐┬  ┬ ┬┬─┐┌─┐
╚╗╔╝├┤ │  │ │├┬┘├┤ 
 ╚╝ └─┘┴─┘└─┘┴└─└─┘
Quick Deploy Script
EOF

echo ""
echo "Este script irá:"
echo "  1. Criar secrets no AWS Secrets Manager"
echo "  2. Provisionar infraestrutura AWS (EKS, RDS, Amazon MQ)"
echo "  3. Configurar Kubernetes (controllers + operators)"
echo "  4. Deploy datastores (MongoDB, Redis)"
echo "  5. Deploy microserviços Velure"
echo "  6. Instalar monitoramento (Prometheus + Grafana)"
echo ""
echo "Tempo estimado: 20-30 minutos"
echo ""

read -p "Continuar? (y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelado"
    exit 0
fi

# Executar bootstrap completo
./scripts/bootstrap.sh

echo ""
echo "✅ Deploy completo!"
echo ""
echo "📊 Próximos passos:"
echo ""
echo "  # Ver status dos pods"
echo "  kubectl get pods -A"
echo ""
echo "  # Obter URL da aplicação"
echo "  kubectl get ingress -n velure"
echo ""
echo "  # Acessar Grafana"
echo "  make eks-grafana"
echo ""
echo "  # Ver logs de um serviço"
echo "  kubectl logs -n velure -l app=velure-auth -f"
echo ""
