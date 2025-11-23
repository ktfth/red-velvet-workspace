# Script para iniciar e testar o Banco Digital
# Requer Docker e Docker Compose instalados

Write-Host "Iniciando o ambiente do Banco Digital..." -ForegroundColor Green

# 1. Iniciar os serviços com Docker Compose
docker-compose up -d --build

if ($LASTEXITCODE -ne 0) {
    Write-Host "Erro ao iniciar os serviços." -ForegroundColor Red
    exit 1
}

Write-Host "Aguardando os serviços inicializarem (45 segundos)..." -ForegroundColor Yellow
Start-Sleep -Seconds 45

# 2. Verificar se a API está respondendo
Write-Host "Verificando status da API..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8081/status/ok" -Method Get
    Write-Host "API está online!" -ForegroundColor Green
}
catch {
    Write-Host "API não está respondendo. Verifique os logs do container." -ForegroundColor Red
    docker-compose logs banco-digital
    exit 1
}

# 3. Executar testes de carga (se k6 estiver instalado)
if (Get-Command k6 -ErrorAction SilentlyContinue) {
    Write-Host "Executando testes de carga com k6..." -ForegroundColor Cyan
    k6 run tests/k6/load-test.js
}
else {
    Write-Host "k6 não encontrado. Pulando testes de carga." -ForegroundColor Yellow
    Write-Host "Para executar os testes manualmente, instale o k6 e execute: k6 run tests/k6/load-test.js"
}

Write-Host "Ambiente pronto e testado!" -ForegroundColor Green
Write-Host "Frontend (Acme Inc.): http://localhost:3000"
Write-Host "API: http://localhost:8081"
Write-Host "Kafka UI: http://localhost:8090"
Write-Host "PgAdmin: http://localhost:5050"
