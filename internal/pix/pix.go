package pix

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Pix struct {
	producer *kafka.Writer
}

type ChavePix struct {
	ID           string    `json:"id"`
	ContaID      string    `json:"conta_id"`
	TipoChave    string    `json:"tipo_chave"` // CPF, Email, Telefone, Aleatória
	Valor        string    `json:"valor"`
	Status       string    `json:"status"` // ativa, inativa
	DataCriacao  time.Time `json:"data_criacao"`
	LimiteDiario float64   `json:"limite_diario"`
	UltimoAcesso time.Time `json:"ultimo_acesso"`
}

type ContatoPix struct {
	ID          string    `json:"id"`
	ContaID     string    `json:"conta_id"`
	ChaveID     string    `json:"chave_id"`
	Nome        string    `json:"nome"`
	Apelido     string    `json:"apelido"`
	DataCriacao time.Time `json:"data_criacao"`
}

type TransferenciaPix struct {
	ID              string    `json:"id"`
	ChaveOrigem     string    `json:"chave_origem"`
	ChaveDestino    string    `json:"chave_destino"`
	Valor           float64   `json:"valor"`
	DataAgendamento time.Time `json:"data_agendamento,omitempty"`
	Status          string    `json:"status"` // pendente, concluida, cancelada
	DataCriacao     time.Time `json:"data_criacao"`
}

type QRCodePix struct {
	ID          string    `json:"id"`
	ChaveID     string    `json:"chave_id"`
	Valor       float64   `json:"valor,omitempty"` // opcional para QR Code estático
	Tipo        string    `json:"tipo"`            // estatico, dinamico
	Descricao   string    `json:"descricao"`
	DataCriacao time.Time `json:"data_criacao"`
	DataExpira  time.Time `json:"data_expira,omitempty"` // apenas para QR Code dinâmico
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
		ID:           uuid.New().String(),
		ContaID:      contaID,
		TipoChave:    tipoChave,
		Valor:        valorChave,
		Status:       "ativa",
		DataCriacao:  time.Now(),
		LimiteDiario: 5000, // Limite padrão de R$ 5.000,00
		UltimoAcesso: time.Now(),
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

func (p *Pix) ConfigurarLimiteDiario(ctx context.Context, chaveID string, novoLimite float64) error {
	msg, err := json.Marshal(map[string]interface{}{
		"tipo":        "limite_pix_alterado",
		"chave_id":    chaveID,
		"novo_limite": novoLimite,
	})
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	return p.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
}

func (p *Pix) AdicionarContato(ctx context.Context, contaID, chaveID, nome, apelido string) (string, error) {
	contato := ContatoPix{
		ID:          uuid.New().String(),
		ContaID:     contaID,
		ChaveID:     chaveID,
		Nome:        nome,
		Apelido:     apelido,
		DataCriacao: time.Now(),
	}

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":    "contato_pix_adicionado",
		"contato": contato,
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

	return contato.ID, nil
}

func (p *Pix) GerarQRCode(ctx context.Context, chaveID string, tipo string, valor float64, descricao string, dataExpira *time.Time) (string, error) {
	qrcode := QRCodePix{
		ID:          uuid.New().String(),
		ChaveID:     chaveID,
		Tipo:        tipo,
		Valor:       valor,
		Descricao:   descricao,
		DataCriacao: time.Now(),
	}

	if dataExpira != nil {
		qrcode.DataExpira = *dataExpira
	}

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":    "qrcode_pix_gerado",
		"qr_code": qrcode,
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

	return qrcode.ID, nil
}

func (p *Pix) AgendarTransferencia(ctx context.Context, chaveOrigem, chaveDestino string, valor float64, dataAgendamento time.Time) (string, error) {
	transferencia := TransferenciaPix{
		ID:              uuid.New().String(),
		ChaveOrigem:     chaveOrigem,
		ChaveDestino:    chaveDestino,
		Valor:           valor,
		DataAgendamento: dataAgendamento,
		Status:          "pendente",
		DataCriacao:     time.Now(),
	}

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":          "transferencia_pix_agendada",
		"transferencia": transferencia,
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

	return transferencia.ID, nil
}

func (p *Pix) Transferir(ctx context.Context, chaveOrigem, chaveDestino string, valor float64) (string, error) {
	transferencia := TransferenciaPix{
		ID:           uuid.New().String(),
		ChaveOrigem:  chaveOrigem,
		ChaveDestino: chaveDestino,
		Valor:        valor,
		Status:       "concluida",
		DataCriacao:  time.Now(),
	}

	msg, err := json.Marshal(map[string]interface{}{
		"tipo":          "transferencia_pix",
		"transferencia": transferencia,
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

	return transferencia.ID, nil
}

func (p *Pix) CancelarAgendamento(ctx context.Context, transferenciaID string) error {
	msg, err := json.Marshal(map[string]interface{}{
		"tipo":             "agendamento_pix_cancelado",
		"transferencia_id": transferenciaID,
	})
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	return p.producer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
}
