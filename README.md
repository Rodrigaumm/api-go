# Go API - Estrutura Reorganizada

## Estrutura do Projeto

```
go-api/
├── main.go                    # Ponto de entrada da aplicação
├── go.mod                     # Dependências do Go
├── go.sum                     # Checksums das dependências
├── docker-compose.yml         # Configuração do Docker
├── sqlc.yaml                  # Configuração do SQLC
├── queries.sql                # Queries SQL para o SQLC
├── schema.sql                 # Schema do banco de dados
├── internal/
│   ├── config/
│   │   └── config.go          # Configurações da aplicação
│   ├── db/                    # Arquivos gerados pelo SQLC
│   │   ├── db.go              # Interface do banco de dados
│   │   ├── models.go          # Modelos/structs do banco
│   │   ├── querier.go         # Interface das queries
│   │   └── queries.sql.go     # Implementação das queries
│   └── handlers/              # Handlers HTTP
│       ├── auth.go            # Autenticação e JWT
│       ├── middleware.go      # Middlewares
│       ├── user_handlers.go   # Handlers de usuários (renomeado)
│       └── process_handler.go # Handlers de processos
├── migrations/                # Migrações do banco (vazio)
└── queries/                   # Queries SQL organizadas
```

## Principais Mudanças

### 1. Organização por Pacotes
- **`internal/db/`**: Todos os arquivos gerados pelo SQLC foram movidos para este diretório
- **`internal/handlers/`**: Todos os handlers HTTP foram organizados neste diretório
- **`internal/config/`**: Configurações da aplicação

### 2. Renomeação de Arquivos
- `handlers.go` → `user_handlers.go` (para maior clareza)

### 3. Correção de Imports
- Todos os imports foram atualizados para usar os novos pacotes
- Referências aos tipos do banco de dados agora usam o prefixo `db.`

### 4. Separação de Responsabilidades
- **Arquivos gerados pelo SQLC**: Isolados em `internal/db/`
- **Lógica de negócio**: Organizada em `internal/handlers/`
- **Configurações**: Centralizadas em `internal/config/`

## Como Executar

1. Configure as variáveis de ambiente ou use os valores padrão
2. Execute o comando: `go run main.go`
3. A API estará disponível na porta 3000 (ou conforme configurado)

## Endpoints Disponíveis

### Autenticação
- `POST /api/v1/auth/login` - Login de usuário

### Usuários
- `GET /api/v1/users/` - Listar usuários
- `GET /api/v1/users/:id` - Buscar usuário por ID
- `POST /api/v1/users/` - Criar usuário
- `PUT /api/v1/users/:id` - Atualizar usuário
- `DELETE /api/v1/users/:id` - Deletar usuário

### Processos (Protegidos por JWT)
- `POST /api/v1/processes/` - Criar informação de processo
- `GET /api/v1/processes/` - Listar processos do usuário
- `GET /api/v1/processes/:id` - Buscar processo por ID
- `GET /api/v1/processes/process/:processId` - Buscar processos por Process ID
- `DELETE /api/v1/processes/:id` - Deletar processo

## Benefícios da Nova Estrutura

1. **Melhor Organização**: Código separado por responsabilidade
2. **Manutenibilidade**: Mais fácil de manter e expandir
3. **Clareza**: Nomes de arquivos mais descritivos
4. **Isolamento**: Arquivos gerados pelo SQLC separados do código manual
5. **Escalabilidade**: Estrutura preparada para crescimento do projeto