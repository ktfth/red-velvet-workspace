package pix

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Pix struct {
	producer *kafka.Writer
}

type ChavePix struct {
	ID        string `json:"id"`
	ContaID   string `json:"conta_id"`
	TipoChave string `json:"tipo_chave"` // CPF, Email, Telefone, Aleatória
	Valor     string `json:"valor"`
}

func (p *Pix) Init(ctx context.Context) error {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "kafka:9092"
	}

	p.producer = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokers},
		Topic:   "transacoes_pix",
	})
	return nil
}

func (p *Pix) RegistrarChave(ctx context.Context, contaID, tipoChave, valorChave string) (string, error) {
	chave := ChavePix{
		ID:        uuid.New().String(),
		ContaID:   contaID,
		TipoChave: tipoChave,
		Valor:     valorChave,
	}

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":      "chave_pix_registrada",
		"chave_pix": chave,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	err = p.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao publicar mensagem: %v", err)
	}

	return chave.ID, nil
}

func (p *Pix) Transferir(ctx context.Context, chaveOrigem, chaveDestino string, valor float64) (string, error) {
	transacaoID := uuid.New().String()

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":          "transferencia_pix",
		"transacao_id":  transacaoID,
		"chave_origem":  chaveOrigem,
		"chave_destino": chaveDestino,
		"valor":         valor,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	err = p.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao publicar mensagem: %v", err)
	}

	return transacaoID, nil
}
