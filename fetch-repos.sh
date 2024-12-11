#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

echo "Atualizando subtrees do repositório..."

# Lista de subtrees para atualizar
subtrees=(
    "velure-auth-service"
    "velure-ui-service"
    "velure-product-service"
)

# Caminho remoto e branch padrão para cada subtree
declare -A subtree_remotes=(
    ["velure-auth-service"]="https://github.com/icl00ud/velure-auth-service.git"
    ["velure-ui-service"]="https://github.com/icl00ud/velure-ui-service.git"
    ["velure-product-service"]="https://github.com/icl00ud/velure-product-service.git"
)

branch="master"

stash_saved=false

# Função para restaurar o stash no caso de interrupção
trap 'if [[ "${stash_saved}" = true ]]; then git stash pop; fi' EXIT

# Verifica alterações na árvore de trabalho
if ! git diff --quiet || ! git diff --cached --quiet; then
    echo "Há modificações pendentes na árvore de trabalho. Salvando no stash..."
    if git stash push -m "Stash automático antes de atualizar subtrees"; then
        stash_saved=true
        echo "Modificações stashed com sucesso."
    else
        echo "Erro ao salvar modificações no stash."
        exit 1
    fi
else
    echo "Nenhuma modificação pendente encontrada."
fi

# Atualiza cada subtree
for subtree in "${subtrees[@]}"; do
    remote="${subtree_remotes[${subtree}]}"

    if [[ -z "${remote}" ]]; then
        echo "Nenhum repositório remoto configurado para ${subtree}. Pulando..."
        continue
    fi

    echo "Atualizando subtree '${subtree}' a partir de '${remote}' na branch '${branch}'..."

    if git subtree pull --prefix="${subtree}" "${remote}" "${branch}" --squash; then
        echo "Subtree '${subtree}' atualizada com sucesso!"
    else
        echo "Erro ao atualizar '${subtree}'. Verifique manualmente."
    fi
done

# Restaura alterações do stash, se necessário
if [[ "${stash_saved}" = true ]]; then
    echo "Restaurando modificações do stash..."
    if git stash pop; then
        echo "Modificações restauradas com sucesso."
    else
        echo "Erro ao restaurar modificações do stash."
        exit 1
    fi
fi

echo "Atualização de subtrees concluída!"
