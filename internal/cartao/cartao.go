package cartao

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Cartao struct {
	producer *kafka.Writer
}

type CartaoCredito struct {
	ID              string    `json:"id"`
	ContaID         string    `json:"conta_id"`
	Limite          float64   `json:"limite"`
	LimiteParcelado float64   `json:"limite_parcelado"`
	LimiteUsado     float64   `json:"limite_usado"`
	NumeroVirtual   string    `json:"numero_virtual"`
	Status          string    `json:"status"`     // ativo, bloqueado
	Vencimento      int       `json:"vencimento"` // dia do vencimento
	DataCriacao     time.Time `json:"data_criacao"`
	FaturaAtual     float64   `json:"fatura_atual"`
}

type Compra struct {
	ID              string    `json:"id"`
	CartaoID        string    `json:"cartao_id"`
	Valor           float64   `json:"valor"`
	Estabelecimento string    `json:"estabelecimento"`
	Data            time.Time `json:"data"`
	Parcelas        int       `json:"parcelas"`
}

func (c *Cartao) Init(ctx context.Context) error {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "kafka:9092"
	}

	c.producer = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokers},
		Topic:   "transacoes_cartao",
	})
	return nil
}

func (c *Cartao) Criar(ctx context.Context, contaID string, limite float64) (string, error) {
	cartao := CartaoCredito{
		ID:              uuid.New().String(),
		ContaID:         contaID,
		Limite:          limite,
		LimiteParcelado: limite * 0.7, // 70% do limite total
		LimiteUsado:     0,
		NumeroVirtual:   fmt.Sprintf("4532-%s", uuid.New().String()[:12]),
		Status:          "ativo",
		Vencimento:      10, // dia 10 como padrão
		DataCriacao:     time.Now(),
		FaturaAtual:     0,
	}

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":   "cartao_criado",
		"cartao": cartao,
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

	return cartao.ID, nil
}

func (c *Cartao) AlterarStatus(ctx context.Context, cartaoID string, novoStatus string) error {
	if novoStatus != "ativo" && novoStatus != "bloqueado" {
		return fmt.Errorf("status inválido: %s", novoStatus)
	}

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":        "status_alterado",
		"cartao_id":   cartaoID,
		"novo_status": novoStatus,
	})
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	return c.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
}

func (c *Cartao) AlterarLimite(ctx context.Context, cartaoID string, novoLimite float64) error {
	msg, err := json.Marshal(map[string]interface{}{
		"tipo":        "limite_alterado",
		"cartao_id":   cartaoID,
		"novo_limite": novoLimite,
	})
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	return c.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
}

func (c *Cartao) AlterarVencimento(ctx context.Context, cartaoID string, novoVencimento int) error {
	if novoVencimento < 1 || novoVencimento > 31 {
		return fmt.Errorf("dia de vencimento inválido: %d", novoVencimento)
	}

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":            "vencimento_alterado",
		"cartao_id":       cartaoID,
		"novo_vencimento": novoVencimento,
	})
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	return c.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
}

func (c *Cartao) Comprar(ctx context.Context, cartaoID string, valor float64, estabelecimento string, parcelas int) (string, error) {
	compraID := uuid.New().String()
	compra := Compra{
		ID:              compraID,
		CartaoID:        cartaoID,
		Valor:           valor,
		Estabelecimento: estabelecimento,
		Data:            time.Now(),
		Parcelas:        parcelas,
	}

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":   "compra_realizada",
		"compra": compra,
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

	return compraID, nil
}

func (c *Cartao) PagarFatura(ctx context.Context, cartaoID string, valor float64) error {
	msg, err := json.Marshal(map[string]interface{}{
		"tipo":      "pagamento_fatura",
		"cartao_id": cartaoID,
		"valor":     valor,
	})
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	return c.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
}

func (c *Cartao) GerarCartaoVirtual(ctx context.Context, cartaoID string) (string, error) {
	numeroVirtual := fmt.Sprintf("4532-%s", uuid.New().String()[:12])

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":           "cartao_virtual_gerado",
		"cartao_id":      cartaoID,
		"numero_virtual": numeroVirtual,
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

	return numeroVirtual, nil
}
