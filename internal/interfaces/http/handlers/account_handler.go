package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/red-velvet-workspace/banco-digital/internal/application/services"
	"github.com/red-velvet-workspace/banco-digital/internal/domain/models"
)

type AccountHandler struct {
	accountService *services.AccountService
}

func NewAccountHandler(kafkaBrokers []string) (*AccountHandler, error) {
	service, err := services.NewAccountService(kafkaBrokers)
	if err != nil {
		return nil, err
	}
	return &AccountHandler{
		accountService: service,
	}, nil
}

type CreateAccountRequest struct {
	Type models.AccountType `json:"type"`
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	account, err := h.accountService.CreateAccount(req.Type)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

type CreateCreditCardRequest struct {
	AccountID uuid.UUID `json:"account_id"`
	Limit     float64   `json:"limit"`
}

func (h *AccountHandler) CreateCreditCard(w http.ResponseWriter, r *http.Request) {
	var req CreateCreditCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	card, err := h.accountService.CreateCreditCard(req.AccountID, req.Limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(card)
}

type RegisterPIXKeyRequest struct {
	AccountID uuid.UUID `json:"account_id"`
	KeyType   string    `json:"key_type"`
	Key       string    `json:"key"`
}

func (h *AccountHandler) RegisterPIXKey(w http.ResponseWriter, r *http.Request) {
	var req RegisterPIXKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pixKey, err := h.accountService.RegisterPIXKey(req.AccountID, req.KeyType, req.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pixKey)
}

type TransactionRequest struct {
	AccountID      uuid.UUID            `json:"account_id"`
	Type           models.TransactionType `json:"type"`
	Amount         float64              `json:"amount"`
	Description    string               `json:"description"`
	DestinationKey *string             `json:"destination_key,omitempty"`
	CreditCardID   *uuid.UUID          `json:"credit_card_id,omitempty"`
}

func (h *AccountHandler) MakeTransaction(w http.ResponseWriter, r *http.Request) {
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	transaction, err := h.accountService.MakeTransaction(
		req.AccountID,
		req.Type,
		req.Amount,
		req.Description,
		req.DestinationKey,
		req.CreditCardID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

func RegisterAccountRoutes(router *mux.Router, kafkaBrokers []string) error {
	handler, err := NewAccountHandler(kafkaBrokers)
	if err != nil {
		return err
	}
	router.HandleFunc("/accounts", handler.CreateAccount).Methods("POST")
	router.HandleFunc("/accounts/credit-cards", handler.CreateCreditCard).Methods("POST")
	router.HandleFunc("/accounts/pix-keys", handler.RegisterPIXKey).Methods("POST")
	router.HandleFunc("/accounts/transactions", handler.MakeTransaction).Methods("POST")
	return nil
}
