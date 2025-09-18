# Exemplos de uso dos endpoints webhook

## 1. Endpoint para iterar processos

**GET** `/api/v1/webhook/iterate-processes?webhookurl=http://localhost:8888`

Este endpoint faz uma requisição para `{webhookurl}/webhook/iterate-processes` e retorna a lista de processos.

### Exemplo de requisição:
```bash
curl -X GET "http://localhost:3000/api/v1/webhook/iterate-processes?webhookurl=http://localhost:8888"
```

### Exemplo de resposta:
```json
{
  "data": {
    "success": true,
    "processCount": 150,
    "processes": [
      {
        "index": 0,
        "processName": "System",
        "processId": 4
      },
      {
        "index": 1,
        "processName": "Registry",
        "processId": 72
      }
    ]
  }
}
```

## 2. Endpoint para obter processo por PID

**POST** `/api/v1/webhook/process-by-pid?webhookurl=http://localhost:8888`

Este endpoint faz uma requisição para `{webhookurl}/webhook/process-by-pid` com o PID especificado no body.

### Exemplo de requisição:
```bash
curl -X POST "http://localhost:3000/api/v1/webhook/process-by-pid?webhookurl=http://localhost:8888" \
  -H "Content-Type: application/json" \
  -d '{"pid": 1234}'
```

### Exemplo de resposta:
```json
{
  "data": {
    "success": true,
    "processInfo": {
      "processId": 1234,
      "parentProcessId": 456,
      "processName": "notepad.exe",
      "threadCount": 2,
      "handleCount": 45,
      "basePriority": 8,
      "createTime": "2024-01-15T10:30:00Z",
      "workingSetSize": "2.5 MB",
      "virtualSize": "15.2 MB"
    }
  }
}
```

## Tratamento de Erros

### Parâmetro webhookurl ausente:
```json
{
  "error": "Parâmetro 'webhookurl' é obrigatório"
}
```

### PID inválido ou ausente:
```json
{
  "error": "PID deve ser um número positivo"
}
```

### Erro na requisição externa:
```json
{
  "error": "Erro ao fazer requisição para http://localhost:8888/webhook/iterate-processes: connection refused"
}
```