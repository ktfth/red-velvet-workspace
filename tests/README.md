# Testes de Carga - Banco Digital

Este diretório contém os testes de carga para o Banco Digital usando k6.

## Pré-requisitos

1. Instale o k6:
   - Windows (usando Chocolatey): `choco install k6`
   - Linux: 
     ```bash
     sudo gpg -k
     sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
     echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
     sudo apt-get update
     sudo apt-get install k6
     ```
   - macOS: `brew install k6`

2. Certifique-se que o Banco Digital está rodando:
   ```bash
   docker compose up -d
   ```

## Executando os Testes

Para executar os testes de carga:

```bash
k6 run load-test.js
```

## Cenários de Teste

O teste de carga inclui três cenários principais:

1. **Criação de Contas**
   - Rampa até 50 usuários virtuais em 1 minuto
   - Mantém 50 usuários por 3 minutos
   - Reduz para 0 em 1 minuto
   - Cria contas e realiza operações básicas

2. **Operações PIX**
   - Rampa até 30 usuários virtuais em 1 minuto
   - Mantém 30 usuários por 5 minutos
   - Reduz para 0 em 1 minuto
   - Registra chaves PIX, gera QR Codes e agenda transferências

3. **Operações com Cartão**
   - Rampa até 20 usuários virtuais em 1 minuto
   - Mantém 20 usuários por 4 minutos
   - Reduz para 0 em 1 minuto
   - Cria cartões, realiza compras e gera cartões virtuais

## Métricas e Thresholds

- 95% das requisições devem completar em menos de 2 segundos
- Taxa de falha deve ser menor que 10%

## Visualizando Resultados

O k6 fornecerá um relatório detalhado após a execução dos testes, incluindo:

- Tempo médio de resposta
- Taxa de requisições por segundo
- Taxa de erro
- Percentis de tempo de resposta
- Contadores de iterações por cenário 