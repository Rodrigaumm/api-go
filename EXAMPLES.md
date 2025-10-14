# Exemplos de Uso da API - Process Snapshots

Este arquivo contém exemplos práticos de como usar a API com curl.

## 1. Criar um usuário

```bash
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "usuario_teste",
    "password": "senha123"
  }'
```

## 2. Fazer login

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "name": "usuario_teste",
    "password": "senha123"
  }'
```

**Resposta:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "name": "usuario_teste"
  }
}
```

**Salve o token para usar nos próximos comandos!**

## 3. Testar webhook SEM autenticação (não persiste)

```bash
curl -X POST http://localhost:3000/api/v1/webhook/iterate-processes \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_url": "http://seu-webhook-url/iterate-processes"
  }'
```

**Resposta:**
```json
{
  "message": "Processes iterated successfully (not persisted - no authentication)",
  "process_count": 150,
  "processes": [...]
}
```

## 4. Capturar processos COM autenticação (persiste em snapshot)

```bash
TOKEN="seu_token_aqui"

curl -X POST http://localhost:3000/api/v1/webhook/iterate-processes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "webhook_url": "http://seu-webhook-url/iterate-processes"
  }'
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

## 5. Consultar processo por PID SEM autenticação

```bash
curl -X POST http://localhost:3000/api/v1/webhook/process-by-pid \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_url": "http://seu-webhook-url/process-by-pid",
    "pid": 1234
  }'
```

## 6. Consultar processo por PID COM autenticação (novo snapshot)

```bash
TOKEN="seu_token_aqui"

curl -X POST http://localhost:3000/api/v1/webhook/process-by-pid \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "webhook_url": "http://seu-webhook-url/process-by-pid",
    "pid": 1234
  }'
```

## 7. Adicionar processo a snapshot existente

```bash
TOKEN="seu_token_aqui"
SNAPSHOT_ID=1

curl -X POST http://localhost:3000/api/v1/webhook/process-by-pid \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "webhook_url": "http://seu-webhook-url/process-by-pid",
    "pid": 5678,
    "snapshot_id": '$SNAPSHOT_ID'
  }'
```

## 8. Listar todos os snapshots do usuário

```bash
TOKEN="seu_token_aqui"

curl -X GET http://localhost:3000/api/v1/processes/snapshots \
  -H "Authorization: Bearer $TOKEN"
```

## 9. Listar snapshots por tipo

```bash
TOKEN="seu_token_aqui"

# Listar apenas snapshots de iteração
curl -X GET http://localhost:3000/api/v1/processes/snapshots/type/iteration \
  -H "Authorization: Bearer $TOKEN"

# Listar apenas snapshots de consulta
curl -X GET http://localhost:3000/api/v1/processes/snapshots/type/query \
  -H "Authorization: Bearer $TOKEN"
```

## 10. Ver detalhes de um snapshot específico

```bash
TOKEN="seu_token_aqui"
SNAPSHOT_ID=1

curl -X GET http://localhost:3000/api/v1/processes/snapshots/$SNAPSHOT_ID \
  -H "Authorization: Bearer $TOKEN"
```

## 11. Ver todos os processos de um snapshot

```bash
TOKEN="seu_token_aqui"
SNAPSHOT_ID=1

curl -X GET http://localhost:3000/api/v1/processes/snapshots/$SNAPSHOT_ID/processes \
  -H "Authorization: Bearer $TOKEN"
```

## 12. Ver histórico de consultas de um snapshot

```bash
TOKEN="seu_token_aqui"
SNAPSHOT_ID=1

curl -X GET http://localhost:3000/api/v1/processes/snapshots/$SNAPSHOT_ID/queries \
  -H "Authorization: Bearer $TOKEN"
```

## 13. Listar todos os processos do usuário

```bash
TOKEN="seu_token_aqui"

curl -X GET http://localhost:3000/api/v1/processes \
  -H "Authorization: Bearer $TOKEN"
```

## 14. Ver detalhes de um processo específico

```bash
TOKEN="seu_token_aqui"
PROCESS_ID=1

curl -X GET http://localhost:3000/api/v1/processes/$PROCESS_ID \
  -H "Authorization: Bearer $TOKEN"
```

## 15. Ver histórico de um PID específico (todas as capturas)

```bash
TOKEN="seu_token_aqui"
PID=1234

curl -X GET http://localhost:3000/api/v1/processes/pid/$PID \
  -H "Authorization: Bearer $TOKEN"
```

## 16. Ver histórico de consultas por PID

```bash
TOKEN="seu_token_aqui"

curl -X GET http://localhost:3000/api/v1/processes/queries/history \
  -H "Authorization: Bearer $TOKEN"
```

## 17. Ver estatísticas do usuário

```bash
TOKEN="seu_token_aqui"

curl -X GET http://localhost:3000/api/v1/processes/statistics \
  -H "Authorization: Bearer $TOKEN"
```

## 18. Deletar um processo específico

```bash
TOKEN="seu_token_aqui"
PROCESS_ID=1

curl -X DELETE http://localhost:3000/api/v1/processes/$PROCESS_ID \
  -H "Authorization: Bearer $TOKEN"
```

## 19. Deletar um snapshot (e todos os processos vinculados)

```bash
TOKEN="seu_token_aqui"
SNAPSHOT_ID=1

curl -X DELETE http://localhost:3000/api/v1/processes/snapshots/$SNAPSHOT_ID \
  -H "Authorization: Bearer $TOKEN"
```

## 20. Verificar saúde da API

```bash
curl -X GET http://localhost:3000/health
```

---

## Fluxo Completo de Exemplo

```bash
# 1. Criar usuário
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name": "teste", "password": "123456"}'

# 2. Fazer login e salvar token
TOKEN=$(curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"name": "teste", "password": "123456"}' \
  | jq -r '.token')

echo "Token: $TOKEN"

# 3. Capturar processos (cria snapshot)
SNAPSHOT_ID=$(curl -X POST http://localhost:3000/api/v1/webhook/iterate-processes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"webhook_url": "http://seu-webhook/iterate-processes"}' \
  | jq -r '.snapshot_id')

echo "Snapshot ID: $SNAPSHOT_ID"

# 4. Adicionar processo específico ao snapshot
curl -X POST http://localhost:3000/api/v1/webhook/process-by-pid \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "webhook_url": "http://seu-webhook/process-by-pid",
    "pid": 1234,
    "snapshot_id": '$SNAPSHOT_ID'
  }'

# 5. Ver todos os processos do snapshot
curl -X GET http://localhost:3000/api/v1/processes/snapshots/$SNAPSHOT_ID/processes \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'

# 6. Ver estatísticas
curl -X GET http://localhost:3000/api/v1/processes/statistics \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'
```

---

## Notas Importantes

1. **Substitua `http://localhost:3000`** pela URL da sua API
2. **Substitua `http://seu-webhook-url`** pela URL do seu webhook real
3. **Use `jq`** para formatar JSON (opcional, mas recomendado)
4. **Salve o token** após o login para usar em requisições autenticadas
5. **Snapshots são isolados por usuário** - você só pode acessar seus próprios snapshots
6. **Deletar um snapshot** remove automaticamente todos os processos vinculados
