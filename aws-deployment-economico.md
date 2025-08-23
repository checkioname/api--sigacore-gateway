# 💰 Deploy Econômico AWS - SigaIsenções

## 📊 Novo Custo Estimado: $35-50 USD/mês

**Redução de 50-60% nos custos** removendo Load Balancer e usando EC2 simples.

---

## 🏗️ Arquitetura Econômica

```
Internet -> Route 53 (Opcional) -> EC2 Instance
                                 -> RDS PostgreSQL
         -> S3 + CloudFront (Frontend)
```

### **Principais Mudanças:**
- ✅ **EC2 t3.micro** ao invés de ECS Fargate
- ✅ **Sem ALB** (economia de $17/mês)
- ✅ **Single-AZ RDS** inicialmente
- ✅ **Docker Compose** na EC2
- ✅ **Nginx** como proxy reverso

---

## 💰 Novo Detalhamento de Custos

### 1. **Compute - EC2 t3.micro**
```
t3.micro: $0.0104 por hora × 730h = $7.59/mês
EBS GP2: 20GB × $0.10 = $2.00/mês

Total EC2: $9.59/mês
```

### 2. **Database - RDS t3.micro** 
```
db.t3.micro: $0.017 por hora × 730h = $12.41/mês
Storage GP2: 20GB × $0.115 = $2.30/mês
Backup: ~5GB × $0.095 = $0.48/mês

Total RDS: $15.19/mês
```

### 3. **Frontend - S3 + CloudFront**
```
S3 + CloudFront: $1.24/mês
(mesmo custo anterior)
```

### 4. **Network & Outros**
```
Data Transfer: $0.50/mês
Route 53 (opcional): $0.54/mês
CloudWatch básico: $2.00/mês

Total Outros: $3.04/mês
```

### **💰 TOTAL: $29.06/mês**
**Com Route 53: $29.60/mês**

---

## 🐳 Deploy com Docker Compose na EC2

### 1. **Configuração da EC2**

#### Launch da Instância
```bash
# Criar Key Pair
aws ec2 create-key-pair \
  --key-name sigaisencoes-key \
  --query 'KeyMaterial' \
  --output text > sigaisencoes-key.pem

chmod 400 sigaisencoes-key.pem

# Launch EC2 Instance
aws ec2 run-instances \
  --image-id ami-0c02fb55956c7d316 \
  --count 1 \
  --instance-type t3.micro \
  --key-name sigaisencoes-key \
  --security-group-ids sg-your-security-group \
  --subnet-id subnet-your-subnet \
  --associate-public-ip-address \
  --user-data file://user-data.sh \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=SigaIsencoes-App}]'
```

#### User Data Script
```bash
#!/bin/bash
# user-data.sh

# Atualizar sistema
yum update -y

# Instalar Docker
yum install -y docker
systemctl start docker
systemctl enable docker
usermod -aG docker ec2-user

# Instalar Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Instalar Nginx
yum install -y nginx
systemctl start nginx
systemctl enable nginx

# Criar diretórios
mkdir -p /opt/sigaisencoes
cd /opt/sigaisencoes

# Clone do repositório (ou copiar arquivos)
# git clone your-repo .
```

### 2. **Docker Compose Simplificado**

```yaml
# docker-compose.prod-ec2.yml
version: '3.8'

services:
  # Gateway/Auth + Core em um container
  backend:
    build:
      context: .
      dockerfile: Dockerfile.backend-combined
    container_name: sigaisencoes-backend
    restart: unless-stopped
    ports:
      - "8080:8080"
      - "8081:8081"
    environment:
      - ENVIRONMENT=production
      - DB_SOURCE=postgresql://admin:${DB_PASSWORD}@${RDS_ENDPOINT}:5432/sigacore?sslmode=require
      - TOKEN_SYMMETRIC_KEY=${TOKEN_KEY}
      - GIN_MODE=release
    networks:
      - app-network

  # Document Worker
  docs:
    build:
      context: ./worker--sigacore-docs
      dockerfile: Dockerfile.production
    container_name: sigaisencoes-docs
    restart: unless-stopped
    ports:
      - "8082:8082"
    environment:
      - ENVIRONMENT=production
      - CORE_SERVICE_URL=http://backend:8081
    depends_on:
      - backend
    networks:
      - app-network

  # Nginx Proxy
  nginx:
    image: nginx:alpine
    container_name: sigaisencoes-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - backend
      - docs
    networks:
      - app-network

networks:
  app-network:
    driver: bridge
```

### 3. **Dockerfile Backend Combinado**

```dockerfile
# Dockerfile.backend-combined
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Build Gateway
COPY api--sigacore-gateway/ ./gateway/
WORKDIR /app/gateway
RUN go mod download
RUN CGO_ENABLED=0 go build -o gateway cmd/main.go

# Build Core
WORKDIR /app
COPY api--sigacore-service/ ./core/
WORKDIR /app/core
RUN go mod download
RUN CGO_ENABLED=0 go build -o core cmd/main.go

# Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates supervisor

WORKDIR /app

# Copiar binários
COPY --from=builder /app/gateway/gateway ./
COPY --from=builder /app/core/core ./

# Supervisor config
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

EXPOSE 8080 8081

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
```

### 4. **Configuração Nginx**

```nginx
# nginx.conf
events {
    worker_connections 1024;
}

http {
    upstream backend {
        server backend:8080;
        server backend:8081;
    }
    
    upstream docs {
        server docs:8082;
    }

    server {
        listen 80;
        server_name your-domain.com;

        # Redirect HTTP to HTTPS (opcional)
        # return 301 https://$server_name$request_uri;

        # API Gateway/Auth
        location /auth/ {
            proxy_pass http://backend:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        # API Core
        location /api/ {
            proxy_pass http://backend:8081;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        # Documents API
        location /docs/ {
            proxy_pass http://docs:8082;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        # Health checks
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}
```

### 5. **Security Groups Simplificados**

```bash
# EC2 Security Group
aws ec2 create-security-group \
  --group-name sigaisencoes-ec2-sg \
  --description "SigaIsencoes EC2 Security Group"

# Permitir SSH
aws ec2 authorize-security-group-ingress \
  --group-id sg-xxx \
  --protocol tcp \
  --port 22 \
  --cidr 0.0.0.0/0

# Permitir HTTP/HTTPS
aws ec2 authorize-security-group-ingress \
  --group-id sg-xxx \
  --protocol tcp \
  --port 80 \
  --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
  --group-id sg-xxx \
  --protocol tcp \
  --port 443 \
  --cidr 0.0.0.0/0

# RDS Security Group (apenas da EC2)
aws ec2 authorize-security-group-ingress \
  --group-id sg-rds-xxx \
  --protocol tcp \
  --port 5432 \
  --source-group sg-xxx
```

---

## 🚀 Script de Deploy Econômico

```bash
#!/bin/bash
# deploy-economico.sh

set -e

# Configurações
INSTANCE_TYPE="t3.micro"
KEY_NAME="sigaisencoes-key"
REGION="us-east-1"

# Criar RDS primeiro
create_rds() {
    echo "🗄️ Criando RDS PostgreSQL..."
    
    aws rds create-db-instance \
      --db-instance-identifier sigaisencoes-db \
      --db-instance-class db.t3.micro \
      --engine postgres \
      --engine-version 15.4 \
      --allocated-storage 20 \
      --storage-type gp2 \
      --master-username admin \
      --master-user-password "$(openssl rand -base64 12)" \
      --vpc-security-group-ids sg-rds-xxx \
      --backup-retention-period 7 \
      --no-multi-az \
      --region $REGION
}

# Deploy da aplicação
deploy_app() {
    echo "🚀 Fazendo deploy da aplicação..."
    
    # Conectar na EC2
    EC2_IP=$(aws ec2 describe-instances \
      --filters "Name=tag:Name,Values=SigaIsencoes-App" \
      --query 'Reservations[0].Instances[0].PublicIpAddress' \
      --output text)
    
    # Transferir arquivos
    scp -i sigaisencoes-key.pem -r . ec2-user@$EC2_IP:/opt/sigaisencoes/
    
    # Executar deploy
    ssh -i sigaisencoes-key.pem ec2-user@$EC2_IP << 'EOF'
cd /opt/sigaisencoes
sudo docker-compose -f docker-compose.prod-ec2.yml up -d
EOF

    echo "✅ Deploy concluído!"
    echo "🌐 Aplicação disponível em: http://$EC2_IP"
}

# Executar
create_rds
deploy_app
```

---

## 📊 Comparação de Custos

| Componente | Deploy Completo | Deploy Econômico | Economia |
|------------|----------------|------------------|----------|
| **Compute** | $47.50 (ECS) | $9.59 (EC2) | **$37.91** |
| **Load Balancer** | $17.20 | $0 | **$17.20** |
| **Database** | $15.19 | $15.19 | $0 |
| **Frontend** | $1.24 | $1.24 | $0 |
| **Outros** | $4.63 | $3.04 | $1.59 |
| **TOTAL** | **$85.76** | **$29.06** | **$56.70** |

### **💰 Economia de 66% nos custos!**

---

## ⚡ Benefícios da Versão Econômica

### ✅ **Vantagens**
- **Custo 66% menor**
- **Setup mais simples**
- **Menos componentes para gerenciar**
- **Ideal para MVP/testes**
- **Fácil monitoramento**

### ⚠️ **Limitações**
- **Single point of failure**
- **Escalabilidade manual**
- **Sem auto-scaling**
- **Backup manual da EC2**

---

## 🎯 Roadmap de Crescimento

### **Fase 1: MVP (0-50 usuários)**
**Custo: $29/mês**
- EC2 t3.micro
- RDS t3.micro Single-AZ
- Docker Compose

### **Fase 2: Crescimento (50-200 usuários)**
**Custo: $45/mês**
- EC2 t3.small
- RDS t3.small
- Backup automatizado

### **Fase 3: Escala (200+ usuários)**
**Custo: $85/mês**
- Migrar para ECS + ALB
- Multi-AZ RDS
- Auto-scaling

---

## 🛠️ Próximos Passos

1. **Executar deploy-economico.sh**
2. **Configurar domínio (opcional)**
3. **Configurar SSL com Let's Encrypt**
4. **Configurar backup básico**
5. **Monitoramento simples com CloudWatch**

Este deploy econômico é **perfeito para começar** e pode **evoluir gradualmente** conforme a necessidade!
