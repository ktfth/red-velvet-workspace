package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/red-velvet-workspace/banco-digital/internal/domain/models"
	"github.com/red-velvet-workspace/banco-digital/internal/infrastructure/database"
	"github.com/red-velvet-workspace/banco-digital/internal/infrastructure/kafka"
	"github.com/red-velvet-workspace/banco-digital/internal/middleware"
	"github.com/red-velvet-workspace/banco-digital/internal/services"
)

func main() {
	// Configurar logger para mostrar data/hora
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("[INFO] Iniciando servidor...")

	// Inicializar conexão com o banco de dados
	db, err := database.InitDBConnection()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize database schema
	if err := database.InitDB(db); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Configurar brokers Kafka
	kafkaBrokers := []string{"kafka:9092"}

	// Inicializar produtor Kafka
	producer, err := kafka.NewProducer(kafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Inicializar serviços
	accountService, err := services.NewAccountService(db, producer)
	if err != nil {
		log.Fatalf("Failed to create account service: %v", err)
	}

	// Inicializar consumidores Kafka
	consumer, err := kafka.NewConsumer(kafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Criar canal para sinais de término
	done := make(chan bool)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar consumidores em goroutines
	go func() {
		if err := consumer.ConsumeAccounts(); err != nil {
			log.Printf("Error consuming accounts: %v", err)
			done <- true
		}
	}()

	go func() {
		if err := consumer.ConsumePIXKeys(); err != nil {
			log.Printf("Error consuming PIX keys: %v", err)
			done <- true
		}
	}()

	go func() {
		if err := consumer.ConsumeCreditCards(); err != nil {
			log.Printf("Error consuming credit cards: %v", err)
			done <- true
		}
	}()

	go func() {
		if err := consumer.ConsumeTransactions(); err != nil {
			log.Printf("Error consuming transactions: %v", err)
			done <- true
		}
	}()

	// Configurar rotas HTTP usando gorilla/mux
	router := mux.NewRouter()

	// Middleware global para logging
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[DEBUG] %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	// Middleware CORS
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})

	// Rota de status
	statusRouter := router.PathPrefix("/status").Subrouter()
	statusRouter.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Rotas de Conta
	contaRouter := router.PathPrefix("/conta").Subrouter()
	log.Printf("[DEBUG] Registrando rotas de conta...")

	contaRouter.HandleFunc("/criar", middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Tipo models.AccountType `json:"tipo"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		accountType := req.Tipo
		if accountType == "" {
			accountType = models.Checking
		}

		account, err := accountService.CriarConta(r.Context(), accountType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(account)
	})).Methods("POST")

	// Rota específica para estado
	log.Printf("[DEBUG] Registrando rota PUT /conta/status")
	contaRouter.Methods("PUT").Path("/status").HandlerFunc(middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] Recebida requisição PUT /conta/status")

		var req struct {
			ContaID string `json:"conta_id"`
			Status  string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[ERROR] Erro ao decodificar body da requisição: %v", err)
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Dados recebidos: ContaID=%s, Status=%s", req.ContaID, req.Status)

		if req.ContaID == "" || req.Status == "" {
			log.Printf("[ERROR] Campos obrigatórios faltando: ContaID=%s, Status=%s", req.ContaID, req.Status)
			http.Error(w, "Conta ID e status são obrigatórios", http.StatusBadRequest)
			return
		}

		accountID, err := uuid.Parse(req.ContaID)
		if err != nil {
			log.Printf("[ERROR] ID da conta inválido: %s - %v", req.ContaID, err)
			http.Error(w, "ID da conta inválido", http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Chamando accountService.UpdateAccountStatus")
		account, err := accountService.UpdateAccountStatus(r.Context(), models.UpdateAccountStatusRequest{
			AccountID: accountID,
			Status:    req.Status,
		})
		if err != nil {
			log.Printf("[ERROR] Erro ao atualizar status da conta: %v", err)
			http.Error(w, fmt.Sprintf("Failed to update account status: %v", err), http.StatusInternalServerError)
			return
		}

		accountData := account.Data.(models.Account)
		log.Printf("[INFO] Status da conta %s atualizado com sucesso para: %s", accountData.ID, accountData.Status)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(account)
	})).Methods("PUT")

	contaRouter.HandleFunc("/cheque-especial", middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ContaID string  `json:"conta_id"`
			Limite  float64 `json:"limite"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if req.ContaID == "" {
			http.Error(w, "Conta ID é obrigatório", http.StatusBadRequest)
			return
		}

		accountID, err := uuid.Parse(req.ContaID)
		if err != nil {
			http.Error(w, "ID da conta inválido", http.StatusBadRequest)
			return
		}

		account, err := accountService.ConfigurarChequeEspecial(r.Context(), accountID, req.Limite)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(account)
	})).Methods("POST")

	// Rota específica para notificações
	log.Printf("[DEBUG] Registrando rota GET /conta/notificacoes")
	contaRouter.Methods("GET").Path("/notificacoes").HandlerFunc(middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] Recebida requisição GET /conta/notificacoes")

		var req struct {
			ContaID string `json:"conta_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[ERROR] Erro ao decodificar body da requisição: %v", err)
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Dados recebidos: ContaID=%s", req.ContaID)

		if req.ContaID == "" {
			log.Printf("[ERROR] Campo obrigatório faltando: ContaID")
			http.Error(w, "Conta ID é obrigatório", http.StatusBadRequest)
			return
		}

		accountID, err := uuid.Parse(req.ContaID)
		if err != nil {
			log.Printf("[ERROR] ID da conta inválido: %s - %v", req.ContaID, err)
			http.Error(w, "ID da conta inválido", http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Chamando accountService.ObterNotificacoes")
		notifications, err := accountService.ObterNotificacoes(r.Context(), accountID)
		if err != nil {
			log.Printf("[ERROR] Erro ao obter notificações: %v", err)
			http.Error(w, fmt.Sprintf("Erro ao obter notificações: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("[INFO] Notificações obtidas com sucesso para a conta %s", accountID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(notifications)
	}))

	contaRouter.HandleFunc("/depositar", middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ContaID string  `json:"conta_id"`
			Valor   float64 `json:"valor"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if req.ContaID == "" || req.Valor <= 0 {
			http.Error(w, "Conta ID e valor positivo são obrigatórios", http.StatusBadRequest)
			return
		}

		accountID, err := uuid.Parse(req.ContaID)
		if err != nil {
			http.Error(w, "ID da conta inválido", http.StatusBadRequest)
			return
		}

		transaction, err := accountService.RealizarTransacao(r.Context(), accountID, models.Credit, req.Valor, "", nil, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transaction)
	})).Methods("POST")

	contaRouter.HandleFunc("/sacar", middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ContaID string  `json:"conta_id"`
			Valor   float64 `json:"valor"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if req.ContaID == "" || req.Valor <= 0 {
			http.Error(w, "Conta ID e valor positivo são obrigatórios", http.StatusBadRequest)
			return
		}

		accountID, err := uuid.Parse(req.ContaID)
		if err != nil {
			http.Error(w, "ID da conta inválido", http.StatusBadRequest)
			return
		}

		transaction, err := accountService.RealizarTransacao(r.Context(), accountID, models.Debit, req.Valor, "", nil, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transaction)
	})).Methods("POST")

	// Rotas de Cartão
	cartaoRouter := router.PathPrefix("/cartao").Subrouter()
	log.Printf("[DEBUG] Registrando rotas de cartão...")

	// Rota para criar cartão
	log.Printf("[DEBUG] Registrando rota POST /cartao/criar")
	cartaoRouter.Methods("POST").Path("/criar").HandlerFunc(middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] Recebida requisição POST /cartao/criar")

		var req struct {
			ContaID string  `json:"conta_id"`
			Limite  float64 `json:"limite"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[ERROR] Erro ao decodificar body da requisição: %v", err)
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Dados recebidos: ContaID=%s, Limite=%.2f", req.ContaID, req.Limite)

		if req.ContaID == "" || req.Limite <= 0 {
			log.Printf("[ERROR] Campos obrigatórios faltando ou inválidos: ContaID=%s, Limite=%.2f", req.ContaID, req.Limite)
			http.Error(w, "Conta ID e limite positivo são obrigatórios", http.StatusBadRequest)
			return
		}

		accountID, err := uuid.Parse(req.ContaID)
		if err != nil {
			log.Printf("[ERROR] ID da conta inválido: %s - %v", req.ContaID, err)
			http.Error(w, "ID da conta inválido", http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Chamando accountService.CriarCartao")
		card, err := accountService.CriarCartao(r.Context(), accountID, req.Limite)
		if err != nil {
			log.Printf("[ERROR] Erro ao criar cartão: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("[INFO] Cartão criado com sucesso para a conta %s com limite %.2f", accountID, req.Limite)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(card)
	}))

	// Rota para alterar status do cartão
	log.Printf("[DEBUG] Registrando rota PUT /cartao/status")
	cartaoRouter.Methods("PUT").Path("/status").HandlerFunc(middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] Recebida requisição PUT /cartao/status")

		var req struct {
			CartaoID string `json:"cartao_id"`
			Status   string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[ERROR] Erro ao decodificar body da requisição: %v", err)
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Dados recebidos: CartaoID=%s, Status=%s", req.CartaoID, req.Status)

		if req.CartaoID == "" || req.Status == "" {
			log.Printf("[ERROR] Campos obrigatórios faltando: CartaoID=%s, Status=%s", req.CartaoID, req.Status)
			http.Error(w, "Cartão ID e status são obrigatórios", http.StatusBadRequest)
			return
		}

		cardID, err := uuid.Parse(req.CartaoID)
		if err != nil {
			log.Printf("[ERROR] ID do cartão inválido: %s - %v", req.CartaoID, err)
			http.Error(w, "ID do cartão inválido", http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Chamando accountService.AlterarStatusCartao")
		cartao, err := accountService.AlterarStatusCartao(r.Context(), cardID, req.Status)
		if err != nil {
			log.Printf("[ERROR] Erro ao alterar status do cartão: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("[INFO] Status do cartão %s atualizado com sucesso para: %s", cardID, req.Status)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartao)
	}))

	// Rota para alterar limite do cartão
	log.Printf("[DEBUG] Registrando rota PUT /cartao/limite")
	cartaoRouter.Methods("PUT").Path("/limite").HandlerFunc(middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] Recebida requisição PUT /cartao/limite")

		var req struct {
			CartaoID string  `json:"cartao_id"`
			Limite   float64 `json:"limite"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[ERROR] Erro ao decodificar body da requisição: %v", err)
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Dados recebidos: CartaoID=%s, Limite=%.2f", req.CartaoID, req.Limite)

		if req.CartaoID == "" || req.Limite <= 0 {
			log.Printf("[ERROR] Campos obrigatórios faltando ou inválidos: CartaoID=%s, Limite=%.2f", req.CartaoID, req.Limite)
			http.Error(w, "Cartão ID e limite positivo são obrigatórios", http.StatusBadRequest)
			return
		}

		cardID, err := uuid.Parse(req.CartaoID)
		if err != nil {
			log.Printf("[ERROR] ID do cartão inválido: %s - %v", req.CartaoID, err)
			http.Error(w, "ID do cartão inválido", http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Chamando accountService.AlterarLimiteCartao")
		cartao, err := accountService.AlterarLimiteCartao(r.Context(), cardID, req.Limite)
		if err != nil {
			log.Printf("[ERROR] Erro ao alterar limite do cartão: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("[INFO] Limite do cartão %s atualizado com sucesso para: %.2f", cardID, req.Limite)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartao)
	}))

	// Rota para gerar cartão virtual
	log.Printf("[DEBUG] Registrando rota POST /cartao/virtual")
	cartaoRouter.Methods("POST").Path("/virtual").HandlerFunc(middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] Recebida requisição POST /cartao/virtual")

		var req struct {
			CartaoID string `json:"cartao_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[ERROR] Erro ao decodificar body da requisição: %v", err)
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Dados recebidos: CartaoID=%s", req.CartaoID)

		if req.CartaoID == "" {
			log.Printf("[ERROR] Campo obrigatório faltando: CartaoID")
			http.Error(w, "Cartão ID é obrigatório", http.StatusBadRequest)
			return
		}

		cardID, err := uuid.Parse(req.CartaoID)
		if err != nil {
			log.Printf("[ERROR] ID do cartão inválido: %s - %v", req.CartaoID, err)
			http.Error(w, "ID do cartão inválido", http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Chamando accountService.GerarCartaoVirtual")
		virtualCard, err := accountService.GerarCartaoVirtual(r.Context(), cardID)
		if err != nil {
			log.Printf("[ERROR] Erro ao gerar cartão virtual: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("[INFO] Cartão virtual gerado com sucesso para o cartão %s", cardID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(virtualCard)
	}))

	// Rota para realizar compra com cartão
	log.Printf("[DEBUG] Registrando rota POST /cartao/comprar")
	cartaoRouter.Methods("POST").Path("/comprar").HandlerFunc(middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] Recebida requisição POST /cartao/comprar")

		var req struct {
			ContaID  string  `json:"conta_id"`
			CartaoID string  `json:"cartao_id"`
			Valor    float64 `json:"valor"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[ERROR] Erro ao decodificar body da requisição: %v", err)
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Dados recebidos: ContaID=%s, CartaoID=%s, Valor=%.2f", req.ContaID, req.CartaoID, req.Valor)

		if req.ContaID == "" || req.CartaoID == "" || req.Valor <= 0 {
			log.Printf("[ERROR] Campos obrigatórios faltando ou inválidos: ContaID=%s, CartaoID=%s, Valor=%.2f", req.ContaID, req.CartaoID, req.Valor)
			http.Error(w, "Conta ID, cartão ID e valor positivo são obrigatórios", http.StatusBadRequest)
			return
		}

		accountID, err := uuid.Parse(req.ContaID)
		if err != nil {
			log.Printf("[ERROR] ID da conta inválido: %s - %v", req.ContaID, err)
			http.Error(w, "ID da conta inválido", http.StatusBadRequest)
			return
		}

		cardID, err := uuid.Parse(req.CartaoID)
		if err != nil {
			log.Printf("[ERROR] ID do cartão inválido: %s - %v", req.CartaoID, err)
			http.Error(w, "ID do cartão inválido", http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Chamando accountService.RealizarTransacao para compra com cartão")
		transaction, err := accountService.RealizarTransacao(r.Context(), accountID, models.CardPurchase, req.Valor, "", nil, &cardID)
		if err != nil {
			log.Printf("[ERROR] Erro ao realizar compra com cartão: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		transactionData := transaction.Data.(models.Transaction)
		log.Printf("[INFO] Compra realizada com sucesso: ID=%s, Valor=%.2f", transactionData.ID, transactionData.Amount)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transaction)
	}))

	// Rota para pagar fatura do cartão
	log.Printf("[DEBUG] Registrando rota POST /cartao/pagar")
	cartaoRouter.Methods("POST").Path("/pagar").HandlerFunc(middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] Recebida requisição POST /cartao/pagar")

		var req struct {
			ContaID  string  `json:"conta_id"`
			CartaoID string  `json:"cartao_id"`
			Valor    float64 `json:"valor"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[ERROR] Erro ao decodificar body da requisição: %v", err)
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Dados recebidos: ContaID=%s, CartaoID=%s, Valor=%.2f", req.ContaID, req.CartaoID, req.Valor)

		if req.ContaID == "" || req.CartaoID == "" || req.Valor <= 0 {
			log.Printf("[ERROR] Campos obrigatórios faltando ou inválidos: ContaID=%s, CartaoID=%s, Valor=%.2f", req.ContaID, req.CartaoID, req.Valor)
			http.Error(w, "Conta ID, cartão ID e valor positivo são obrigatórios", http.StatusBadRequest)
			return
		}

		accountID, err := uuid.Parse(req.ContaID)
		if err != nil {
			log.Printf("[ERROR] ID da conta inválido: %s - %v", req.ContaID, err)
			http.Error(w, "ID da conta inválido", http.StatusBadRequest)
			return
		}

		cardID, err := uuid.Parse(req.CartaoID)
		if err != nil {
			log.Printf("[ERROR] ID do cartão inválido: %s - %v", req.CartaoID, err)
			http.Error(w, "ID do cartão inválido", http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Chamando accountService.RealizarTransacao para pagamento de fatura")
		transaction, err := accountService.RealizarTransacao(r.Context(), accountID, models.CardPayment, req.Valor, "", nil, &cardID)
		if err != nil {
			log.Printf("[ERROR] Erro ao realizar pagamento de fatura: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		transactionData := transaction.Data.(models.Transaction)
		log.Printf("[INFO] Pagamento de fatura realizado com sucesso: ID=%s, Valor=%.2f", transactionData.ID, transactionData.Amount)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transaction)
	}))

	// Rotas de PIX
	pixRouter := router.PathPrefix("/pix").Subrouter()

	pixRouter.HandleFunc("/registrar", middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ContaID string `json:"conta_id"`
			Tipo    string `json:"tipo"`
			Chave   string `json:"chave"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if req.ContaID == "" || req.Tipo == "" || req.Chave == "" {
			http.Error(w, "Conta ID, tipo e chave são obrigatórios", http.StatusBadRequest)
			return
		}

		accountID, err := uuid.Parse(req.ContaID)
		if err != nil {
			http.Error(w, "ID da conta inválido", http.StatusBadRequest)
			return
		}

		pixKey, err := accountService.RegistrarChavePix(r.Context(), accountID, req.Tipo, req.Chave)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pixKey)
	})).Methods("POST")

	pixRouter.HandleFunc("/enviar", middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ContaID      string  `json:"conta_id"`
			Valor        float64 `json:"valor"`
			ChaveDestino string  `json:"chave_destino"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if req.ContaID == "" || req.Valor <= 0 || req.ChaveDestino == "" {
			http.Error(w, "Conta ID, valor positivo e chave de destino são obrigatórios", http.StatusBadRequest)
			return
		}

		accountID, err := uuid.Parse(req.ContaID)
		if err != nil {
			http.Error(w, "ID da conta inválido", http.StatusBadRequest)
			return
		}

		transaction, err := accountService.RealizarTransacao(r.Context(), accountID, models.PIXSent, req.Valor, "", &req.ChaveDestino, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transaction)
	})).Methods("POST")

	pixRouter.HandleFunc("/qrcode", middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ContaID string  `json:"conta_id"`
			Valor   float64 `json:"valor"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if req.ContaID == "" || req.Valor <= 0 {
			http.Error(w, "Conta ID e valor positivo são obrigatórios", http.StatusBadRequest)
			return
		}

		accountID, err := uuid.Parse(req.ContaID)
		if err != nil {
			http.Error(w, "ID da conta inválido", http.StatusBadRequest)
			return
		}

		qrCode, err := accountService.GerarQRCodePix(r.Context(), accountID, req.Valor, "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "QR Code PIX gerado com sucesso!",
			"qrcode":  qrCode,
			"valor":   req.Valor,
		})
	})).Methods("POST")

	pixRouter.HandleFunc("/cancelar", middleware.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			PixID string `json:"pix_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if req.PixID == "" {
			http.Error(w, "PIX ID é obrigatório", http.StatusBadRequest)
			return
		}

		pixID, err := uuid.Parse(req.PixID)
		if err != nil {
			http.Error(w, "ID do PIX inválido", http.StatusBadRequest)
			return
		}

		pix, err := accountService.CancelarAgendamentoPix(r.Context(), pixID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pix)
	})).Methods("POST")

	// Log todas as rotas registradas
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		log.Printf("[DEBUG] Rota registrada: %s [%v]", path, methods)
		return nil
	})

	// Configurar servidor HTTP
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Printf("[INFO] Starting HTTP server on %s", srv.Addr)

	// Iniciar servidor HTTP
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("[ERROR] HTTP server error: %v", err)
	}

	// Aguardar sinal de término
	select {
	case <-signals:
		log.Println("Received shutdown signal")
	case <-done:
		log.Println("Shutting down due to error")
	}
}
