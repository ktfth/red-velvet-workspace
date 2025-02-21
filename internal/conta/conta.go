package conta

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Conta struct {
	producer *kafka.Writer
}

type ContaModel struct {
	ID      string  `json:"id"`
	Titular string  `json:"titular"`
	Saldo   float64 `json:"saldo"`
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
		ID:      uuid.New().String(),
		Titular: titular,
		Saldo:   saldoInicial,
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

func (c *Conta) Depositar(ctx context.Context, contaID string, valor float64) error {
	msg, err := json.Marshal(map[string]interface{}{
		"tipo":     "deposito",
		"conta_id": contaID,
		"valor":    valor,
	})
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	return c.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
}

func (c *Conta) Sacar(ctx context.Context, contaID string, valor float64) error {
	msg, err := json.Marshal(map[string]interface{}{
		"tipo":     "saque",
		"conta_id": contaID,
		"valor":    valor,
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
