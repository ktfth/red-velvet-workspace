# 🏦 Guia de Uso - Banco Digital "Acme Inc."

## 🚀 Como Iniciar o Sistema

Execute o script automatizado:
```powershell
.\run-and-test.ps1
```

## 🌐 URLs Disponíveis

- **Frontend (Acme Inc.)**: http://localhost:3000
- **API (via Kong)**: http://localhost:8000/api
- **API Direta**: http://localhost:8081
- **Kafka UI**: http://localhost:8090
- **PgAdmin**: http://localhost:5050
- **Kong Admin**: http://localhost:8001

## 📱 Como Usar o Frontend

### 1. Criar uma Conta (via API)

Primeiro, você precisa criar uma conta usando a API:

```powershell
$body = @{tipo='corrente'} | ConvertTo-Json
curl -X POST http://localhost:8000/api/conta/criar `
  -H 'Content-Type: application/json' `
  -H 'Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiIzN1U0VmpGQVhBSXRxZnBwSXJaY3NmaE1CbTVoRm94TCIsImV4cCI6MTc5NTQ1MDc3MH0.cwoty-5iPTplC5DpdcXxabx66ZC5id40qEeNdIjVXvs' `
  -d $body
```

**Resposta**:
```json
{
  "success": true,
  "data": {
    "id": "3f800934-ff0b-4613-a5e7-5d3547132b61",
    "type": "corrente",
    "number": "1763915108",
    "balance": 0
  }
}
```

**💡 Copie o `id` retornado!**

### 2. Acessar o Frontend

1. Abra o navegador em: http://localhost:3000
2. Na tela de login, cole o **Account ID** que você copiou
3. Clique em "Entrar"

### 3. Realizar Transações

No dashboard, você pode:
- ✅ **Fazer Depósitos**: Selecione "Depósito", insira o valor e confirme
- ✅ **Fazer Saques**: Selecione "Saque", insira o valor e confirme
- 📊 O saldo será atualizado localmente (simulado)

## 🔐 Autenticação JWT

O sistema usa **Kong API Gateway** com autenticação JWT.

### Token JWT Válido
```
Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiIzN1U0VmpGQVhBSXRxZnBwSXJaY3NmaE1CbTVoRm94TCIsImV4cCI6MTc5NTQ1MDc3MH0.cwoty-5iPTplC5DpdcXxabx66ZC5id40qEeNdIjVXvs
```

**Validade**: 1 ano (até 2026)

### Gerar Novo Token

Se precisar gerar um novo token JWT:
```powershell
node scripts\generate-jwt.js
```

## 📝 Exemplo de Account ID Criado

Para testes rápidos, use este Account ID já criado:
```
3f800934-ff0b-4613-a5e7-5d3547132b61
```

## 🔧 Testando a API Diretamente

### Criar Conta
```powershell
$body = @{tipo='corrente'} | ConvertTo-Json
curl -X POST http://localhost:8000/api/conta/criar `
  -H 'Content-Type: application/json' `
  -H 'Authorization: Bearer eyJ...' `
  -d $body
```

### Fazer Depósito
```powershell
$body = @{conta_id='3f800934-ff0b-4613-a5e7-5d3547132b61'; valor=100.50} | ConvertTo-Json
curl -X POST http://localhost:8000/api/conta/depositar `
  -H 'Content-Type: application/json' `
  -H 'Authorization: Bearer eyJ...' `
  -d $body
```

### Fazer Saque
```powershell
$body = @{conta_id='3f800934-ff0b-4613-a5e7-5d3547132b61'; valor=50.00} | ConvertTo-Json
curl -X POST http://localhost:8000/api/conta/sacar `
  -H 'Content-Type: application/json' `
  -H 'Authorization: Bearer eyJ...' `
  -d $body
```

## 🛠️ Troubleshooting

### Frontend não carrega
```powershell
docker-compose restart frontend
```

### API não responde
```powershell
docker logs red-velvet-workspace-banco-digital-1
```

### Kafka com problemas
```powershell
docker logs red-velvet-workspace-kafka-1
```

### Resetar tudo
```powershell
docker-compose down -v
.\run-and-test.ps1
```

## ✨ Funcionalidades Implementadas

- ✅ Criação de contas
- ✅ Depósitos
- ✅ Saques
- ✅ Autenticação JWT via Kong
- ✅ Frontend Vue.js responsivo
- ✅ Event-driven com Kafka
- ✅ Persistência PostgreSQL
- ✅ CORS configurado
- ✅ API Gateway (Kong)

## 🎯 Próximos Passos

Para expandir o sistema, você pode:
1. Adicionar endpoint para consultar saldo real
2. Implementar histórico de transações
3. Adicionar funcionalidades de cartão de crédito
4. Implementar transferências PIX
5. Criar sistema de notificações em tempo real

---

**🎉 Divirta-se explorando o Banco Digital "Acme Inc."!**
