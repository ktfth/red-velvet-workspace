package models

import (
	"time"

	"github.com/google/uuid"
)

type AccountType string

const (
	Checking AccountType = "CHECKING"
	Savings  AccountType = "SAVINGS"
)

type Account struct {
	ID        uuid.UUID   `json:"id" gorm:"primaryKey;type:uuid"`
	Type      AccountType `json:"type" gorm:"type:varchar(10)"`
	Number    string      `json:"number" gorm:"unique"`
	Status    string      `json:"status"`
	Balance   float64     `json:"balance"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type CreditCard struct {
	ID              uuid.UUID `json:"id" gorm:"primaryKey;type:uuid"`
	AccountID       uuid.UUID `json:"account_id" gorm:"type:uuid"`
	Number          string    `json:"number" gorm:"unique"`
	ExpirationDate  time.Time `json:"expiration_date"`
	CVV             string    `json:"-" gorm:"type:varchar(3)"`
	CreditLimit     float64   `json:"credit_limit"`
	AvailableLimit  float64   `json:"available_limit"`
	StatementDate   int       `json:"statement_date"`
	DueDate         int       `json:"due_date"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type VirtualCard struct {
	ID              uuid.UUID `json:"id" gorm:"primaryKey;type:uuid"`
	CreditCardID    uuid.UUID `json:"credit_card_id" gorm:"type:uuid"`
	Number          string    `json:"number" gorm:"unique"`
	ExpirationDate  time.Time `json:"expiration_date"`
	CVV             string    `json:"-" gorm:"type:varchar(3)"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PIXKey struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid"`
	AccountID uuid.UUID `json:"account_id" gorm:"type:uuid"`
	KeyType   string    `json:"key_type" gorm:"type:varchar(20)"`
	Key       string    `json:"key" gorm:"unique"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PIXQRCode struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey;type:uuid"`
	AccountID   uuid.UUID `json:"account_id" gorm:"type:uuid"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	QRCode      string    `json:"qr_code"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type TransactionType string

const (
	Debit         TransactionType = "DEBIT"
	Credit        TransactionType = "CREDIT"
	PIXSent       TransactionType = "PIX_SENT"
	PIXReceived   TransactionType = "PIX_RECEIVED"
	CardPurchase  TransactionType = "CARD_PURCHASE"
	CardPayment   TransactionType = "CARD_PAYMENT"
)

type Transaction struct {
	ID              uuid.UUID       `json:"id" gorm:"primaryKey;type:uuid"`
	AccountID       uuid.UUID       `json:"account_id" gorm:"type:uuid"`
	Type            TransactionType `json:"type" gorm:"type:varchar(20)"`
	Amount          float64         `json:"amount"`
	Description     string          `json:"description"`
	DestinationKey  *string        `json:"destination_key,omitempty"`
	CreditCardID    *uuid.UUID     `json:"credit_card_id,omitempty" gorm:"type:uuid"`
	CreatedAt       time.Time       `json:"created_at"`
}

type Notification struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid"`
	AccountID uuid.UUID `json:"account_id" gorm:"type:uuid"`
	Type      string    `json:"type" gorm:"type:varchar(50)"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateAccountStatusRequest struct {
	AccountID uuid.UUID `json:"account_id"`
	Status    string    `json:"status"`
}
