# SigaCore Gateway

Um sistema de gateway reverso com autenticação baseado em microserviços, incluindo whitelist de usuários e middlewares de segurança.

## 🏗️ Arquitetura

O projeto está organizado seguindo as convenções de projetos Go e separação de responsabilidades:

```
api--sigacore-gateway/
├── cmd/                    # Pontos de entrada da aplicação
│   ├── auth-service/       # Serviço de autenticação independente
│   ├── gateway/           # Gateway independente
│   └── combined/          # Execução combinada (desenvolvimento)
├── internal/              # Código interno da aplicação
│   ├── auth/              # Módulo de autenticação
│   │   ├── handlers/      # Handlers HTTP
│   │   ├── services/      # Lógica de negócio
│   │   └── models/        # Estruturas de dados
│   ├── gateway/           # Módulo do gateway
│   │   ├── middleware/    # Middlewares específicos
│   │   ├── proxy/         # Utilitários de proxy
│   │   ├── router/        # Configuração de rotas
│   │   └── server/        # Servidor do gateway
│   └── shared/            # Código compartilhado
│       ├── config/        # Configuração
│       ├── database/      # Abstração de banco
│       └── middleware/    # Middlewares compartilhados
├── pkg/                   # Código reutilizável
│   ├── token/             # Gerenciamento de tokens
│   └── utils/             # Utilitários gerais
├── db/                    # Banco de dados
│   ├── migration/         # Migrações SQL
│   ├── query/             # Queries SQL
│   └── sqlc/              # Código gerado pelo SQLC
└── api/                   # Definições de API (legado)
```

## 🚀 Como Executar

### Pré-requisitos

1. **Go 1.24+**
2. **PostgreSQL**
3. **Ferramentas de desenvolvimento** (opcional):
   ```bash
   make install-tools
   ```

### Configuração

1. **Configure o banco de dados** editando `app.env`:
   ```env
   DB_SOURCE=postgresql://root:secret@localhost:5432/sigacore?sslmode=disable
   TOKEN_SYMMETRIC_KEY=12345678901234567890123456789012
   ACCESS_TOKEN_DURATION=15m
   REFRESH_TOKEN_DURATION=24h
   AUTH_SERVER_ADDRESS=0.0.0.0:8080
   GATEWAY_SERVER_ADDRESS=0.0.0.0:8081
   ALLOWED_IPS=127.0.0.1,172.17.0.1,::1
   ```

2. **Execute as migrações**:
   ```bash
   make migrate-up
   ```

### Executando os Serviços

#### Opção 1: Desenvolvimento (Recomendado)
Executa ambos os serviços simultaneamente:
```bash
make dev
# ou
make run-combined
```

#### Opção 2: Serviços Separados
Em terminais diferentes:
```bash
# Terminal 1 - Serviço de Autenticação
make run-auth

# Terminal 2 - Gateway
make run-gateway
```

#### Opção 3: Executáveis Compilados
```bash
make build
./bin/combined
```

## 📡 Endpoints

### Gateway (Porta 8081)

- **Rotas Públicas (Autenticação)**:
  - `POST /auth/users` - Criar usuário
  - `POST /auth/users/login` - Login
  - `POST /auth/token/renew` - Renovar token

- **Rotas Protegidas**:
  - `GET /auth/users/:username` - Obter usuário
  - `GET /users/*` - Serviço de usuários
  - `GET /reports/*` - Serviço de relatórios
  - `GET /notifications/*` - Serviço de notificações

- **Utilitárias**:
  - `GET /health` - Health check

### Serviço de Autenticação (Porta 8080)

Endpoints diretos para desenvolvimento/teste:
- `POST /users` - Criar usuário
- `POST /users/login` - Login
- `POST /token/renew` - Renovar token
- `GET /users/:username` - Obter usuário (protegido)
- `GET /health` - Health check

## 🔒 Segurança

### Whitelist de Usuários
- Apenas usuários com `is_whitelisted=true` podem fazer login
- Configure usuarios whitelistados ao criar: `{"is_whitelisted": true}`

### Middlewares
- **IP Whitelist**: Apenas IPs configurados em `ALLOWED_IPS`
- **Rate Limiting**: 5 requisições/segundo, burst de 10
- **Autenticação**: Tokens PASETO para rotas protegidas

### Fluxo de Autenticação
1. **Criar usuário**: `POST /auth/users` (publico)
2. **Login**: `POST /auth/users/login` (publico, retorna tokens)
3. **Acessar rotas protegidas**: Header `Authorization: Bearer <token>`
4. **Renovar token**: `POST /auth/token/renew` (publico, com refresh token)

## 🛠️ Desenvolvimento

### Comandos Úteis

```bash
# Compilar tudo
make build

# Executar testes
make test

# Gerar código SQLC
make sqlc

# Criar nova migração
make migrate-create name=add_new_feature

# Docker
make docker-build
make docker-run

# Limpar builds
make clean

# Ver todos os comandos
make help
```

### Estrutura de Pastas

#### `cmd/`
Pontos de entrada da aplicação. Cada subdiretório contém um `main.go` para um executável diferente.

#### `internal/`
Código interno que não deve ser importado por outros projetos:
- **`auth/`**: Lógica de autenticação e gerenciamento de usuários
- **`gateway/`**: Lógica de proxy reverso e roteamento
- **`shared/`**: Código compartilhado entre módulos

#### `pkg/`
Código que pode ser reutilizado por outros projetos:
- **`token/`**: Gerenciamento de tokens JWT/PASETO
- **`utils/`**: Utilitários gerais

### Adicionando Novos Microserviços

1. **Configure o endereço** em `app.env`:
   ```env
   NEW_SERVICE_ADDRESS=http://localhost:8085
   ```

2. **Atualize a configuração** em `internal/shared/config/config.go`:
   ```go
   type Config struct {
       // ... existing fields
       NewServiceAddress string `mapstructure:"NEW_SERVICE_ADDRESS"`
   }
   ```

3. **Adicione a rota** em `internal/gateway/router/router.go`:
   ```go
   newServiceProxy, err := proxy.NewReverseProxy(r.config.NewServiceAddress)
   // ... add route mounting
   ```

## 🐳 Docker

```dockerfile
# Exemplo de uso do Dockerfile incluído
docker build -t sigacore-gateway .
docker run -p 8081:8081 sigacore-gateway
```

## 📝 Licença

Este projeto é licenciado sob [LICENSE].

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📞 Suporte

Para suporte, envie um email para [email] ou abra uma issue no GitHub. 