# Documentation Site Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a professional static documentation site using Docusaurus to explain the Velure e-commerce microservices platform.

**Architecture:** A standard Docusaurus v3 installation inside the `docs-site/` directory, containing Markdown and MDX files. It will use the classic theme with dark mode and a side navigation bar.

**Tech Stack:** Docusaurus, React, Markdown, Mermaid.

### Task 1: Initialize Docusaurus Project

**Step 1: Run Docusaurus init command**
Run: `npx create-docusaurus@latest docs-site classic --typescript`
Expected: A new `docs-site` folder is created with the Docusaurus skeleton.

**Step 2: Verify installation**
Run: `cd docs-site && npm run build`
Expected: Build passes without errors.

**Step 3: Commit**
```bash
git add docs-site
git commit -m "docs: initialize docusaurus project"
```

### Task 2: Configure Docusaurus Settings

**Files:**
- Modify: `docs-site/docusaurus.config.ts`

**Step 1: Update configuration**
Update title to "Velure Docs", tagline to "E-commerce Microservices Platform", and configure GitHub links to the repository. Enable Mermaid support if needed.

**Step 2: Verify build**
Run: `cd docs-site && npm run build`
Expected: Build passes.

**Step 3: Commit**
```bash
git add docs-site/docusaurus.config.ts
git commit -m "docs: configure docusaurus site metadata"
```

### Task 3: Create Overview and Quickstart Pages

**Files:**
- Create: `docs-site/docs/01-overview.md`
- Create: `docs-site/docs/02-quickstart.md`
- Delete: Default docusaurus docs in `docs-site/docs/`

**Step 1: Write Overview**
Write the high-level architecture, tech stack, and reference the AWS diagram. Copy the AWS diagram from `diagrams/` to `docs-site/static/img/`.

**Step 2: Write Quickstart**
Write the local setup instructions (`make local-up`), prerequisites, and the mandatory `/etc/hosts` warning.

**Step 3: Verify docs**
Run: `cd docs-site && npm run build`
Expected: Build passes.

**Step 4: Commit**
```bash
git add docs-site/docs/ docs-site/static/
git commit -m "docs: add overview and quickstart guides"
```

### Task 4: Create Core Architecture & Event Flow Page

**Files:**
- Create: `docs-site/docs/03-core-architecture.md`

**Step 1: Write Event Flow**
Write the Mermaid sequence diagram explaining the order lifecycle (Frontend -> publish-order -> RabbitMQ -> process-order -> publish-order -> SSE -> Frontend).

**Step 2: Verify build**
Run: `cd docs-site && npm run build`
Expected: Build passes.

**Step 3: Commit**
```bash
git add docs-site/docs/03-core-architecture.md
git commit -m "docs: add core architecture and event flow diagram"
```

### Task 5: Create Microservices Reference

**Files:**
- Create: `docs-site/docs/microservices/auth-service.md`
- Create: `docs-site/docs/microservices/product-service.md`
- Create: `docs-site/docs/microservices/publish-order-service.md`
- Create: `docs-site/docs/microservices/process-order-service.md`
- Create: `docs-site/docs/microservices/ui-service.md`

**Step 1: Write microservices docs**
Document each service based on the design, focusing on tech stack, DB, endpoints, and environment variables.

**Step 2: Verify build**
Run: `cd docs-site && npm run build`
Expected: Build passes.

**Step 3: Commit**
```bash
git add docs-site/docs/microservices/
git commit -m "docs: add microservices reference documentation"
```

### Task 6: Update Root README.md

**Files:**
- Modify: `README.md` (at project root)

**Step 1: Update root README**
Replace the huge README with a concise introduction that points users to the new `docs-site/` for full documentation. Add instructions on how to run the documentation site locally (`cd docs-site && npm run start`).

**Step 2: Commit**
```bash
git add README.md
git commit -m "docs: update root readme to point to docusaurus site"
```
