#!/bin/bash

# URL base
BASE_URL="http://localhost:8000/product"

# Função para fazer requisição
make_request() {
  ENDPOINT=$1
  curl -s "${BASE_URL}${ENDPOINT}" > /dev/null
}

# Endpoints a serem testados
endpoints=(
  "/"
  "/getProductsByName/test"
  "/getProductsByPage?page=1&pageSize=10"
  "/getProductsByPageAndCategory?page=1&pageSize=10&category=electronics"
  "/getProductsCount"
)

# Loop para fazer 10 requisições por segundo
while true; do
  for endpoint in "${endpoints[@]}"; do
    for i in {1..2}; do
      make_request "$endpoint" &
    done
  done
  sleep 1
done
