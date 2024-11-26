#!/bin/bash

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

# Verifica alterações na árvore de trabalho
if ! git diff --quiet || ! git diff --cached --quiet; then
    echo "Há modificações pendentes na árvore de trabalho. Salvando no stash..."
    git stash push -m "Stash automático antes de atualizar subtrees"
    stash_saved=true
else
    stash_saved=false
fi

# Atualiza cada subtree
for subtree in "${subtrees[@]}"; do
    remote="${subtree_remotes[$subtree]}"
    
    if [ -z "$remote" ]; then
        echo "Nenhum repositório remoto configurado para $subtree. Pule."
        continue
    fi
    
    echo "Atualizando subtree $subtree..."
    git subtree pull --prefix="$subtree" "$remote" "$branch" --squash
    if [ $? -eq 0 ]; then
        echo "Subtree $subtree atualizada com sucesso!"
    else
        echo "Erro ao atualizar $subtree. Verifique manualmente."
    fi
done

# Restaura alterações do stash, se necessário
if [ "$stash_saved" = true ]; then
    echo "Restaurando modificações do stash..."
    git stash pop
fi

echo "Atualização de subtrees concluída!"
