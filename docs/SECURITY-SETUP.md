# 🔐 Guia de Gestão Segura de Credenciais

Este guia explica como configurar e usar o sistema de gestão segura de credenciais implementado no SIGA Core Gateway.

## 📋 **ÍNDICE**

1. [Configuração Inicial](#configuração-inicial)
2. [Desenvolvimento](#desenvolvimento)
3. [Produção](#produção)
4. [Comandos Úteis](#comandos-úteis)
5. [Troubleshooting](#troubleshooting)
6. [Checklist de Segurança](#checklist-de-segurança)

---

## 🚀 **CONFIGURAÇÃO INICIAL**

### 1. **Setup para Desenvolvimento**

```bash
# 1. Configurar ferramentas
make install-tools

# 2. Gerar chave segura para desenvolvimento
make generate-key

# 3. Configurar arquivo app.env (já deve estar configurado)
# Verificar se TOKEN_SYMMETRIC_KEY foi atualizada com a chave gerada

# 4. Verificar configurações de segurança
make check-security

# 5. Executar sistema
make dev
```

### 2. **Verificar Configuração**

```bash
# Verificar se tudo está funcionando
curl http://localhost:8081/health

# Testar criação de usuário
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","full_name":"Test User","password":"TestPass123!"}'
```

---

## 💻 **DESENVOLVIMENTO**

### **Arquivo app.env**

O arquivo `app.env` é seguro para desenvolvimento e já está configurado com:

- ✅ Chave de desenvolvimento identificada como não-produção
- ✅ Validações de entrada habilitadas
- ✅ Configurações de localhost
- ✅ SSL desabilitado (OK para desenvolvimento)

### **Gerando Nova Chave**

```bash
# Gerar uma nova chave
make generate-key

# Gerar múltiplas chaves para rotação
make generate-key-multiple
```

### **Validação de Senha**

- **Desenvolvimento**: Mínimo 8 caracteres
- **Produção**: Deve conter maiúscula, minúscula, número e caractere especial

---

## 🏭 **PRODUÇÃO**

### **1. Preparar Configuração**

```bash
# 1. Criar template de produção
make setup-prod-example

# 2. Copiar template para local seguro (FORA do repositório)
cp deployment/production.env.example /path/seguro/production.env

# 3. Gerar chave de produção
make generate-key

# 4. Editar arquivo com valores reais
nano /path/seguro/production.env
```

### **2. Configurar Valores de Produção**

Edite `/path/seguro/production.env`:

```bash
# Ambiente
ENVIRONMENT=production

# Banco com SSL obrigatório
DB_SOURCE=postgresql://usuario_real:senha_forte@host-db:5432/sigacore_prod?sslmode=require

# Chave gerada pelo comando make generate-key
TOKEN_SYMMETRIC_KEY=sua_chave_gerada_32_caracteres

# IPs de produção
ALLOWED_IPS=203.0.113.100,198.51.100.0/24

# Serviços de produção
USER_SERVICE_ADDRESS=https://user-service.interno:443
# ... etc
```

### **3. Deploy de Produção**

```bash
# 1. Carregar configuração
source /path/seguro/production.env

# 2. Verificar configuração
make check-prod-config

# 3. Build da aplicação
make build

# 4. Executar
./bin/combined
```

### **4. Com Docker**

```bash
# Build da imagem
make docker-build

# Executar com variáveis de ambiente
docker run -d \
  -e ENVIRONMENT=production \
  -e DB_SOURCE="postgresql://..." \
  -e TOKEN_SYMMETRIC_KEY="$(cat /path/seguro/key.txt)" \
  -p 8081:8081 \
  sigacore-gateway:latest
```

---

## 🛠️ **COMANDOS ÚTEIS**

### **Segurança**

```bash
# Verificar configurações de segurança
make check-security

# Gerar chave segura
make generate-key

# Scan de vulnerabilidades
make security-scan

# Validar configuração de produção
make check-prod-config
```

### **Desenvolvimento**

```bash
# Setup completo
make setup-dev

# Executar em desenvolvimento
make dev

# Executar testes
make test

# Linter
make lint
```

### **Database**

```bash
# Migrar banco
make migrate-up

# Gerar código SQL
make sqlc
```

---

## 🐛 **TROUBLESHOOTING**

### **Erro: "unsafe symmetric key detected"**

```bash
# Problema: Usando chave padrão em produção
# Solução:
make generate-key
# Copie a chave gerada para TOKEN_SYMMETRIC_KEY
```

### **Erro: "TOKEN_SYMMETRIC_KEY must be exactly 32 characters"**

```bash
# Problema: Chave com tamanho incorreto
# Solução: Use o gerador de chaves
make generate-key
```

### **Erro: "localhost IPs detected in production environment"**

```bash
# Problema: IP localhost em produção
# Solução: Configure IPs reais em ALLOWED_IPS
export ALLOWED_IPS="203.0.113.100,198.51.100.0/24"
```

### **Erro: "SSL required for database connection in production"**

```bash
# Problema: SSL não configurado para banco
# Solução: Adicione sslmode=require na string de conexão
DB_SOURCE="postgresql://user:pass@host:5432/db?sslmode=require"
```

### **Erro: "password must be at least 8 characters with uppercase..."**

```bash
# Problema: Senha fraca em produção
# Solução: Use senha forte com:
# - Pelo menos 8 caracteres
# - Maiúscula e minúscula
# - Número
# - Caractere especial
```

---

## ✅ **CHECKLIST DE SEGURANÇA**

### **Desenvolvimento**
- [ ] `make setup-dev` executado com sucesso
- [ ] `make check-security` sem erros
- [ ] Chave diferente da padrão no app.env
- [ ] Testes passando

### **Produção**
- [ ] Template de produção criado e configurado
- [ ] Chave gerada criptograficamente
- [ ] SSL habilitado para banco de dados
- [ ] IPs de produção configurados
- [ ] Credenciais padrão removidas
- [ ] Variáveis de ambiente configuradas no sistema
- [ ] `make check-prod-config` sem erros
- [ ] HTTPS configurado (certificados válidos)
- [ ] Firewall configurado
- [ ] Monitoramento implementado
- [ ] Backup das configurações realizado

### **Operacional**
- [ ] Logs estruturados funcionando
- [ ] Rate limiting ativo
- [ ] Validação de entrada rigorosa
- [ ] Context timeouts configurados
- [ ] Error handling não vaza informações sensíveis
- [ ] Middleware de segurança ativo

---

## 🚨 **EM CASO DE COMPROMISSO DE SEGURANÇA**

1. **Imediato**: Trocar todas as chaves
2. **Revogar**: Todas as sessões ativas
3. **Investigar**: Logs de acesso
4. **Alertar**: Usuários afetados
5. **Documentar**: Incidente para análise

---

## 📞 **CONTATO**

Para dúvidas sobre implementação:
- Consulte a documentação técnica
- Verifique os logs da aplicação
- Execute `make help` para comandos disponíveis

---

**⚠️ IMPORTANTE**: Nunca comite credenciais reais no repositório. Use sempre o sistema de gestão de segredos adequado para seu ambiente. 