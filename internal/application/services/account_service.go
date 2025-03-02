package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/red-velvet-workspace/banco-digital/internal/domain/models"
	"github.com/red-velvet-workspace/banco-digital/internal/infrastructure/database"
	"github.com/red-velvet-workspace/banco-digital/internal/infrastructure/kafka"
)

type AccountService struct {
	producer *kafka.Producer
}

func NewAccountService(kafkaBrokers []string) (*AccountService, error) {
	producer, err := kafka.NewProducer(kafkaBrokers)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %v", err)
	}

	return &AccountService{
		producer: producer,
	}, nil
}

func (s *AccountService) CreateAccount(accountType models.AccountType) (*models.Account, error) {
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
		return nil, fmt.Errorf("failed to publish account creation: %v", err)
	}

	return account, nil
}

func (s *AccountService) CreateCreditCard(accountID uuid.UUID, limit float64) (*models.CreditCard, error) {
	card := &models.CreditCard{
		ID:             uuid.New(),
		AccountID:      accountID,
		Number:         fmt.Sprintf("%d", time.Now().UnixNano())[:16],
		ExpirationDate: time.Now().AddDate(5, 0, 0),
		CVV:            fmt.Sprintf("%d", time.Now().UnixNano())[:3],
		CreditLimit:    limit,
		AvailableLimit: limit,
		StatementDate:  5,
		DueDate:        15,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Publica no Kafka
	message := kafka.CreditCardMessage{
		Operation:   "CREATE",
		CreditCard:  *card,
	}
	if err := s.producer.PublishMessage(kafka.TopicCreditCards, message); err != nil {
		return nil, fmt.Errorf("failed to publish credit card creation: %v", err)
	}

	return card, nil
}

func (s *AccountService) RegisterPIXKey(accountID uuid.UUID, keyType, key string) (*models.PIXKey, error) {
	pixKey := &models.PIXKey{
		ID:        uuid.New(),
		AccountID: accountID,
		KeyType:   keyType,
		Key:       key,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Publica no Kafka
	message := kafka.PIXKeyMessage{
		Operation: "CREATE",
		PIXKey:    *pixKey,
	}
	if err := s.producer.PublishMessage(kafka.TopicPIX, message); err != nil {
		return nil, fmt.Errorf("failed to publish PIX key registration: %v", err)
	}

	return pixKey, nil
}

func (s *AccountService) MakeTransaction(
	accountID uuid.UUID,
	transactionType models.TransactionType,
	amount float64,
	description string,
	destinationKey *string,
	creditCardID *uuid.UUID,
) (*models.Transaction, error) {
	var account models.Account
	if err := database.DB.First(&account, accountID).Error; err != nil {
		return nil, err
	}

	transaction := &models.Transaction{
		ID:             uuid.New(),
		AccountID:      accountID,
		Type:           transactionType,
		Amount:         amount,
		Description:    description,
		DestinationKey: destinationKey,
		CreditCardID:   creditCardID,
		CreatedAt:      time.Now(),
	}

	// Publica a transação no Kafka
	message := kafka.TransactionMessage{
		Transaction: *transaction,
	}
	if err := s.producer.PublishMessage(kafka.TopicTransactions, message); err != nil {
		return nil, fmt.Errorf("failed to publish transaction: %v", err)
	}

	switch transactionType {
	case models.Debit, models.PIXSent:
		if account.Balance < amount {
			return nil, errors.New("insufficient funds")
		}
		account.Balance -= amount
		
		// Publica atualização da conta no Kafka
		accountMsg := kafka.AccountMessage{
			Operation: "UPDATE",
			Account:   account,
		}
		if err := s.producer.PublishMessage(kafka.TopicAccounts, accountMsg); err != nil {
			return nil, fmt.Errorf("failed to publish account update: %v", err)
		}

	case models.Credit, models.PIXReceived:
		account.Balance += amount
		
		// Publica atualização da conta no Kafka
		accountMsg := kafka.AccountMessage{
			Operation: "UPDATE",
			Account:   account,
		}
		if err := s.producer.PublishMessage(kafka.TopicAccounts, accountMsg); err != nil {
			return nil, fmt.Errorf("failed to publish account update: %v", err)
		}

	case models.CardPurchase:
		if creditCardID == nil {
			return nil, errors.New("credit card ID is required for card purchase")
		}
		var card models.CreditCard
		if err := database.DB.First(&card, creditCardID).Error; err != nil {
			return nil, err
		}
		if card.AvailableLimit < amount {
			return nil, errors.New("insufficient credit limit")
		}
		card.AvailableLimit -= amount
		
		// Publica atualização do cartão no Kafka
		cardMsg := kafka.CreditCardMessage{
			Operation:   "UPDATE",
			CreditCard:  card,
		}
		if err := s.producer.PublishMessage(kafka.TopicCreditCards, cardMsg); err != nil {
			return nil, fmt.Errorf("failed to publish credit card update: %v", err)
		}

	case models.CardPayment:
		if creditCardID == nil {
			return nil, errors.New("credit card ID is required for card payment")
		}
		var card models.CreditCard
		if err := database.DB.First(&card, creditCardID).Error; err != nil {
			return nil, err
		}
		if account.Balance < amount {
			return nil, errors.New("insufficient funds for credit card payment")
		}
		account.Balance -= amount
		card.AvailableLimit += amount
		
		// Publica atualizações no Kafka
		accountMsg := kafka.AccountMessage{
			Operation: "UPDATE",
			Account:   account,
		}
		if err := s.producer.PublishMessage(kafka.TopicAccounts, accountMsg); err != nil {
			return nil, fmt.Errorf("failed to publish account update: %v", err)
		}

		cardMsg := kafka.CreditCardMessage{
			Operation:   "UPDATE",
			CreditCard:  card,
		}
		if err := s.producer.PublishMessage(kafka.TopicCreditCards, cardMsg); err != nil {
			return nil, fmt.Errorf("failed to publish credit card update: %v", err)
		}
	}

	return transaction, nil
}

func (s *AccountService) Close() error {
	return s.producer.Close()
}
