#!/bin/bash

# =============================================================================
# Script de Deploy Econômico para AWS
# SigaIsenções - Versão Econômica (EC2 + RDS)
# =============================================================================

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configurações
AWS_REGION="us-east-1"
PROJECT_NAME="sigaisencoes"
INSTANCE_TYPE="t3.micro"
KEY_NAME="${PROJECT_NAME}-key"
DB_PASSWORD=$(openssl rand -base64 12)
TOKEN_KEY=$(openssl rand -base64 24)

# Função para log
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ❌ $1${NC}"
}

warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] ⚠️  $1${NC}"
}

# Verificar dependências
check_dependencies() {
    log "🔍 Verificando dependências..."
    
    # AWS CLI
    if ! command -v aws &> /dev/null; then
        error "AWS CLI não encontrado. Instale: https://aws.amazon.com/cli/"
        exit 1
    fi
    
    # Docker
    if ! command -v docker &> /dev/null; then
        error "Docker não encontrado. Instale: https://docker.com"
        exit 1
    fi
    
    # Credenciais AWS
    if ! aws sts get-caller-identity > /dev/null 2>&1; then
        error "AWS não configurado. Execute: aws configure"
        exit 1
    fi
    
    log "✅ Dependências verificadas"
}

# Obter VPC padrão
get_default_vpc() {
    log "🌐 Obtendo VPC padrão..."
    
    VPC_ID=$(aws ec2 describe-vpcs \
        --filters "Name=is-default,Values=true" \
        --query 'Vpcs[0].VpcId' \
        --output text \
        --region $AWS_REGION)
    
    if [ "$VPC_ID" = "None" ] || [ -z "$VPC_ID" ]; then
        error "VPC padrão não encontrada. Crie uma VPC primeiro."
        exit 1
    fi
    
    SUBNET_ID=$(aws ec2 describe-subnets \
        --filters "Name=vpc-id,Values=$VPC_ID" "Name=default-for-az,Values=true" \
        --query 'Subnets[0].SubnetId' \
        --output text \
        --region $AWS_REGION)
    
    log "✅ VPC: $VPC_ID, Subnet: $SUBNET_ID"
}

# Criar Key Pair
create_key_pair() {
    log "🔑 Criando Key Pair..."
    
    if aws ec2 describe-key-pairs --key-names $KEY_NAME --region $AWS_REGION > /dev/null 2>&1; then
        warning "Key pair $KEY_NAME já existe"
    else
        aws ec2 create-key-pair \
            --key-name $KEY_NAME \
            --query 'KeyMaterial' \
            --output text \
            --region $AWS_REGION > ${KEY_NAME}.pem
        
        chmod 400 ${KEY_NAME}.pem
        log "✅ Key pair criado: ${KEY_NAME}.pem"
    fi
}

# Criar Security Groups
create_security_groups() {
    log "🛡️ Criando Security Groups..."
    
    # EC2 Security Group
    EC2_SG_ID=$(aws ec2 create-security-group \
        --group-name "${PROJECT_NAME}-ec2-sg" \
        --description "SigaIsencoes EC2 Security Group" \
        --vpc-id $VPC_ID \
        --query 'GroupId' \
        --output text \
        --region $AWS_REGION 2>/dev/null || \
    aws ec2 describe-security-groups \
        --filters "Name=group-name,Values=${PROJECT_NAME}-ec2-sg" \
        --query 'SecurityGroups[0].GroupId' \
        --output text \
        --region $AWS_REGION)
    
    # RDS Security Group
    RDS_SG_ID=$(aws ec2 create-security-group \
        --group-name "${PROJECT_NAME}-rds-sg" \
        --description "SigaIsencoes RDS Security Group" \
        --vpc-id $VPC_ID \
        --query 'GroupId' \
        --output text \
        --region $AWS_REGION 2>/dev/null || \
    aws ec2 describe-security-groups \
        --filters "Name=group-name,Values=${PROJECT_NAME}-rds-sg" \
        --query 'SecurityGroups[0].GroupId' \
        --output text \
        --region $AWS_REGION)
    
    # Regras EC2 Security Group
    aws ec2 authorize-security-group-ingress \
        --group-id $EC2_SG_ID \
        --protocol tcp \
        --port 22 \
        --cidr 0.0.0.0/0 \
        --region $AWS_REGION 2>/dev/null || true
        
    aws ec2 authorize-security-group-ingress \
        --group-id $EC2_SG_ID \
        --protocol tcp \
        --port 80 \
        --cidr 0.0.0.0/0 \
        --region $AWS_REGION 2>/dev/null || true
        
    aws ec2 authorize-security-group-ingress \
        --group-id $EC2_SG_ID \
        --protocol tcp \
        --port 443 \
        --cidr 0.0.0.0/0 \
        --region $AWS_REGION 2>/dev/null || true
    
    # Regras RDS Security Group (apenas da EC2)
    aws ec2 authorize-security-group-ingress \
        --group-id $RDS_SG_ID \
        --protocol tcp \
        --port 5432 \
        --source-group $EC2_SG_ID \
        --region $AWS_REGION 2>/dev/null || true
    
    log "✅ Security Groups criados: EC2($EC2_SG_ID), RDS($RDS_SG_ID)"
}

# Criar RDS
create_rds() {
    log "🗄️ Criando RDS PostgreSQL..."
    
    DB_IDENTIFIER="${PROJECT_NAME}-db"
    
    # Verificar se já existe
    if aws rds describe-db-instances \
        --db-instance-identifier $DB_IDENTIFIER \
        --region $AWS_REGION > /dev/null 2>&1; then
        warning "RDS $DB_IDENTIFIER já existe"
        RDS_ENDPOINT=$(aws rds describe-db-instances \
            --db-instance-identifier $DB_IDENTIFIER \
            --query 'DBInstances[0].Endpoint.Address' \
            --output text \
            --region $AWS_REGION)
    else
        log "Criando nova instância RDS..."
        
        aws rds create-db-instance \
            --db-instance-identifier $DB_IDENTIFIER \
            --db-instance-class db.t3.micro \
            --engine postgres \
            --engine-version 15.4 \
            --allocated-storage 20 \
            --storage-type gp2 \
            --master-username admin \
            --master-user-password "$DB_PASSWORD" \
            --vpc-security-group-ids $RDS_SG_ID \
            --backup-retention-period 7 \
            --no-multi-az \
            --publicly-accessible \
            --region $AWS_REGION
        
        log "⏳ Aguardando RDS ficar disponível (pode levar 5-10 minutos)..."
        aws rds wait db-instance-available \
            --db-instance-identifier $DB_IDENTIFIER \
            --region $AWS_REGION
        
        RDS_ENDPOINT=$(aws rds describe-db-instances \
            --db-instance-identifier $DB_IDENTIFIER \
            --query 'DBInstances[0].Endpoint.Address' \
            --output text \
            --region $AWS_REGION)
    fi
    
    log "✅ RDS criado: $RDS_ENDPOINT"
}

# Criar arquivo de configuração
create_env_file() {
    log "📝 Criando arquivos de configuração..."
    
    cat > .env.production << EOF
# Configurações de produção
DB_PASSWORD=$DB_PASSWORD
TOKEN_KEY=$TOKEN_KEY
RDS_ENDPOINT=$RDS_ENDPOINT
EC2_SG_ID=$EC2_SG_ID
RDS_SG_ID=$RDS_SG_ID
EOF

    cat > docker-compose.prod.yml << 'EOF'
version: '3.8'

services:
  gateway:
    build:
      context: ./api--sigacore-gateway
      dockerfile: Dockerfile.production
    container_name: sigaisencoes-gateway
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=production
      - DB_SOURCE=postgresql://admin:${DB_PASSWORD}@${RDS_ENDPOINT}:5432/sigacore?sslmode=require
      - TOKEN_SYMMETRIC_KEY=${TOKEN_KEY}
      - ACCESS_TOKEN_DURATION=15m
      - REFRESH_TOKEN_DURATION=24h
      - AUTH_SERVER_ADDRESS=0.0.0.0:8080
      - GATEWAY_SERVER_ADDRESS=0.0.0.0:8081
      - ALLOWED_IPS=0.0.0.0/0
      - GIN_MODE=release
    networks:
      - app-network

  core:
    build:
      context: ./api--sigacore-service
      dockerfile: Dockerfile.production
    container_name: sigaisencoes-core
    restart: unless-stopped
    ports:
      - "8081:8081"
    environment:
      - ENVIRONMENT=production
      - DATABASE_HOST=${RDS_ENDPOINT}
      - DATABASE_USER=admin
      - DATABASE_PASSWORD=${DB_PASSWORD}
      - DATABASE_NAME=sigacore
      - DATABASE_PORT=5432
      - GIN_MODE=release
    networks:
      - app-network

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
      - CORE_SERVICE_URL=http://core:8081
      - GIN_MODE=release
    depends_on:
      - core
    networks:
      - app-network

  nginx:
    image: nginx:alpine
    container_name: sigaisencoes-nginx
    restart: unless-stopped
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - gateway
      - core
      - docs
    networks:
      - app-network

networks:
  app-network:
    driver: bridge
EOF

    cat > nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    upstream gateway {
        server gateway:8080;
    }
    
    upstream core {
        server core:8081;
    }
    
    upstream docs {
        server docs:8082;
    }

    server {
        listen 80;
        server_name _;

        # API Gateway/Auth
        location /auth/ {
            proxy_pass http://gateway/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }

        # API Core (Clientes)
        location /api/ {
            proxy_pass http://core/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }

        # Documents API
        location /docs/ {
            proxy_pass http://docs/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }

        # Health check
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }

        # Frontend redirect (opcional)
        location / {
            return 302 https://your-frontend-domain.com;
        }
    }
}
EOF

    log "✅ Arquivos de configuração criados"
}

# Criar User Data para EC2
create_user_data() {
    cat > user-data.sh << 'EOF'
#!/bin/bash

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

# Criar diretório da aplicação
mkdir -p /opt/sigaisencoes
chown ec2-user:ec2-user /opt/sigaisencoes

# Instalar Git
yum install -y git

# Log de conclusão
echo "EC2 setup completo: $(date)" > /opt/setup-complete.log
EOF
}

# Criar instância EC2
create_ec2() {
    log "🖥️ Criando instância EC2..."
    
    create_user_data
    
    # Verificar se já existe
    EXISTING_INSTANCE=$(aws ec2 describe-instances \
        --filters "Name=tag:Name,Values=${PROJECT_NAME}-app" "Name=instance-state-name,Values=running,pending" \
        --query 'Reservations[0].Instances[0].InstanceId' \
        --output text \
        --region $AWS_REGION 2>/dev/null)
    
    if [ "$EXISTING_INSTANCE" != "None" ] && [ ! -z "$EXISTING_INSTANCE" ]; then
        warning "Instância EC2 já existe: $EXISTING_INSTANCE"
        INSTANCE_ID=$EXISTING_INSTANCE
    else
        log "Criando nova instância EC2..."
        
        INSTANCE_ID=$(aws ec2 run-instances \
            --image-id ami-0c02fb55956c7d316 \
            --count 1 \
            --instance-type $INSTANCE_TYPE \
            --key-name $KEY_NAME \
            --security-group-ids $EC2_SG_ID \
            --subnet-id $SUBNET_ID \
            --associate-public-ip-address \
            --user-data file://user-data.sh \
            --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=${PROJECT_NAME}-app}]" \
            --query 'Instances[0].InstanceId' \
            --output text \
            --region $AWS_REGION)
        
        log "⏳ Aguardando instância ficar disponível..."
        aws ec2 wait instance-running \
            --instance-ids $INSTANCE_ID \
            --region $AWS_REGION
    fi
    
    # Obter IP público
    EC2_IP=$(aws ec2 describe-instances \
        --instance-ids $INSTANCE_ID \
        --query 'Reservations[0].Instances[0].PublicIpAddress' \
        --output text \
        --region $AWS_REGION)
    
    log "✅ EC2 criado: $INSTANCE_ID ($EC2_IP)"
}

# Deploy da aplicação
deploy_application() {
    log "🚀 Fazendo deploy da aplicação..."
    
    # Aguardar EC2 estar completamente pronta
    log "⏳ Aguardando EC2 finalizar setup..."
    sleep 60
    
    # Testar conexão SSH
    max_attempts=10
    attempt=1
    while [ $attempt -le $max_attempts ]; do
        if ssh -i ${KEY_NAME}.pem -o ConnectTimeout=10 -o StrictHostKeyChecking=no ec2-user@$EC2_IP "echo 'SSH OK'" 2>/dev/null; then
            log "✅ Conexão SSH estabelecida"
            break
        else
            warning "Tentativa SSH $attempt/$max_attempts falhou, aguardando..."
            sleep 30
            ((attempt++))
        fi
    done
    
    if [ $attempt -gt $max_attempts ]; then
        error "Não foi possível conectar via SSH após $max_attempts tentativas"
        exit 1
    fi
    
    # Transferir arquivos
    log "📁 Transferindo arquivos..."
    ssh -i ${KEY_NAME}.pem -o StrictHostKeyChecking=no ec2-user@$EC2_IP "sudo mkdir -p /opt/sigaisencoes && sudo chown ec2-user:ec2-user /opt/sigaisencoes"
    
    # Enviar arquivos essenciais
    scp -i ${KEY_NAME}.pem -o StrictHostKeyChecking=no .env.production ec2-user@$EC2_IP:/opt/sigaisencoes/
    scp -i ${KEY_NAME}.pem -o StrictHostKeyChecking=no docker-compose.prod.yml ec2-user@$EC2_IP:/opt/sigaisencoes/
    scp -i ${KEY_NAME}.pem -o StrictHostKeyChecking=no nginx.conf ec2-user@$EC2_IP:/opt/sigaisencoes/
    
    # Enviar código fonte
    tar czf app-source.tar.gz api--sigacore-gateway/ api--sigacore-service/ worker--sigacore-docs/
    scp -i ${KEY_NAME}.pem -o StrictHostKeyChecking=no app-source.tar.gz ec2-user@$EC2_IP:/opt/sigaisencoes/
    
    # Executar deploy
    log "🐳 Iniciando containers..."
    ssh -i ${KEY_NAME}.pem -o StrictHostKeyChecking=no ec2-user@$EC2_IP << EOF
cd /opt/sigaisencoes
tar xzf app-source.tar.gz
source .env.production
sudo docker-compose -f docker-compose.prod.yml up -d --build
EOF

    # Cleanup
    rm -f user-data.sh app-source.tar.gz
    
    log "✅ Deploy concluído!"
}

# Deploy do frontend no S3
deploy_frontend() {
    log "🌐 Deploy do frontend no S3..."
    
    BUCKET_NAME="${PROJECT_NAME}-frontend-$(date +%Y%m%d)"
    
    # Verificar se bucket existe
    if aws s3 ls "s3://$BUCKET_NAME" > /dev/null 2>&1; then
        warning "Bucket $BUCKET_NAME já existe"
    else
        log "Criando bucket S3: $BUCKET_NAME"
        aws s3 mb s3://$BUCKET_NAME --region $AWS_REGION
        
        # Configurar para website
        aws s3 website s3://$BUCKET_NAME \
            --index-document index.html \
            --error-document 404.html
        
        # Política pública
        cat > bucket-policy.json << EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "PublicReadGetObject",
            "Effect": "Allow",
            "Principal": "*",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::$BUCKET_NAME/*"
        }
    ]
}
EOF
        aws s3api put-bucket-policy --bucket $BUCKET_NAME --policy file://bucket-policy.json
        rm bucket-policy.json
    fi
    
    # Build e upload do frontend
    if [ -d "frontend-sigaisencoes" ]; then
        log "📦 Fazendo build do frontend..."
        cd frontend-sigaisencoes
        
        # Configurar variáveis de ambiente
        cat > .env.local << EOF
NEXT_PUBLIC_API_URL=http://$EC2_IP/api
NEXT_PUBLIC_DOCS_API_URL=http://$EC2_IP/docs
NEXT_PUBLIC_AUTH_API_URL=http://$EC2_IP/auth
EOF
        
        npm ci > /dev/null 2>&1
        npm run build > /dev/null 2>&1
        
        # Upload para S3
        aws s3 sync out/ s3://$BUCKET_NAME --delete
        
        cd ..
        
        FRONTEND_URL="http://$BUCKET_NAME.s3-website-$AWS_REGION.amazonaws.com"
        log "✅ Frontend disponível em: $FRONTEND_URL"
    else
        warning "Diretório frontend-sigaisencoes não encontrado, pulando deploy do frontend"
    fi
}

# Função principal
main() {
    echo -e "${BLUE}"
    echo "=============================================="
    echo "💰 SigaIsenções - Deploy Econômico AWS"
    echo "=============================================="
    echo -e "${NC}"
    
    # Verificações iniciais
    check_dependencies
    get_default_vpc
    
    # Confirmação
    echo -e "${YELLOW}💰 Este deploy custará aproximadamente $29/mês${NC}"
    echo -e "${YELLOW}⚠️  Continuar com o deploy? (y/N)${NC}"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Deploy cancelado.${NC}"
        exit 0
    fi
    
    # Executar deploy
    log "🎯 Iniciando deploy econômico..."
    
    create_key_pair
    create_security_groups
    create_rds
    create_env_file
    create_ec2
    deploy_application
    deploy_frontend
    
    echo -e "${GREEN}"
    echo "=============================================="
    echo "✅ Deploy Econômico Concluído!"
    echo "=============================================="
    echo -e "${NC}"
    
    log "📋 Informações do Deploy:"
    echo "🖥️  Instância EC2: $INSTANCE_ID"
    echo "🌐 IP Público: $EC2_IP"
    echo "🗄️  RDS Endpoint: $RDS_ENDPOINT"
    echo "🔑 Key Pair: ${KEY_NAME}.pem"
    echo ""
    log "🔗 URLs da Aplicação:"
    echo "📊 API Gateway: http://$EC2_IP/auth"
    echo "👥 API Core: http://$EC2_IP/api"
    echo "📄 API Docs: http://$EC2_IP/docs"
    if [ ! -z "$FRONTEND_URL" ]; then
        echo "🌐 Frontend: $FRONTEND_URL"
    fi
    echo ""
    log "💰 Custo estimado: ~$29 USD/mês"
    echo ""
    log "🔒 Credenciais salvas em: .env.production"
    
    warning "IMPORTANTE: Guarde o arquivo ${KEY_NAME}.pem em local seguro!"
}

# Verificar se está sendo executado como script
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
