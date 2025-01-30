<h1>Velure Product Service</h1>

<p align="center">
  <img src="https://img.shields.io/static/v1?label=nestjs&message=framework&color=blue&style=for-the-badge&logo=nestjs"/>
  <img src="https://img.shields.io/static/v1?label=Redis&message=caching&color=blue&style=for-the-badge&logo=redis"/>
  <img src="http://img.shields.io/static/v1?label=License&message=MIT&color=green&style=for-the-badge"/>
  <img src="http://img.shields.io/static/v1?label=Status&message=IN%20PROGRESS&color=yellow&style=for-the-badge"/>
</p>

> Status do Projeto: :warning: Em desenvolvimento

### Tópicos

:small_blue_diamond: [Descrição do projeto](#descrição-do-projeto)

:small_blue_diamond: [Funcionalidades](#funcionalidades)

:small_blue_diamond: [Pré-requisitos](#pré-requisitos)

:small_blue_diamond: [Como rodar a aplicação](#como-rodar-a-aplicação-arrow_forward)

:small_blue_diamond: [Como rodar os testes](#como-rodar-os-testes)

## Descrição do projeto

<p align="justify">
  Este projeto é um serviço de gerenciamento de produtos, construído utilizando o framework NestJS. O serviço fornece endpoints para criar, ler, atualizar e excluir produtos, além de cachear dados com Redis e fornecer métricas de monitoramento via Prometheus.
</p>

## Funcionalidades

:heavy_check_mark: Cadastrar novos produtos

:heavy_check_mark: Listar todos os produtos

:heavy_check_mark: Buscar produtos por nome

:heavy_check_mark: Excluir produtos por ID ou nome

## Pré-requisitos

:warning: [Node](https://nodejs.org/en/download/)

:warning: [Redis](https://redis.io/download)

## Como rodar a aplicação :arrow_forward:

No terminal, clone o projeto: 

```
git clone https://github.com/icl00ud/velure-product-service
```

Instale as dependências:

```
npm install
```

Execute a aplicação:

```
npm run start
```
