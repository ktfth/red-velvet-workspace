package services

import (
	"context"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/red-velvet-workspace/banco-digital/internal/domain/models"
	"github.com/red-velvet-workspace/banco-digital/internal/repositories"
	"gorm.io/gorm"
)

type AccountService struct {
	notificationRepo *repositories.NotificationRepository
	db               *gorm.DB
}

func NewAccountService(db *gorm.DB) (*AccountService, error) {
	return &AccountService{
		notificationRepo: repositories.NewNotificationRepository(db),
		db:               db,
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

func (s *AccountService) CriarConta(ctx context.Context, accountType models.AccountType) (*models.Account, error) {
	account := &models.Account{
		ID:        uuid.New(),
		Type:      accountType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	err := s.createNotification(ctx, account.ID, "WELCOME", 
		"Bem-vindo ao Banco Digital! Sua conta foi criada com sucesso.")
	if err != nil {
		return nil, fmt.Errorf("erro ao criar notificação de boas-vindas: %v", err)
	}
	
	return account, nil
}

func (s *AccountService) AlterarStatus(ctx context.Context, accountID uuid.UUID, status string) error {
	return s.createNotification(ctx, accountID, "STATUS_CHANGE", 
		fmt.Sprintf("O status da sua conta foi alterado para: %s", status))
}

func (s *AccountService) ConfigurarChequeEspecial(ctx context.Context, accountID uuid.UUID, limite float64) error {
	return s.createNotification(ctx, accountID, "OVERDRAFT_LIMIT",
		fmt.Sprintf("Seu limite de cheque especial foi configurado para: R$ %.2f", limite))
}

func (s *AccountService) ObterNotificacoes(ctx context.Context, accountID uuid.UUID) ([]models.Notification, error) {
	if err := s.notificationRepo.DeleteOldNotifications(ctx, accountID, "30 days"); err != nil {
		return nil, fmt.Errorf("erro ao limpar notificações antigas: %v", err)
	}
	return s.notificationRepo.GetByAccountID(ctx, accountID)
}

func (s *AccountService) RealizarTransacao(ctx context.Context, accountID uuid.UUID, tipo models.TransactionType, valor float64, descricao string, chaveDestino *string, cartaoID *uuid.UUID) (*models.Transaction, error) {
	transaction := &models.Transaction{
		ID:          uuid.New(),
		AccountID:   accountID,
		Type:        tipo,
		Amount:      valor,
		Description: descricao,
		CreatedAt:   time.Now(),
	}

	var message string
	switch tipo {
	case models.Credit:
		message = fmt.Sprintf("Depósito realizado: R$ %.2f - %s", valor, descricao)
	case models.Debit:
		message = fmt.Sprintf("Saque realizado: R$ %.2f - %s", valor, descricao)
	case models.PIXSent:
		message = fmt.Sprintf("PIX enviado: R$ %.2f - %s", valor, descricao)
	case models.PIXReceived:
		message = fmt.Sprintf("PIX recebido: R$ %.2f - %s", valor, descricao)
	case models.CardPurchase:
		message = fmt.Sprintf("Compra com cartão: R$ %.2f - %s", valor, descricao)
	case models.CardPayment:
		message = fmt.Sprintf("Pagamento de fatura: R$ %.2f", valor)
	}

	if err := s.createNotification(ctx, accountID, "TRANSACTION", message); err != nil {
		return nil, fmt.Errorf("erro ao criar notificação de transação: %v", err)
	}

	return transaction, nil
}

func (s *AccountService) CriarCartao(ctx context.Context, accountID uuid.UUID, limite float64) (*models.CreditCard, error) {
	card := &models.CreditCard{
		ID:        uuid.New(),
		AccountID: accountID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := s.createNotification(ctx, accountID, "CARD_CREATED", 
		fmt.Sprintf("Cartão criado com limite de: R$ %.2f", limite)); err != nil {
		return nil, fmt.Errorf("erro ao criar notificação de cartão: %v", err)
	}
	
	return card, nil
}

func (s *AccountService) AlterarStatusCartao(ctx context.Context, cardID uuid.UUID, status string) error {
	return s.createNotification(ctx, cardID, "CARD_STATUS_CHANGE",
		fmt.Sprintf("O status do seu cartão foi alterado para: %s", status))
}

func (s *AccountService) AlterarLimiteCartao(ctx context.Context, cardID uuid.UUID, limite float64) error {
	return s.createNotification(ctx, cardID, "CARD_LIMIT_CHANGE",
		fmt.Sprintf("O limite do seu cartão foi alterado para: R$ %.2f", limite))
}

func (s *AccountService) GerarCartaoVirtual(ctx context.Context, cardID uuid.UUID) (*models.VirtualCard, error) {
	card := &models.VirtualCard{
		ID:          uuid.New(),
		CreditCardID: cardID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	if err := s.createNotification(ctx, cardID, "VIRTUAL_CARD_CREATED", 
		"Cartão virtual criado com sucesso"); err != nil {
		return nil, fmt.Errorf("erro ao criar notificação de cartão virtual: %v", err)
	}
	
	return card, nil
}

func (s *AccountService) RegistrarChavePix(ctx context.Context, accountID uuid.UUID, keyType string, key string) (*models.PIXKey, error) {
	pixKey := &models.PIXKey{
		ID:        uuid.New(),
		AccountID: accountID,
		KeyType:   keyType,
		Key:       key,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := s.createNotification(ctx, accountID, "PIX_KEY_REGISTERED",
		fmt.Sprintf("Chave PIX registrada com sucesso: %s - %s", keyType, key)); err != nil {
		return nil, fmt.Errorf("erro ao criar notificação de chave PIX: %v", err)
	}
	
	return pixKey, nil
}

func (s *AccountService) GerarQRCodePix(ctx context.Context, accountID uuid.UUID, valor float64, descricao string) (*models.PIXQRCode, error) {
	qrCode := &models.PIXQRCode{
		ID:          uuid.New(),
		AccountID:   accountID,
		Amount:      valor,
		Description: descricao,
		CreatedAt:   time.Now(),
	}
	
	if err := s.createNotification(ctx, accountID, "PIX_QR_CODE_GENERATED",
		fmt.Sprintf("Código QR PIX gerado: R$ %.2f - %s", valor, descricao)); err != nil {
		return nil, fmt.Errorf("erro ao criar notificação de QR code: %v", err)
	}
	
	return qrCode, nil
}

func (s *AccountService) CancelarAgendamentoPix(ctx context.Context, pixID uuid.UUID) error {
	return s.createNotification(ctx, pixID, "PIX_SCHEDULING_CANCELLED",
		"Agendamento PIX cancelado com sucesso")
}

func (s *AccountService) UpdateAccountStatus(ctx context.Context, req models.UpdateAccountStatusRequest) (*models.Account, error) {
	var account models.Account
	if err := s.db.First(&account, "id = ?", req.AccountID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to find account: %v", err)
	}

	account.Status = req.Status
	account.UpdatedAt = time.Now()

	if err := s.db.Save(&account).Error; err != nil {
		return nil, fmt.Errorf("failed to update account status: %v", err)
	}

	// Create notification for status change
	notification := &models.Notification{
		ID:        uuid.New(),
		AccountID: account.ID,
		Type:      "STATUS_CHANGE",
		Message:   fmt.Sprintf("Account status updated to %s", req.Status),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.Create(notification).Error; err != nil {
		// Log the error but don't fail the status update
		fmt.Printf("Failed to create notification: %v\n", err)
	}

	return &account, nil
}
