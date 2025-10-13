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
- `POST /api/v1/webhook/iterate-processes` - Itera todos os processos e cria um novo snapshot
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

### 1. Capturar todos os processos (Criar novo snapshot)

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
  "message": "Processes iterated and persisted successfully",
  "snapshot_id": 1,
  "process_count": 150,
  "processes": [...]
}
```

### 2. Consultar processo por PID (Criar novo snapshot)

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
  "message": "Process queried and persisted successfully",
  "snapshot_id": 2,
  "process_info_id": 151,
  "process": {...}
}
```

### 3. Consultar processo por PID (Adicionar a snapshot existente)

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

### 4. Listar todos os snapshots

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

### Cenário 1: Visualizar histórico de capturas
1. Chamar `GET /api/v1/processes/snapshots` para listar todos os snapshots
2. Exibir lista de snapshots com data, tipo e quantidade de processos
3. Ao clicar em um snapshot, chamar `GET /api/v1/processes/snapshots/:id/processes`
4. Exibir lista detalhada de processos daquele snapshot

### Cenário 2: Capturar novos processos
1. Usuário clica em "Capturar Processos"
2. Frontend chama `POST /api/v1/webhook/iterate-processes`
3. API cria novo snapshot e retorna `snapshot_id`
4. Frontend redireciona para visualização do novo snapshot

### Cenário 3: Consultar processo específico
1. Usuário digita um PID e clica em "Consultar"
2. Frontend pode:
   - Criar novo snapshot: `POST /api/v1/webhook/process-by-pid` (sem `snapshot_id`)
   - Adicionar a snapshot existente: `POST /api/v1/webhook/process-by-pid` (com `snapshot_id`)
3. Exibir detalhes do processo consultado

### Cenário 4: Comparar processos ao longo do tempo
1. Chamar `GET /api/v1/processes/pid/:pid` para obter todas as capturas de um PID específico
2. Exibir timeline com as diferentes capturas
3. Permitir comparação de métricas (memória, CPU, etc.)

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
