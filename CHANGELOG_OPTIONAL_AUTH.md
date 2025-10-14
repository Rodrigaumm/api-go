# Changelog - Autenticação Opcional em Webhooks

## Data: 2024

## Mudanças Implementadas

### 1. Novo Middleware: OptionalJWTMiddleware

**Arquivo:** `internal/handlers/middleware.go`

Criado um novo middleware que permite autenticação opcional:
- Se o token JWT estiver presente e válido, extrai as informações do usuário
- Se o token não estiver presente ou for inválido, permite a requisição continuar
- Não bloqueia requisições não autenticadas (diferente do JWTMiddleware)

### 2. Modificações no WebhookHandler

**Arquivo:** `internal/handlers/webhook_handler.go`

#### IterateProcesses
- Agora verifica se há usuário autenticado usando `GetUserFromContext`
- **Com autenticação**: Cria snapshot vinculado ao usuário e persiste todos os processos
- **Sem autenticação**: Retorna os dados do webhook sem persistir nada

#### ProcessByPid
- Verifica autenticação antes de persistir
- **Com autenticação**: 
  - Pode criar novo snapshot
  - Pode adicionar a snapshot existente (com verificação de propriedade)
- **Sem autenticação**: Retorna os dados sem persistir
- **Segurança**: Verifica se o snapshot pertence ao usuário antes de adicionar processo

### 3. Atualização das Rotas

**Arquivo:** `main.go`

Rotas de webhook agora usam `OptionalJWTMiddleware`:
```go
webhook := api.Group("/webhook")
webhook.Use(handlers.OptionalJWTMiddleware())
webhook.Post("/iterate-processes", webhookHandler.IterateProcesses)
webhook.Post("/process-by-pid", webhookHandler.ProcessByPid)
```

### 4. Documentação Atualizada

**Arquivo:** `README_SNAPSHOTS.md`

Adicionadas seções:
- Explicação sobre autenticação opcional nos webhooks
- Exemplos de uso com e sem autenticação
- Seção de segurança e controle de acesso
- Tabela de comportamento dos webhooks
- Fluxos de trabalho atualizados

### 5. Novo Arquivo de Exemplos

**Arquivo:** `EXAMPLES.md`

Criado arquivo completo com exemplos de curl para:
- Todos os endpoints da API
- Casos de uso com e sem autenticação
- Fluxo completo de exemplo
- Scripts prontos para copiar e colar

### 6. Correções de Código

- Substituído `interface{}` por `any` (Go 1.18+)
- Mescladas declarações de variáveis com atribuições (S1021)
- Removidos avisos do linter

## Benefícios

1. **Flexibilidade**: Webhooks podem ser usados com ou sem autenticação
2. **Testes Facilitados**: Desenvolvedores podem testar sem criar usuários
3. **Segurança Mantida**: Dados só são persistidos com autenticação válida
4. **Isolamento**: Snapshots são isolados por usuário
5. **Backward Compatible**: Não quebra funcionalidades existentes

## Comportamento dos Endpoints

| Endpoint | Sem Auth | Com Auth |
|----------|----------|----------|
| `/webhook/iterate-processes` | Retorna dados | Cria snapshot + persiste |
| `/webhook/process-by-pid` | Retorna dados | Cria/adiciona a snapshot |
| `/processes/*` | ❌ Bloqueado | ✅ Permitido |

## Segurança

- Snapshots são isolados por usuário
- Tentativa de acessar snapshot de outro usuário retorna 403
- Token inválido não causa erro, apenas não persiste
- Verificação de propriedade ao adicionar a snapshot existente

## Próximos Passos Sugeridos

1. Adicionar rate limiting nos endpoints de webhook
2. Implementar cache para consultas frequentes
3. Adicionar paginação nos endpoints de listagem
4. Criar testes automatizados para os novos fluxos
5. Adicionar logs estruturados para auditoria
