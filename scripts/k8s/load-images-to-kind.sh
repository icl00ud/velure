#!/bin/bash

# Velure - Load Docker Images to kind Cluster
# Este script carrega as imagens Docker locais no cluster kind

set -e

BOLD="\033[1m"
GREEN="\033[0;32m"
BLUE="\033[0;34m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
NC="\033[0m"

CLUSTER_NAME="${1:-velure}"

echo ""
echo -e "${BOLD}${BLUE}🐳 Carregando imagens Docker no kind cluster '${CLUSTER_NAME}'${NC}"
echo ""

# Verificar se kind está instalado
if ! command -v kind &> /dev/null; then
    echo -e "${RED}❌ kind não está instalado${NC}"
    exit 1
fi

# Verificar se cluster existe
if ! kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo -e "${RED}❌ Cluster '${CLUSTER_NAME}' não existe${NC}"
    echo ""
    echo -e "${YELLOW}Clusters disponíveis:${NC}"
    kind get clusters
    exit 1
fi

IMAGES=(
    "velure-auth-service:latest"
    "velure-product-service:latest"
    "velure-publish-order-service:latest"
    "velure-process-order-service:latest"
    "velure-ui-service:latest"
)

echo -e "${BLUE}Verificando imagens...${NC}"
echo ""

MISSING_IMAGES=()
for img in "${IMAGES[@]}"; do
    if docker image inspect "$img" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $img"
    else
        echo -e "${RED}✗${NC} $img ${YELLOW}(não encontrada)${NC}"
        MISSING_IMAGES+=("$img")
    fi
done

echo ""

if [ ${#MISSING_IMAGES[@]} -gt 0 ]; then
    echo -e "${RED}❌ ${#MISSING_IMAGES[@]} imagem(ns) não encontrada(s)${NC}"
    echo ""
    echo -e "${YELLOW}Execute primeiro:${NC}"
    echo "  make docker-build"
    exit 1
fi

echo -e "${BLUE}Carregando imagens no cluster '${CLUSTER_NAME}'...${NC}"
echo ""

for img in "${IMAGES[@]}"; do
    echo -e "  → ${BOLD}$img${NC}"
    kind load docker-image "$img" --name "$CLUSTER_NAME"
done

echo ""
echo -e "${GREEN}✅ Todas as imagens foram carregadas com sucesso!${NC}"
echo ""
echo -e "${YELLOW}Próximo passo:${NC}"
echo "  Fazer deploy dos serviços com: make kind-deploy"
echo ""
