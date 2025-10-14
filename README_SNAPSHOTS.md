# Go API - Estrutura com Snapshots de Processos

## Visão Geral da Nova Estrutura

A API foi reorganizada para usar o conceito de **"Process Snapshots"** (capturas de processos), que representa uma sessão de captura de dados de processos em um momento específico.

## Conceitos Principais

### 1. Process Snapshot
Um snapshot é uma "sessão" que agrupa processos capturados em um momento específico. Existem dois tipos:
- **iteration**: Criado quando o endpoint `iterate-processes` é chamado
- **query**: Criado quando o endpoint `process-by-pid` é chamado (ou pode adicionar a um snapshot existente)

### 2. Process Info
Representa os dados detalhados de um processo específico. Cada processo está sempre vinculado a um snapshot.

### 3. Process Query
Registra o histórico de consultas individuais por PID, vinculadas a um snapshot.

## Estrutura do Banco de Dados

```
users
  └── process_snapshots (sessões de captura)
       ├── process_info (processos capturados)
       └── process_queries (histórico de consultas por PID)
```

## Endpoints da API

### Autenticação
- `POST /api/v1/auth/login` - Login de usuário

### Usuários
- `GET /api/v1/users` - Listar todos os usuários
- `GET /api/v1/users/:id` - Obter usuário específico
- `POST /api/v1/users` - Criar novo usuário
- `PUT /api/v1/users/:id` - Atualizar usuário
- `DELETE /api/v1/users/:id` - Deletar usuário

### Webhooks (Captura de Processos)

**Autenticação Opcional**: Estes endpoints funcionam com ou sem autenticação JWT.

- **Com autenticação** (Header: `Authorization: Bearer <token>`):
  - Cria snapshot vinculado ao usuário
  - Persiste todos os dados no banco de dados
  - Permite adicionar processos a snapshots existentes do usuário

- **Sem autenticação**:
  - Retorna os dados do webhook sem persistir
  - Útil para testes ou consultas rápidas
  - Não cria snapshots nem salva no banco

#### Endpoints:
- `POST /api/v1/webhook/iterate-processes` - Itera todos os processos
- `POST /api/v1/webhook/process-by-pid` - Consulta processo por PID

### Snapshots (Requer JWT)
- `GET /api/v1/processes/snapshots` - Listar todos os snapshots do usuário
- `GET /api/v1/processes/snapshots/type/:type` - Listar snapshots por tipo (iteration/query)
- `GET /api/v1/processes/snapshots/:id` - Obter snapshot específico
- `GET /api/v1/processes/snapshots/:id/processes` - Listar todos os processos de um snapshot
- `GET /api/v1/processes/snapshots/:id/queries` - Listar todas as consultas de um snapshot
- `DELETE /api/v1/processes/snapshots/:id` - Deletar snapshot (e todos os processos vinculados)

### Process Info (Requer JWT)
- `GET /api/v1/processes` - Listar todos os processos do usuário
- `GET /api/v1/processes/:id` - Obter processo específico
- `GET /api/v1/processes/pid/:pid` - Listar todos os processos com um PID específico (em diferentes snapshots)
- `DELETE /api/v1/processes/:id` - Deletar processo específico

### Histórico e Estatísticas (Requer JWT)
- `GET /api/v1/processes/queries/history` - Histórico de consultas por PID
- `GET /api/v1/processes/statistics` - Estatísticas do usuário

## Exemplos de Uso

### 1. Capturar todos os processos - SEM autenticação (não persiste)

```bash
POST /api/v1/webhook/iterate-processes
Content-Type: application/json

{
  "webhook_url": "http://localhost:8080/iterate-processes"
}
```

**Resposta:**
```json
{
  "message": "Processes iterated successfully (not persisted - no authentication)",
  "process_count": 150,
  "processes": [...]
}
```

### 2. Capturar todos os processos - COM autenticação (persiste em snapshot)

```bash
POST /api/v1/webhook/iterate-processes
Content-Type: application/json
Authorization: Bearer <token>

{
  "webhook_url": "http://localhost:8080/iterate-processes"
}
```

**Resposta:**
```json
{
  "message": "Processes iterated and persisted successfully",
  "snapshot_id": 1,
  "process_count": 150,
  "processes": [...]
}
```

### 3. Consultar processo por PID - SEM autenticação (não persiste)

```bash
POST /api/v1/webhook/process-by-pid
Content-Type: application/json

{
  "webhook_url": "http://localhost:8080/process-by-pid",
  "pid": 1234
}
```

**Resposta:**
```json
{
  "message": "Process queried successfully (not persisted - no authentication)",
  "process": {...}
}
```

### 4. Consultar processo por PID - COM autenticação (cria novo snapshot)

```bash
POST /api/v1/webhook/process-by-pid
Content-Type: application/json
Authorization: Bearer <token>

{
  "webhook_url": "http://localhost:8080/process-by-pid",
  "pid": 1234
}
```

**Resposta:**
```json
{
  "message": "Process queried and persisted successfully",
  "snapshot_id": 2,
  "process_info_id": 151,
  "process": {...}
}
```

### 5. Consultar processo por PID - Adicionar a snapshot existente

```bash
POST /api/v1/webhook/process-by-pid
Content-Type: application/json
Authorization: Bearer <token>

{
  "webhook_url": "http://localhost:8080/process-by-pid",
  "pid": 5678,
  "snapshot_id": 1
}
```

**Resposta:**
```json
{
  "message": "Process queried and persisted successfully",
  "snapshot_id": 1,
  "process_info_id": 152,
  "process": {...}
}
```

**Nota**: Ao adicionar a um snapshot existente, o sistema verifica se o snapshot pertence ao usuário autenticado. Se não pertencer, retorna erro 403 (Forbidden).

### 6. Listar todos os snapshots

```bash
GET /api/v1/processes/snapshots
Authorization: Bearer <token>
```

**Resposta:**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "webhook_url": "http://localhost:8080/iterate-processes",
    "snapshot_type": "iteration",
    "process_count": 150,
    "success": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  {
    "id": 2,
    "user_id": 1,
    "webhook_url": "http://localhost:8080/process-by-pid",
    "snapshot_type": "query",
    "process_count": 1,
    "success": true,
    "created_at": "2024-01-15T11:00:00Z",
    "updated_at": "2024-01-15T11:00:00Z"
  }
]
```

### 5. Ver todos os processos de um snapshot

```bash
GET /api/v1/processes/snapshots/1/processes
Authorization: Bearer <token>
```

**Resposta:**
```json
{
  "snapshot": {
    "id": 1,
    "snapshot_type": "iteration",
    "process_count": 150,
    ...
  },
  "processes": [
    {
      "id": 1,
      "snapshot_id": 1,
      "process_id": 4,
      "process_name": "System",
      "thread_count": 200,
      ...
    },
    ...
  ]
}
```

### 6. Ver estatísticas

```bash
GET /api/v1/processes/statistics
Authorization: Bearer <token>
```

**Resposta:**
```json
{
  "total_processes": 500,
  "total_snapshots": 10,
  "total_queries": 25,
  "most_queried_processes": [
    {
      "requested_pid": 1234,
      "query_count": 5
    },
    ...
  ],
  "snapshot_statistics": {
    "total_snapshots": 10,
    "total_processes": 500,
    "avg_processes_per_snapshot": 50
  }
}
```

## Fluxo de Trabalho no Frontend

### Cenário 1: Teste rápido sem autenticação
1. Usuário quer apenas visualizar processos sem salvar
2. Frontend chama `POST /api/v1/webhook/iterate-processes` sem token
3. API retorna os processos mas não persiste nada
4. Útil para demonstrações ou testes rápidos

### Cenário 2: Visualizar histórico de capturas (requer autenticação)
1. Usuário faz login e obtém token JWT
2. Frontend chama `GET /api/v1/processes/snapshots` com o token
3. Exibir lista de snapshots com data, tipo e quantidade de processos
4. Ao clicar em um snapshot, chamar `GET /api/v1/processes/snapshots/:id/processes`
5. Exibir lista detalhada de processos daquele snapshot

### Cenário 3: Capturar e salvar processos (requer autenticação)
1. Usuário autenticado clica em "Capturar Processos"
2. Frontend chama `POST /api/v1/webhook/iterate-processes` com token JWT
3. API cria novo snapshot vinculado ao usuário e retorna `snapshot_id`
4. Frontend redireciona para visualização do novo snapshot

### Cenário 4: Consultar processo específico
**Opção A - Sem autenticação (apenas visualizar):**
1. Usuário digita um PID e clica em "Consultar"
2. Frontend chama `POST /api/v1/webhook/process-by-pid` sem token
3. Exibir detalhes do processo (não salvo)

**Opção B - Com autenticação (salvar em novo snapshot):**
1. Usuário autenticado digita um PID
2. Frontend chama `POST /api/v1/webhook/process-by-pid` com token (sem `snapshot_id`)
3. API cria novo snapshot e salva o processo
4. Exibir detalhes e opção de ver o snapshot

**Opção C - Com autenticação (adicionar a snapshot existente):**
1. Usuário está visualizando um snapshot específico
2. Usuário digita um PID para adicionar ao snapshot atual
3. Frontend chama `POST /api/v1/webhook/process-by-pid` com token e `snapshot_id`
4. API adiciona o processo ao snapshot existente
5. Atualizar a lista de processos do snapshot

### Cenário 5: Comparar processos ao longo do tempo (requer autenticação)
1. Chamar `GET /api/v1/processes/pid/:pid` para obter todas as capturas de um PID específico
2. Exibir timeline com as diferentes capturas
3. Permitir comparação de métricas (memória, CPU, etc.)

### Cenário 6: Gerenciar snapshots (requer autenticação)
1. Listar snapshots por tipo: `GET /api/v1/processes/snapshots/type/iteration` ou `/type/query`
2. Ver detalhes de um snapshot: `GET /api/v1/processes/snapshots/:id`
3. Ver histórico de consultas de um snapshot: `GET /api/v1/processes/snapshots/:id/queries`
4. Deletar snapshot antigo: `DELETE /api/v1/processes/snapshots/:id`

## Migração de Dados Existentes

Se você já tem dados na estrutura antiga, execute o script de migração:

```bash
psql -U seu_usuario -d seu_banco -f migration_to_snapshots.sql
```

Este script irá:
1. Criar as novas tabelas
2. Migrar dados de `process_iteration_history` para `process_snapshots`
3. Migrar dados de `process_query_history` para `process_queries`
4. Vincular processos existentes aos snapshots apropriados
5. Criar índices para melhor performance

## Vantagens da Nova Estrutura

1. **Organização Clara**: Cada captura de processos é uma "sessão" bem definida
2. **Histórico Completo**: Fácil visualizar e comparar capturas ao longo do tempo
3. **Flexibilidade**: Pode adicionar processos a snapshots existentes ou criar novos
4. **Performance**: Índices otimizados para consultas comuns
5. **Cascata**: Deletar um snapshot remove automaticamente todos os processos vinculados
6. **Rastreabilidade**: Cada consulta por PID é registrada com seu snapshot
7. **Autenticação Opcional**: Endpoints de webhook funcionam com ou sem autenticação
8. **Segurança**: Snapshots são isolados por usuário, impedindo acesso não autorizado

## Segurança e Controle de Acesso

### Endpoints Públicos (Sem Autenticação)
- `POST /api/v1/auth/login` - Login
- `GET/POST/PUT/DELETE /api/v1/users/*` - Gerenciamento de usuários
- `POST /api/v1/webhook/*` - Webhooks (modo somente leitura, sem persistência)

### Endpoints Protegidos (Requer JWT)
- `GET/DELETE /api/v1/processes/*` - Todos os endpoints de processos e snapshots
- `POST /api/v1/webhook/*` - Webhooks (modo persistência, cria snapshots)

### Isolamento de Dados
- Cada snapshot pertence a um usuário específico
- Usuários só podem acessar seus próprios snapshots
- Ao adicionar processo a snapshot existente, verifica-se a propriedade
- Tentativa de acesso a snapshot de outro usuário retorna 403 (Forbidden)

### Comportamento dos Webhooks

| Situação | Comportamento |
|----------|---------------|
| Sem token JWT | Retorna dados do webhook sem persistir |
| Com token JWT válido | Cria snapshot vinculado ao usuário e persiste |
| Com token JWT inválido | Retorna dados sem persistir (não falha) |
| Adicionar a snapshot de outro usuário | Retorna 403 Forbidden |
| Adicionar a snapshot inexistente | Retorna 404 Not Found |

## Estrutura do Projeto

```
go-api/
├── main.go                    # Ponto de entrada e rotas
├── schema.sql                 # Schema do banco (nova estrutura)
├── queries.sql                # Queries SQL para o SQLC
├── migration_to_snapshots.sql # Script de migração
├── internal/
│   ├── config/
│   │   └── config.go          # Configurações
│   ├── db/                    # Arquivos gerados pelo SQLC
│   │   ├── db.go
│   │   ├── models.go
│   │   ├── querier.go
│   │   └── queries.sql.go
│   └── handlers/              # Handlers HTTP
│       ├── auth.go            # Autenticação
│       ├── middleware.go      # Middlewares
│       ├── user_handler.go    # CRUD de usuários
│       ├── webhook_handler.go # Captura de processos
│       └── process_handler.go # Gerenciamento de snapshots
└── docker-compose.yml         # Docker Compose
```

## Próximos Passos

1. Execute `sqlc generate` para gerar os arquivos Go a partir das queries
2. Execute o script de migração se tiver dados existentes
3. Teste os endpoints com Postman ou similar
4. Implemente o frontend consumindo os novos endpoints
