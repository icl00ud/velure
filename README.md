# Velure - Cloud-Native E-Commerce Platform

<div align="center">

![Velure Architecture](https://img.shields.io/badge/Architecture-Microservices-blue)
![Infrastructure](https://img.shields.io/badge/Infrastructure-AWS_EKS-orange)
![IaC](https://img.shields.io/badge/IaC-Terraform-purple)
![Orchestration](https://img.shields.io/badge/Orchestration-Kubernetes-326CE5)

**Plataforma de e-commerce construída como projeto de aprendizado para demonstrar práticas modernas de DevOps, Cloud-Native Architecture e Site Reliability Engineering (SRE)**

</div>

---

## 📚 Documentação

Nós movemos toda a nossa extensa documentação (Arquitetura, Setup Local, AWS Deploy, e detalhes de cada Microserviço) para um **Portal de Documentação dedicado**.

Para visualizar a documentação completa, você tem duas opções:

### Opção 1: Visualizar Localmente (Recomendado)
O site de documentação foi construído com Docusaurus e está dentro da pasta `docs-site/`.

```bash
cd docs-site
npm install
npm run start
```
*Isso abrirá a documentação interativa em `http://localhost:3000` no seu navegador.*

### Opção 2: Ler os Arquivos Markdown Diretamente
Se preferir, você pode navegar pelos arquivos `.md` diretamente aqui no GitHub, dentro da pasta [`docs-site/docs/`](./docs-site/docs/).
- [Visão Geral](./docs-site/docs/01-overview.md)
- [Quick Start](./docs-site/docs/02-quickstart.md)
- [Arquitetura Core](./docs-site/docs/03-core-architecture.md)
- [Documentação dos Microserviços](./docs-site/docs/microservices/)

---

## 🚀 Quick Start (Resumo)

Se você já conhece o projeto e quer apenas rodar rapidamente:

**Pré-requisitos:** Docker, Make.
*Aviso: É obrigatório mapear `127.0.0.1 velure.local` no seu `/etc/hosts`.*

```bash
# Subir aplicação COMPLETA (infra + services + monitoring) localmente
make local-up

# Acesse: https://velure.local
```

Para AWS, use `make cloud-up`.

---
## 📄 Licença

Este projeto é licenciado sob a MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes.
