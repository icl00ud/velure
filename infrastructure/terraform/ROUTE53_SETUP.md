# Route53 Setup Guide - Registro.br Domain

Este guia explica como configurar seu domínio do Registro.br para usar o Route53 da AWS.

## 📋 Pré-requisitos

- Domínio registrado no Registro.br (`velure.app.br`)
- Terraform aplicado com `create_dns_record = false` (primeiro deploy)
- ALB criado pelo Helm deployment

## 🚀 Passo a Passo

### **1. Primeiro Deploy do Terraform (Criar Hosted Zone)**

```bash
cd /Users/icl00ud/repos/velure/infrastructure/terraform

# Verificar que create_dns_record = false no main.tf
grep "create_dns_record" main.tf

# Aplicar Terraform
terraform plan
terraform apply
```

### **2. Obter os Nameservers da AWS**

Após o `terraform apply`, copie os nameservers:

```bash
terraform output route53_name_servers
```

Você receberá algo como:
```
[
  "ns-1234.awsdns-56.org",
  "ns-789.awsdns-12.com",
  "ns-345.awsdns-67.net",
  "ns-678.awsdns-90.co.uk"
]
```

### **3. Configurar Nameservers no Registro.br**

1. Acesse: https://registro.br/
2. Faça login com suas credenciais
3. Vá em **"Meus Domínios"** → Selecione `velure.app.br`
4. Clique em **"Alterar Servidores DNS"**
5. Selecione **"Usar outros servidores DNS"**
6. Adicione os 4 nameservers da AWS (sem o ponto final):
   ```
   ns-1234.awsdns-56.org
   ns-789.awsdns-12.com
   ns-345.awsdns-67.net
   ns-678.awsdns-90.co.uk
   ```
7. Salve as alterações

**⏰ Aguarde**: A propagação DNS pode levar de 24h a 48h.

### **4. Verificar Propagação DNS**

```bash
# Verificar nameservers atuais do domínio
dig NS velure.app.br +short

# Ou use ferramentas online:
# - https://dnschecker.org/
# - https://www.whatsmydns.net/
```

Quando os nameservers AWS aparecerem, a propagação está completa! ✅

### **5. Deploy da Aplicação (Criar ALB via Helm)**

Após nameservers configurados, faça o deploy da aplicação:

```bash
# Deploy dos charts Helm (cria ALB)
cd /Users/icl00ud/repos/velure
./scripts/deploy-eks.sh
```

Aguarde o ALB ser criado:
```bash
# Verificar se ALB foi criado
kubectl get ingress -A

# Verificar tags do ALB
aws elbv2 describe-load-balancers \
  --query "LoadBalancers[?contains(LoadBalancerName, 'velure')].DNSName"
```

### **6. Segundo Deploy do Terraform (Criar DNS Record)**

Agora que o ALB existe, podemos criar o registro DNS:

```bash
cd /Users/icl00ud/repos/velure/infrastructure/terraform

# Alterar create_dns_record para true
sed -i '' 's/create_dns_record   = false/create_dns_record   = true/' main.tf

# Aplicar
terraform plan  # Deve mostrar +1 aws_route53_record.main
terraform apply
```

### **7. Testar o Domínio**

```bash
# Verificar registro A
dig A velure.app.br +short

# Testar HTTPS
curl -I https://velure.app.br

# Abrir no navegador
open https://velure.app.br
```

## 🔍 Troubleshooting

### Erro: "No Load Balancer found with tags"

**Causa**: ALB ainda não foi criado ou tags estão incorretas.

**Solução**:
```bash
# Listar todos ALBs
aws elbv2 describe-load-balancers

# Verificar tags
aws elbv2 describe-tags --resource-arns <ALB-ARN>
```

O ALB deve ter a tag:
```
elbv2.k8s.aws/cluster = velure-production
```

### Propagação DNS lenta

**Causa**: Registro.br pode levar até 48h para propagar.

**Solução**:
- Aguardar pacientemente
- Verificar com `dig NS velure.app.br @8.8.8.8`
- Usar https://dnschecker.org/ para ver propagação global

### Certificado SSL não funciona

**Causa**: ALB precisa de certificado ACM configurado.

**Solução**: Configurar AWS Certificate Manager (ACM) no Helm chart:
```yaml
# values.yaml do chart ui-service
ingress:
  annotations:
    alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:...
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTPS": 443}]'
```

## 💰 Custos

- **Route53 Hosted Zone**: $0.50/mês
- **Route53 Queries**: $0.40/milhão de queries (primeiros 1B/mês)
- **Health Check** (desabilitado): $0.50/mês economizado ✅

**Total**: ~$0.50/mês

## 📚 Referências

- [Route53 Documentation](https://docs.aws.amazon.com/route53/)
- [Registro.br - Alterar DNS](https://registro.br/ajuda/dns/)
- [AWS Load Balancer Controller](https://kubernetes-sigs.github.io/aws-load-balancer-controller/)
