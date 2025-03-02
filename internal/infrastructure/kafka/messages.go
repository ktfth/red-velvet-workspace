package kafka

import (
    "github.com/red-velvet-workspace/banco-digital/internal/domain/models"
)

type AccountMessage struct {
    Operation string         `json:"operation"` // CREATE, UPDATE
    Account   models.Account `json:"account"`
}

type PIXKeyMessage struct {
    Operation string        `json:"operation"` // CREATE, DELETE
    PIXKey    models.PIXKey `json:"pix_key"`
}

type CreditCardMessage struct {
    Operation   string            `json:"operation"` // CREATE, UPDATE
    CreditCard  models.CreditCard `json:"credit_card"`
}

type TransactionMessage struct {
    Transaction models.Transaction `json:"transaction"`
}
