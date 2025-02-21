package conta

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Conta struct {
	producer *kafka.Writer
}

type ContaModel struct {
	ID             string    `json:"id"`
	Titular        string    `json:"titular"`
	Saldo          float64   `json:"saldo"`
	Status         string    `json:"status"` // ativa, bloqueada
	DataCriacao    time.Time `json:"data_criacao"`
	ChequeEspecial float64   `json:"cheque_especial"`
	LimiteCheque   float64   `json:"limite_cheque"`
	UltimoAcesso   time.Time `json:"ultimo_acesso"`
	Notificacoes   bool      `json:"notificacoes"`
}

type Transacao struct {
	ID        string    `json:"id"`
	ContaID   string    `json:"conta_id"`
	Tipo      string    `json:"tipo"` // deposito, saque, transferencia
	Valor     float64   `json:"valor"`
	Data      time.Time `json:"data"`
	Categoria string    `json:"categoria"`
	Descricao string    `json:"descricao"`
}

type Agendamento struct {
	ID            string    `json:"id"`
	ContaID       string    `json:"conta_id"`
	Valor         float64   `json:"valor"`
	DataPagamento time.Time `json:"data_pagamento"`
	Beneficiario  string    `json:"beneficiario"`
	Status        string    `json:"status"` // pendente, executado, cancelado
}

func (c *Conta) Init(ctx context.Context) error {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "kafka:9092"
	}

	c.producer = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokers},
		Topic:   "transacoes",
	})
	return nil
}

func (c *Conta) Criar(ctx context.Context, titular string, saldoInicial float64) (string, error) {
	conta := ContaModel{
		ID:             uuid.New().String(),
		Titular:        titular,
		Saldo:          saldoInicial,
		Status:         "ativa",
		DataCriacao:    time.Now(),
		ChequeEspecial: 0,
		LimiteCheque:   500, // Limite padrão de cheque especial
		UltimoAcesso:   time.Now(),
		Notificacoes:   true,
	}

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":  "conta_criada",
		"conta": conta,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	err = c.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao publicar mensagem: %v", err)
	}

	return conta.ID, nil
}

func (c *Conta) AlterarStatus(ctx context.Context, contaID string, novoStatus string) error {
	if novoStatus != "ativa" && novoStatus != "bloqueada" {
		return fmt.Errorf("status inválido: %s", novoStatus)
	}

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":        "status_alterado",
		"conta_id":    contaID,
		"novo_status": novoStatus,
	})
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	return c.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
}

func (c *Conta) Depositar(ctx context.Context, contaID string, valor float64, categoria string, descricao string) (string, error) {
	transacao := Transacao{
		ID:        uuid.New().String(),
		ContaID:   contaID,
		Tipo:      "deposito",
		Valor:     valor,
		Data:      time.Now(),
		Categoria: categoria,
		Descricao: descricao,
	}

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":      "deposito",
		"transacao": transacao,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	err = c.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao publicar mensagem: %v", err)
	}

	return transacao.ID, nil
}

func (c *Conta) Sacar(ctx context.Context, contaID string, valor float64, categoria string, descricao string) (string, error) {
	transacao := Transacao{
		ID:        uuid.New().String(),
		ContaID:   contaID,
		Tipo:      "saque",
		Valor:     valor,
		Data:      time.Now(),
		Categoria: categoria,
		Descricao: descricao,
	}

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":      "saque",
		"transacao": transacao,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	err = c.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao publicar mensagem: %v", err)
	}

	return transacao.ID, nil
}

func (c *Conta) AgendarPagamento(ctx context.Context, contaID string, valor float64, dataPagamento time.Time, beneficiario string) (string, error) {
	agendamento := Agendamento{
		ID:            uuid.New().String(),
		ContaID:       contaID,
		Valor:         valor,
		DataPagamento: dataPagamento,
		Beneficiario:  beneficiario,
		Status:        "pendente",
	}

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":        "pagamento_agendado",
		"agendamento": agendamento,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	err = c.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao publicar mensagem: %v", err)
	}

	return agendamento.ID, nil
}

func (c *Conta) ConfigurarChequeEspecial(ctx context.Context, contaID string, novoLimite float64) error {
	msg, err := json.Marshal(map[string]interface{}{
		"tipo":        "limite_cheque_alterado",
		"conta_id":    contaID,
		"novo_limite": novoLimite,
	})
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	return c.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
}

func (c *Conta) ConfigurarNotificacoes(ctx context.Context, contaID string, ativar bool) error {
	msg, err := json.Marshal(map[string]interface{}{
		"tipo":                "notificacoes_alteradas",
		"conta_id":            contaID,
		"notificacoes_ativas": ativar,
	})
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	return c.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
}

func (c *Conta) ObterSaldo(ctx context.Context, contaID string) (float64, error) {
	return 0, fmt.Errorf("não implementado")
}
