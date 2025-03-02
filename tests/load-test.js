import { check, sleep } from 'k6';

import http from 'k6/http';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Variáveis globais para armazenar IDs
let contasIds = new Set();
let chavesPixIds = new Set();
let cartoesIds = new Set();

// Configuração dos cenários de teste
export const options = {
  scenarios: {
    // Teste de criação de contas
    criar_contas: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 50 },  // Aumenta para 50 VUs em 1 minuto
        { duration: '3m', target: 50 },  // Mantém 50 VUs por 3 minutos
        { duration: '1m', target: 0 },   // Reduz para 0 VUs em 1 minuto
      ],
      exec: 'criarContas',
    },
    
    // Teste de operações PIX
    operacoes_pix: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 30 },  // Aumenta para 30 VUs em 1 minuto
        { duration: '5m', target: 30 },  // Mantém 30 VUs por 5 minutos
        { duration: '1m', target: 0 },   // Reduz para 0 VUs em 1 minuto
      ],
      exec: 'operacoesPix',
    },
    
    // Teste de operações com cartão
    operacoes_cartao: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 20 },  // Aumenta para 20 VUs em 1 minuto
        { duration: '4m', target: 20 },  // Mantém 20 VUs por 4 minutos
        { duration: '1m', target: 0 },   // Reduz para 0 VUs em 1 minuto
      ],
      exec: 'operacoesCartao',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% das requisições devem completar em 2s
    http_req_failed: ['rate<0.1'],     // Taxa de falha menor que 10%
  },
};

// Função auxiliar para extrair ID da resposta
function extrairId(resposta) {
  try {
    return resposta.split('ID: ')[1].trim();
  } catch (e) {
    console.error('Erro ao extrair ID:', e);
    return null;
  }
}

// Cenário: Criação de Contas
export function criarContas() {
  const titular = `Titular ${randomString(8)}`;
  
  const res = http.post('http://localhost:8080/conta/criar', {
    titular: titular,
  });
  
  check(res, {
    'criação de conta bem sucedida': (r) => r.status === 200,
  });
  
  if (res.status === 200) {
    const contaId = extrairId(res.body);
    if (contaId) {
      contasIds.add(contaId);
      
      // Realiza operações básicas na conta
      const depositoRes = http.post('http://localhost:8080/conta/depositar', {
        conta_id: contaId,
        valor: '1000',
        categoria: 'Depósito inicial',
        descricao: 'Teste de carga',
      });
      
      check(depositoRes, {
        'depósito realizado com sucesso': (r) => r.status === 200,
      });
      
      const chequeRes = http.post('http://localhost:8080/conta/cheque-especial', {
        conta_id: contaId,
        limite: '500',
      });
      
      check(chequeRes, {
        'cheque especial configurado com sucesso': (r) => r.status === 200,
      });
    }
  }
  
  sleep(1);
}

// Cenário: Operações PIX
export function operacoesPix() {
  if (contasIds.size < 2) {
    console.log('Aguardando criação de contas para operações PIX...');
    sleep(1);
    return;
  }
  
  const contasArray = Array.from(contasIds);
  const contaOrigem = contasArray[Math.floor(Math.random() * contasArray.length)];
  
  // Registra chave PIX
  const chaveRes = http.post('http://localhost:8080/pix/registrar', {
    conta_id: contaOrigem,
    tipo_chave: 'email',
    valor_chave: `teste${randomString(8)}@teste.com`,
  });
  
  check(chaveRes, {
    'registro de chave PIX bem sucedido': (r) => r.status === 200,
  });
  
  if (chaveRes.status === 200) {
    const chaveId = extrairId(chaveRes.body);
    if (chaveId) {
      chavesPixIds.add(chaveId);
      
      // Gera QR Code
      const qrRes = http.post('http://localhost:8080/pix/qrcode', {
        chave_id: chaveId,
        tipo: 'estatico',
        valor: '100',
        descricao: 'Teste de carga QR Code',
      });
      
      check(qrRes, {
        'geração de QR Code bem sucedida': (r) => r.status === 200,
      });
      
      // Agenda transferência
      if (chavesPixIds.size >= 2) {
        const chavesArray = Array.from(chavesPixIds);
        const chaveDestino = chavesArray[Math.floor(Math.random() * chavesArray.length)];
        
        const agendamentoRes = http.post('http://localhost:8080/pix/agendar', {
          chave_origem: chaveId,
          chave_destino: chaveDestino,
          valor: '50',
          data: new Date(Date.now() + 86400000).toISOString(), // Amanhã
        });
        
        check(agendamentoRes, {
          'agendamento PIX bem sucedido': (r) => r.status === 200,
        });
      }
    }
  }
  
  sleep(1);
}

// Cenário: Operações com Cartão
export function operacoesCartao() {
  if (contasIds.size === 0) {
    console.log('Aguardando criação de contas para operações com cartão...');
    sleep(1);
    return;
  }
  
  const contasArray = Array.from(contasIds);
  const contaId = contasArray[Math.floor(Math.random() * contasArray.length)];
  
  // Cria cartão
  const cartaoRes = http.post('http://localhost:8080/cartao/criar', {
    conta_id: contaId,
  });
  
  check(cartaoRes, {
    'criação de cartão bem sucedida': (r) => r.status === 200,
  });
  
  if (cartaoRes.status === 200) {
    const cartaoId = extrairId(cartaoRes.body);
    if (cartaoId) {
      cartoesIds.add(cartaoId);
      
      // Realiza compras
      const compraRes = http.post('http://localhost:8080/cartao/comprar', {
        cartao_id: cartaoId,
        valor: '200',
        estabelecimento: 'Loja Teste',
        parcelas: '3',
      });
      
      check(compraRes, {
        'compra com cartão bem sucedida': (r) => r.status === 200,
      });
      
      // Gera cartão virtual
      const virtualRes = http.post('http://localhost:8080/cartao/virtual', {
        cartao_id: cartaoId,
      });
      
      check(virtualRes, {
        'geração de cartão virtual bem sucedida': (r) => r.status === 200,
      });
      
      // Paga fatura
      const faturaRes = http.post('http://localhost:8080/cartao/pagar', {
        cartao_id: cartaoId,
        valor: '100',
      });
      
      check(faturaRes, {
        'pagamento de fatura bem sucedido': (r) => r.status === 200,
      });
    }
  }
  
  sleep(1);
}

// Função principal que será executada por padrão
export default function() {
  criarContas();
  operacoesPix();
  operacoesCartao();
} 