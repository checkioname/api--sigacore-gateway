# 🚀 Deploy Automático do Gateway

Este repositório está configurado para deploy automático no EC2 usando GitHub Actions.

## 📋 Pré-requisitos

### 1. EC2 Instance
- ✅ Instância EC2 rodando (Amazon Linux 2 recomendado)
- ✅ Security Group liberando porta 8080
- ✅ Chave SSH (.pem) para acesso

### 2. Banco de Dados
- ✅ PostgreSQL configurado (RDS ou local)
- ✅ Database `sigacore` criada
- ✅ Migrations executadas

## 🔐 Configurar Secrets no GitHub

Vá em **Settings → Secrets and variables → Actions** e adicione:

```
EC2_SSH_KEY          # Conteúdo da sua chave .pem
EC2_HOST             # IP público do EC2 (ex: 54.123.45.67)
EC2_USER             # Usuário SSH (ex: ec2-user)
DB_HOST              # Host do banco (ex: localhost ou RDS endpoint)
DB_PORT              # Porta do banco (ex: 5432)
DB_USER              # Usuário do banco
DB_PASSWORD          # Senha do banco
DB_NAME              # Nome do banco (ex: sigacore)
JWT_SECRET           # Chave secreta JWT (gere uma aleatória)
SIGACORE_SERVICE_URL # URL do microservice (ex: http://localhost:8081)
SIGADOCS_SERVICE_URL # URL do docs service (ex: http://localhost:8082)
```

## 🎯 Como Funciona

### Deploy Automático
1. **Push para main/master** → Dispara deploy automaticamente
2. **Pull Request** → Roda testes (sem deploy)
3. **Manual** → Pode disparar via "Actions" tab

### Processo do Deploy
```
1. 🧪 Roda testes Go
2. 🔨 Build da aplicação
3. 📡 Conecta no EC2 via SSH
4. 📦 Envia código para /opt/sigacore-gateway
5. 🐳 Build e start do Docker container
6. 🏥 Health check na porta 8080
7. ✅ Notifica sucesso/falha
```

## 📁 Estrutura no EC2

```
/opt/sigacore-gateway/
├── cmd/
├── internal/
├── go.mod
├── go.sum
├── Dockerfile.production
├── docker-compose.production.yml
├── .env                    # Criado automaticamente
└── bin/
    └── gateway            # Binário compilado
```

## 🐳 Duas Opções de Deploy

### Opção 1: Deploy Direto (deploy.yml)
- Compila Go diretamente no EC2
- Usa systemd para gerenciar o serviço
- Mais leve, menos dependências

### Opção 2: Deploy com Docker (deploy-docker.yml)
- Usa Docker containers
- Mais isolado e portável
- Recomendado para produção

## 🔧 Comandos Úteis no EC2

### Verificar Status (Opção 1 - Systemd)
```bash
sudo systemctl status sigacore-gateway
sudo journalctl -u sigacore-gateway -f
```

### Verificar Status (Opção 2 - Docker)
```bash
cd /opt/sigacore-gateway
docker-compose -f docker-compose.production.yml ps
docker-compose -f docker-compose.production.yml logs -f
```

### Restart Manual
```bash
# Systemd
sudo systemctl restart sigacore-gateway

# Docker
docker-compose -f docker-compose.production.yml restart
```

## 🏥 Health Check

O Gateway expõe um endpoint de saúde:
```bash
curl http://SEU-IP:8080/health
```

Resposta esperada:
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## 🐛 Troubleshooting

### Deploy Falhou?
1. Verifique os logs no GitHub Actions
2. Confirme se todos os secrets estão configurados
3. Teste conexão SSH manualmente:
   ```bash
   ssh -i sua-chave.pem ec2-user@seu-ip
   ```

### Gateway não responde?
1. Verifique se a porta 8080 está liberada no Security Group
2. Confirme se o banco está acessível
3. Verifique logs:
   ```bash
   # Systemd
   sudo journalctl -u sigacore-gateway -n 50
   
   # Docker
   docker-compose -f docker-compose.production.yml logs --tail=50
   ```

### Banco não conecta?
1. Verifique se o Security Group do RDS permite conexão do EC2
2. Confirme as credenciais nos secrets
3. Teste conexão manual:
   ```bash
   psql -h SEU-DB-HOST -U SEU-USER -d sigacore
   ```

## 🔄 Rollback

Se algo der errado, você pode fazer rollback:

1. **Via GitHub**: Faça revert do commit e push
2. **Manual no EC2**:
   ```bash
   # Systemd
   sudo systemctl stop sigacore-gateway
   # Restaure versão anterior e restart
   
   # Docker
   docker-compose -f docker-compose.production.yml down
   # Restaure código anterior e up novamente
   ```

## 📊 Monitoramento

### Logs em Tempo Real
```bash
# Systemd
sudo journalctl -u sigacore-gateway -f

# Docker
docker-compose -f docker-compose.production.yml logs -f
```

### Métricas do Sistema
```bash
# CPU e Memória
htop

# Espaço em disco
df -h

# Conexões de rede
netstat -tlnp | grep :8080
```

---

## 🎉 Pronto!

Agora toda vez que você fizer `git push`, o Gateway será automaticamente deployado no seu EC2! 🚀
