# Go API - Reestruturação Completa

## Resumo das Alterações

Este projeto foi completamente reestruturado para suportar a persistência de dados de processos do Windows com base na estrutura `ProcessInfo` do webhook_handler.go.

## Principais Mudanças

### 1. Schema do Banco de Dados (schema.sql)

Foram criadas as seguintes tabelas:

#### `process_info`
- Armazena informações completas de processos do Windows
- Inclui todos os campos do struct `ProcessInfo` do webhook
- Suporta processos adjacentes (next/previous) inline
- Campo `user_id` é **NULLABLE** - permite armazenar dados sem autenticação
- Campos de tempo e memória armazenados como TEXT para compatibilidade com formato do webhook

#### `process_iteration_history`
- Rastreia cada chamada ao endpoint `iterate-processes`
- Armazena: webhook URL, contagem de processos, sucesso/erro
- Campo `user_id` é **NULLABLE**

#### `iteration_processes`
- Tabela de ligação many-to-many entre iterações e processos
- Permite rastrear quais processos foram retornados em cada iteração

#### `process_query_history`
- Rastreia cada chamada ao endpoint `processByPid`
- Armazena: webhook URL, PID solicitado, sucesso/erro
- Referência ao processo retornado
- Campo `user_id` é **NULLABLE**

### 2. Queries SQL (queries.sql)

Foram criadas queries para:

**Process Info:**
- `CreateProcessInfo` - Criar novo registro de processo
- `GetProcessInfosByUser` - Buscar todos os processos de um usuário
- `GetProcessInfo` - Buscar processo específico por ID
- `GetProcessInfoByID` - Buscar processo por ID (sem filtro de usuário)
- `GetProcessInfosByProcessID` - Buscar histórico de um PID específico
- `DeleteProcessInfo` - Deletar registro de processo

**Iteration History:**
- `CreateProcessIterationHistory` - Criar registro de iteração
- `GetProcessIterationHistoryByUser` - Buscar histórico de iterações do usuário
- `GetProcessIterationHistory` - Buscar iteração específica
- `GetProcessesByIterationID` - Buscar todos os processos de uma iteração

**Iteration Processes (Link):**
- `CreateIterationProcess` - Vincular processo a iteração
- `GetIterationsByProcessInfoID` - Buscar iterações que contêm um processo

**Query History:**
- `CreateProcessQueryHistory` - Criar registro de query
- `GetProcessQueryHistoryByUser` - Buscar histórico de queries do usuário
- `GetProcessQueryHistory` - Buscar query específica
- `GetProcessQueryHistoryByPID` - Buscar histórico de queries para um PID

**Estatísticas:**
- `GetUserProcessCount` - Contar processos do usuário
- `GetUserIterationCount` - Contar iterações do usuário
- `GetUserQueryCount` - Contar queries do usuário
- `GetMostQueriedProcesses` - Processos mais consultados

### 3. Webhook Handler (webhook_handler.go)

**Lógica de Persistência Condicional:**

#### `IterateProcesses`
1. Faz requisição HTTP para o webhook externo
2. **SE** houver JWT válido no contexto:
   - Cria registro em `process_iteration_history`
   - Para cada processo retornado:
     - Persiste em `process_info`
     - Vincula à iteração em `iteration_processes`
3. **SE NÃO** houver JWT: apenas retorna os dados sem persistir
4. Em caso de erro, registra a falha no histórico (se autenticado)

#### `ProcessByPid`
1. Faz requisição HTTP para o webhook externo
2. **SE** houver JWT válido no contexto:
   - Persiste o processo em `process_info`
   - Cria registro em `process_query_history`
3. **SE NÃO** houver JWT: apenas retorna os dados sem persistir
4. Em caso de erro, registra a falha no histórico (se autenticado)

**Função Helper:**
- `persistProcessInfo` - Converte `ProcessInfo` do webhook para formato do banco

### 4. Process Handler (process_handler.go)

**Novos Endpoints:**

#### Gestão de Processos (existentes, atualizados)
- `CreateProcessInfo` - Criar processo manualmente
- `GetProcessInfos` - Listar todos os processos do usuário
- `GetProcessInfo` - Buscar processo específico
- `GetProcessInfosByProcessID` - Histórico de um PID
- `DeleteProcessInfo` - Deletar processo

#### Histórico de Iterações (novos)
- `GetIterationHistory` - Listar todas as iterações do usuário
- `GetIterationProcesses` - Ver processos de uma iteração específica

#### Histórico de Queries (novos)
- `GetQueryHistory` - Listar todas as queries do usuário
  - Inclui informações do processo retornado

#### Estatísticas (novo)
- `GetStatistics` - Estatísticas do usuário:
  - Total de processos armazenados
  - Total de iterações realizadas
  - Total de queries realizadas
  - Processos mais consultados

**Estruturas de Resposta:**
- `ProcessInfoResponse` - Inclui processos adjacentes (next/previous)
- `IterationHistoryResponse` - Informações de iteração
- `QueryHistoryResponse` - Informações de query com processo opcional

## Próximos Passos

### 1. Gerar Código SQLC

Execute o comando para gerar os arquivos Go baseados nas queries:

```bash
sqlc generate
```

Isso irá gerar/atualizar os arquivos em `internal/db/`:
- `db.go` - Interface do banco
- `models.go` - Structs dos modelos
- `querier.go` - Interface das queries
- `queries.sql.go` - Implementação das queries

### 2. Atualizar main.go

Certifique-se de que o `main.go` está passando o `dbpool` para o `NewWebhookHandler`:

```go
webhookHandler := handlers.NewWebhookHandler(dbpool)
```

### 3. Executar Migrations

Execute o schema.sql no banco de dados PostgreSQL:

```bash
psql -U seu_usuario -d seu_banco -f schema.sql
```

Ou use uma ferramenta de migration como golang-migrate.

### 4. Configurar Rotas

Adicione as novas rotas no `main.go`:

```go
// Rotas de histórico (requerem autenticação)
api.Get("/process-history/iterations", processHandler.GetIterationHistory)
api.Get("/process-history/iterations/:iterationId/processes", processHandler.GetIterationProcesses)
api.Get("/process-history/queries", processHandler.GetQueryHistory)
api.Get("/process-history/statistics", processHandler.GetStatistics)
```

### 5. Testar

#### Sem Autenticação (não persiste)
```bash
# Iterate processes
curl -X POST "http://localhost:3000/api/webhook/iterate-processes?webhookurl=http://webhook-url"

# Process by PID
curl -X POST "http://localhost:3000/api/webhook/process-by-pid?webhookurl=http://webhook-url" \
  -H "Content-Type: application/json" \
  -d '{"pid": 1234}'
```

#### Com Autenticação (persiste)
```bash
# Login
TOKEN=$(curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"name":"usuario","password":"senha"}' | jq -r '.token')

# Iterate processes (com persistência)
curl -X POST "http://localhost:3000/api/webhook/iterate-processes?webhookurl=http://webhook-url" \
  -H "Authorization: Bearer $TOKEN"

# Ver histórico de iterações
curl http://localhost:3000/api/process-history/iterations \
  -H "Authorization: Bearer $TOKEN"

# Ver processos de uma iteração
curl http://localhost:3000/api/process-history/iterations/1/processes \
  -H "Authorization: Bearer $TOKEN"

# Ver histórico de queries
curl http://localhost:3000/api/process-history/queries \
  -H "Authorization: Bearer $TOKEN"

# Ver estatísticas
curl http://localhost:3000/api/process-history/statistics \
  -H "Authorization: Bearer $TOKEN"
```

## Arquitetura da Solução

### Fluxo de Dados

```
Cliente → API → Webhook Externo → Resposta
                    ↓
              JWT Válido?
                    ↓
              Sim → Persistir no DB
                    ↓
              Não → Apenas retornar dados
```

### Vantagens da Abordagem

1. **Flexibilidade**: Funciona com ou sem autenticação
2. **Rastreabilidade**: Todo histórico é mantido quando autenticado
3. **Análise**: Estatísticas sobre uso e processos
4. **Auditoria**: Sabe-se quando e quem consultou cada processo
5. **Performance**: Índices otimizados para queries comuns

### Estrutura de Dados

```
User (1) ─────┐
              ├──→ ProcessInfo (N)
              │
              ├──→ ProcessIterationHistory (N)
              │         │
              │         └──→ IterationProcesses ──→ ProcessInfo
              │
              └──→ ProcessQueryHistory (N)
                        │
                        └──→ ProcessInfo (1)
```

## Considerações Importantes

1. **Campos TEXT vs BIGINT**: Alguns campos foram mantidos como TEXT (tempos, tamanhos de memória) para compatibilidade direta com o formato do webhook

2. **User ID Nullable**: Permite armazenar dados mesmo sem autenticação (se necessário no futuro)

3. **Processos Adjacentes**: Armazenados inline na tabela principal para simplificar queries

4. **Unique Constraint**: `(process_id, current_process_address, created_at)` previne duplicatas exatas

5. **Cascade Delete**: Deletar um usuário remove todos os seus dados relacionados

## Troubleshooting

### SQLC não encontrado
```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### Erros de compilação após gerar SQLC
- Verifique se os tipos em `queries.sql` correspondem aos do `schema.sql`
- Execute `go mod tidy` para atualizar dependências

### Banco de dados não conecta
- Verifique as variáveis de ambiente no `.env`
- Confirme que o PostgreSQL está rodando
- Teste a conexão com `psql`
