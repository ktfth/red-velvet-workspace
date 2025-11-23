package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/red-velvet-workspace/banco-digital/internal/domain/models"
	"github.com/red-velvet-workspace/banco-digital/internal/infrastructure/kafka"
	"github.com/red-velvet-workspace/banco-digital/internal/repositories"
	"gorm.io/gorm"
)

type AccountService struct {
	notificationRepo *repositories.NotificationRepository
	db               *gorm.DB
	producer         *kafka.Producer
}

func NewAccountService(db *gorm.DB, producer *kafka.Producer) (*AccountService, error) {
	return &AccountService{
		notificationRepo: repositories.NewNotificationRepository(db),
		db:               db,
		producer:         producer,
	}, nil
}

func (s *AccountService) Close() error {
	return nil
}

func (s *AccountService) createNotification(ctx context.Context, accountID uuid.UUID, notificationType string, message string) error {
	notification := &models.Notification{
		ID:        uuid.New(),
		AccountID: accountID,
		Type:      notificationType,
		Message:   message,
		Read:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return s.notificationRepo.Create(ctx, notification)
}

func (s *AccountService) CriarConta(ctx context.Context, accountType models.AccountType) (*models.APIResponse, error) {
	account := &models.Account{
		ID:        uuid.New(),
		Type:      accountType,
		Number:    fmt.Sprintf("%d", time.Now().UnixNano())[:10],
		Balance:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Publica no Kafka
	message := kafka.AccountMessage{
		Operation: "CREATE",
		Account:   *account,
	}
	if err := s.producer.PublishMessage(kafka.TopicAccounts, message); err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("erro ao publicar criação da conta: %v", err),
			Data:    nil,
		}, nil
	}

	err := s.createNotification(ctx, account.ID, "WELCOME",
		"Bem-vindo ao Banco Digital! Sua conta foi criada com sucesso.")
	if err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("erro ao criar notificação de boas-vindas: %v", err),
			Data:    nil,
		}, nil
	}

	return &models.APIResponse{
		Success: true,
		Data:    account,
		Message: "Conta criada com sucesso",
	}, nil
}

func (s *AccountService) AlterarStatus(ctx context.Context, accountID uuid.UUID, status string) (*models.APIResponse, error) {
	err := s.createNotification(ctx, accountID, "STATUS_CHANGE",
		fmt.Sprintf("O status da sua conta foi alterado para: %s", status))
	if err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Erro ao criar notificação de alteração de status: %v", err),
			Data:    nil,
		}, nil
	}

	return &models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Status da conta alterado para: %s", status),
		Data:    map[string]string{"status": status},
	}, nil
}

func (s *AccountService) ConfigurarChequeEspecial(ctx context.Context, accountID uuid.UUID, limite float64) (*models.APIResponse, error) {
	err := s.createNotification(ctx, accountID, "OVERDRAFT_LIMIT",
		fmt.Sprintf("Seu limite de cheque especial foi configurado para: R$ %.2f", limite))
	if err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Erro ao criar notificação de cheque especial: %v", err),
			Data:    nil,
		}, nil
	}

	return &models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Limite de cheque especial configurado para: R$ %.2f", limite),
		Data: map[string]float64{
			"limite": limite,
		},
	}, nil
}

func (s *AccountService) ObterNotificacoes(ctx context.Context, accountID uuid.UUID) (*models.APIResponse, error) {
	if err := s.notificationRepo.DeleteOldNotifications(ctx, accountID, "30 days"); err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("erro ao limpar notificações antigas: %v", err),
			Data:    nil,
		}, nil
	}

	notifications, err := s.notificationRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("erro ao obter notificações: %v", err),
			Data:    nil,
		}, nil
	}

	return &models.APIResponse{
		Success: true,
		Message: "Notificações obtidas com sucesso",
		Data:    notifications,
	}, nil
}

func (s *AccountService) RealizarTransacao(ctx context.Context, accountID uuid.UUID, tipo models.TransactionType, valor float64, descricao string, chaveDestino *string, cartaoID *uuid.UUID) (*models.APIResponse, error) {
	// 1. Validar se a conta existe
	var account models.Account
	if err := s.db.WithContext(ctx).First(&account, "id = ?", accountID).Error; err != nil {
		return &models.APIResponse{
			Success: false,
			Message: "Conta não encontrada",
			Data:    nil,
		}, nil
	}

	// 2. Validações de Saldo/Limite (Otimista)
	switch tipo {
	case models.Debit, models.PIXSent:
		// Verificar saldo + cheque especial (se houver lógica de cheque especial, por enquanto saldo simples)
		// TODO: Adicionar lógica de cheque especial se necessário
		if account.Balance < valor {
			return &models.APIResponse{
				Success: false,
				Message: "Saldo insuficiente",
				Data:    nil,
			}, nil
		}
	case models.CardPurchase:
		if cartaoID == nil {
			return &models.APIResponse{
				Success: false,
				Message: "ID do cartão é obrigatório para compras",
				Data:    nil,
			}, nil
		}
		var card models.CreditCard
		if err := s.db.WithContext(ctx).First(&card, "id = ?", *cartaoID).Error; err != nil {
			return &models.APIResponse{
				Success: false,
				Message: "Cartão não encontrado",
				Data:    nil,
			}, nil
		}
		if card.AvailableLimit < valor {
			return &models.APIResponse{
				Success: false,
				Message: "Limite insuficiente",
				Data:    nil,
			}, nil
		}
	}

	// 3. Criar objeto de transação
	transaction := models.Transaction{
		ID:             uuid.New(),
		AccountID:      accountID,
		Type:           tipo,
		Amount:         valor,
		Description:    descricao,
		DestinationKey: chaveDestino,
		CreditCardID:   cartaoID,
		CreatedAt:      time.Now(),
	}

	// 4. Publicar no Kafka
	message := kafka.TransactionMessage{
		Operation:   "CREATE",
		Transaction: transaction,
	}

	if err := s.producer.PublishMessage(kafka.TopicTransactions, message); err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("erro ao processar transação: %v", err),
			Data:    nil,
		}, nil
	}

	// 5. Mensagem de sucesso
	var msgSuccess string
	switch tipo {
	case models.Credit:
		msgSuccess = fmt.Sprintf("Depósito em processamento: R$ %.2f", valor)
	case models.Debit:
		msgSuccess = fmt.Sprintf("Saque em processamento: R$ %.2f", valor)
	case models.PIXSent:
		msgSuccess = fmt.Sprintf("PIX em processamento: R$ %.2f", valor)
	case models.PIXReceived:
		msgSuccess = fmt.Sprintf("PIX recebido em processamento: R$ %.2f", valor)
	case models.CardPurchase:
		msgSuccess = fmt.Sprintf("Compra em processamento: R$ %.2f", valor)
	case models.CardPayment:
		msgSuccess = fmt.Sprintf("Pagamento de fatura em processamento: R$ %.2f", valor)
	}

	// Notificação será criada pelo consumidor após sucesso, ou podemos criar uma "Em processamento" aqui
	// Por simplicidade, vamos deixar o consumidor criar a notificação final de sucesso/falha
	// Mas para feedback imediato, podemos criar uma aqui também.
	_ = s.createNotification(ctx, accountID, "TRANSACTION_PENDING", msgSuccess)

	return &models.APIResponse{
		Success: true,
		Message: msgSuccess,
		Data:    transaction,
	}, nil
}

func (s *AccountService) CriarCartao(ctx context.Context, accountID uuid.UUID, limite float64) (*models.APIResponse, error) {
	card := &models.CreditCard{
		ID:             uuid.New(),
		AccountID:      accountID,
		AvailableLimit: limite,
		CreditLimit:    limite * 0.7, // 70% of total limit
		Number:         fmt.Sprintf("4532-%s", uuid.New().String()[:12]),
		DueDate:        10, // default to 10th
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Publish to Kafka
	message := kafka.CreditCardMessage{
		Operation:  "CREATE",
		CreditCard: *card,
	}
	if err := s.producer.PublishMessage(kafka.TopicCreditCards, message); err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("erro ao publicar criação do cartão: %v", err),
			Data:    nil,
		}, nil
	}

	if err := s.createNotification(ctx, accountID, "CARD_CREATED",
		fmt.Sprintf("Cartão criado com limite de: R$ %.2f", limite)); err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("erro ao criar notificação de cartão: %v", err),
			Data:    nil,
		}, nil
	}

	return &models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Cartão criado com limite de: R$ %.2f", limite),
		Data:    card,
	}, nil
}

func (s *AccountService) AlterarStatusCartao(ctx context.Context, cardID uuid.UUID, status string) (*models.APIResponse, error) {
	err := s.createNotification(ctx, cardID, "CARD_STATUS_CHANGE",
		fmt.Sprintf("O status do seu cartão foi alterado para: %s", status))
	if err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Erro ao criar notificação de alteração de status do cartão: %v", err),
			Data:    nil,
		}, nil
	}

	return &models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Status do cartão alterado para: %s", status),
		Data: map[string]string{
			"status": status,
		},
	}, nil
}

func (s *AccountService) AlterarLimiteCartao(ctx context.Context, cardID uuid.UUID, limite float64) (*models.APIResponse, error) {
	err := s.createNotification(ctx, cardID, "CARD_LIMIT_CHANGE",
		fmt.Sprintf("O limite do seu cartão foi alterado para: R$ %.2f", limite))
	if err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Erro ao criar notificação de alteração de limite: %v", err),
			Data:    nil,
		}, nil
	}

	return &models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Limite do cartão alterado para: R$ %.2f", limite),
		Data: map[string]float64{
			"limite": limite,
		},
	}, nil
}

func (s *AccountService) GerarCartaoVirtual(ctx context.Context, cardID uuid.UUID) (*models.APIResponse, error) {
	card := &models.VirtualCard{
		ID:           uuid.New(),
		CreditCardID: cardID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.createNotification(ctx, cardID, "VIRTUAL_CARD_CREATED",
		"Cartão virtual criado com sucesso"); err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("erro ao criar notificação de cartão virtual: %v", err),
			Data:    nil,
		}, nil
	}

	return &models.APIResponse{
		Success: true,
		Message: "Cartão virtual criado com sucesso",
		Data:    card,
	}, nil
}

func (s *AccountService) RegistrarChavePix(ctx context.Context, accountID uuid.UUID, keyType string, key string) (*models.APIResponse, error) {
	pixKey := models.PIXKey{
		ID:        uuid.New(),
		AccountID: accountID,
		KeyType:   keyType,
		Key:       key,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Publish to Kafka
	message := kafka.PIXKeyMessage{
		Operation: "CREATE",
		PIXKey:    pixKey,
	}
	if err := s.producer.PublishMessage(kafka.TopicPIX, message); err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("erro ao publicar registro de chave PIX: %v", err),
			Data:    nil,
		}, nil
	}

	if err := s.createNotification(ctx, accountID, "PIX_KEY_REGISTERED",
		fmt.Sprintf("Chave PIX registrada com sucesso: %s - %s", keyType, key)); err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("erro ao criar notificação de chave PIX: %v", err),
			Data:    nil,
		}, nil
	}

	return &models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Chave PIX registrada com sucesso: %s - %s", keyType, key),
		Data:    pixKey,
	}, nil
}

func (s *AccountService) GerarQRCodePix(ctx context.Context, accountID uuid.UUID, valor float64, descricao string) (*models.APIResponse, error) {
	qrCode := &models.PIXQRCode{
		ID:          uuid.New(),
		AccountID:   accountID,
		Amount:      valor,
		Description: descricao,
		CreatedAt:   time.Now(),
	}

	if err := s.createNotification(ctx, accountID, "PIX_QR_CODE_GENERATED",
		fmt.Sprintf("Código QR PIX gerado: R$ %.2f - %s", valor, descricao)); err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("erro ao criar notificação de QR code: %v", err),
			Data:    nil,
		}, nil
	}

	return &models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("QR Code PIX gerado com sucesso: R$ %.2f", valor),
		Data:    qrCode,
	}, nil
}

func (s *AccountService) CancelarAgendamentoPix(ctx context.Context, pixID uuid.UUID) (*models.APIResponse, error) {
	if err := s.createNotification(ctx, pixID, "PIX_SCHEDULING_CANCELLED",
		"Agendamento PIX cancelado com sucesso"); err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("erro ao criar notificação de cancelamento: %v", err),
			Data:    nil,
		}, nil
	}

	return &models.APIResponse{
		Success: true,
		Message: "Agendamento PIX cancelado com sucesso",
		Data:    map[string]interface{}{"pixID": pixID},
	}, nil
}

func (s *AccountService) UpdateAccountStatus(ctx context.Context, req models.UpdateAccountStatusRequest) (*models.APIResponse, error) {
	var account models.Account
	if err := s.db.WithContext(ctx).First(&account, "id = ?", req.AccountID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &models.APIResponse{
				Success: false,
				Message: "Account not found",
				Data:    nil,
			}, nil
		}
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to find account: %v", err),
			Data:    nil,
		}, nil
	}

	account.Status = req.Status
	account.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Save(&account).Error; err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update account status: %v", err),
			Data:    nil,
		}, nil
	}

	// Use o método createNotification existente ao invés de criar diretamente
	if err := s.createNotification(ctx, account.ID, "STATUS_CHANGE",
		fmt.Sprintf("Account status updated to %s", req.Status)); err != nil {
		return &models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create notification: %v", err),
			Data:    nil,
		}, nil
	}

	return &models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Account status updated to %s", req.Status),
		Data:    account,
	}, nil
}
