# Velure - Cloud-Native E-Commerce Platform

<div align="center">

![Velure Architecture](https://img.shields.io/badge/Architecture-Microservices-blue)
![Infrastructure](https://img.shields.io/badge/Infrastructure-AWS_EKS-orange)
![IaC](https://img.shields.io/badge/IaC-Terraform-purple)
![Orchestration](https://img.shields.io/badge/Orchestration-Kubernetes-326CE5)

**An e-commerce platform built as a learning project to demonstrate modern DevOps, Cloud-Native Architecture, and Site Reliability Engineering (SRE) practices**

</div>

---

## 📚 Documentation

We have moved our extensive documentation (Architecture, Local Setup, AWS Deploy, and Microservices details) to a **dedicated Documentation Portal**.

To view the complete documentation, you have two options:

### Option 1: View Locally via Docker (Recommended)
The documentation site was built with Docusaurus and packaged into a container for easy access.

```bash
make docs-up
```
*This will build the container and open the documentation on port 3000 in the background. You can access it at `http://localhost:3000` in your browser.*

When you're done, simply run:
```bash
make docs-down
```

### Option 2: Read the Markdown Files Directly
If you prefer, you can navigate through the `.md` files directly here on GitHub, inside the [`docs-site/docs/`](./docs-site/docs/) directory.
- [Overview](./docs-site/docs/01-overview.md)
- [Quick Start](./docs-site/docs/02-quickstart.md)
- [Core Architecture](./docs-site/docs/03-core-architecture.md)
- [Microservices Documentation](./docs-site/docs/microservices/)

---

## 🚀 Quick Start (Summary)

If you are already familiar with the project and just want to run it quickly:

**Prerequisites:** Docker, Make.
*Warning: It is mandatory to map `127.0.0.1 velure.local` in your `/etc/hosts`.*

```bash
# Bring up the ENTIRE application (infra + services + monitoring) locally
make local-up

# Access: https://velure.local
```

For AWS, use `make cloud-up`.

---
## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
