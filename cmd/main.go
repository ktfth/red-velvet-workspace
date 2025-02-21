package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/red-velvet-workspace/banco-digital/internal/cartao"
	"github.com/red-velvet-workspace/banco-digital/internal/conta"
	"github.com/red-velvet-workspace/banco-digital/internal/pix"
)

func main() {
	ctx := context.Background()

	// Inicializar os serviços
	contaService := &conta.Conta{}
	pixService := &pix.Pix{}
	cartaoService := &cartao.Cartao{}

	if err := contaService.Init(ctx); err != nil {
		log.Fatal(err)
	}
	if err := pixService.Init(ctx); err != nil {
		log.Fatal(err)
	}
	if err := cartaoService.Init(ctx); err != nil {
		log.Fatal(err)
	}

	// Configurar rotas HTTP
	http.HandleFunc("/conta/criar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		titular := r.FormValue("titular")
		if titular == "" {
			http.Error(w, "Titular é obrigatório", http.StatusBadRequest)
			return
		}

		id, err := contaService.Criar(r.Context(), titular, 0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Conta criada com sucesso! ID: %s", id)
	})

	http.HandleFunc("/conta/depositar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		contaID := r.FormValue("conta_id")
		valorStr := r.FormValue("valor")
		categoria := r.FormValue("categoria")
		descricao := r.FormValue("descricao")

		if contaID == "" || valorStr == "" {
			http.Error(w, "Conta ID e valor são obrigatórios", http.StatusBadRequest)
			return
		}

		valor, err := strconv.ParseFloat(valorStr, 64)
		if err != nil {
			http.Error(w, "Valor inválido", http.StatusBadRequest)
			return
		}

		if categoria == "" {
			categoria = "Geral"
		}

		transacaoID, err := contaService.Depositar(r.Context(), contaID, valor, categoria, descricao)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Depósito realizado com sucesso! ID da transação: %s", transacaoID)
	})

	http.HandleFunc("/conta/sacar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		contaID := r.FormValue("conta_id")
		valorStr := r.FormValue("valor")
		categoria := r.FormValue("categoria")
		descricao := r.FormValue("descricao")

		if contaID == "" || valorStr == "" {
			http.Error(w, "Conta ID e valor são obrigatórios", http.StatusBadRequest)
			return
		}

		valor, err := strconv.ParseFloat(valorStr, 64)
		if err != nil {
			http.Error(w, "Valor inválido", http.StatusBadRequest)
			return
		}

		if categoria == "" {
			categoria = "Geral"
		}

		transacaoID, err := contaService.Sacar(r.Context(), contaID, valor, categoria, descricao)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Saque realizado com sucesso! ID da transação: %s", transacaoID)
	})

	http.HandleFunc("/conta/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		contaID := r.FormValue("conta_id")
		novoStatus := r.FormValue("status")

		if contaID == "" || novoStatus == "" {
			http.Error(w, "Conta ID e status são obrigatórios", http.StatusBadRequest)
			return
		}

		err := contaService.AlterarStatus(r.Context(), contaID, novoStatus)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Status da conta alterado com sucesso para: %s", novoStatus)
	})

	http.HandleFunc("/conta/agendar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		contaID := r.FormValue("conta_id")
		valorStr := r.FormValue("valor")
		dataStr := r.FormValue("data")
		beneficiario := r.FormValue("beneficiario")

		if contaID == "" || valorStr == "" || dataStr == "" || beneficiario == "" {
			http.Error(w, "Todos os campos são obrigatórios", http.StatusBadRequest)
			return
		}

		valor, err := strconv.ParseFloat(valorStr, 64)
		if err != nil {
			http.Error(w, "Valor inválido", http.StatusBadRequest)
			return
		}

		data, err := time.Parse("2006-01-02", dataStr)
		if err != nil {
			http.Error(w, "Data inválida. Use o formato AAAA-MM-DD", http.StatusBadRequest)
			return
		}

		agendamentoID, err := contaService.AgendarPagamento(r.Context(), contaID, valor, data, beneficiario)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Pagamento agendado com sucesso! ID: %s", agendamentoID)
	})

	http.HandleFunc("/conta/cheque-especial", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		contaID := r.FormValue("conta_id")
		limiteStr := r.FormValue("limite")

		if contaID == "" || limiteStr == "" {
			http.Error(w, "Conta ID e limite são obrigatórios", http.StatusBadRequest)
			return
		}

		limite, err := strconv.ParseFloat(limiteStr, 64)
		if err != nil {
			http.Error(w, "Limite inválido", http.StatusBadRequest)
			return
		}

		err = contaService.ConfigurarChequeEspecial(r.Context(), contaID, limite)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Limite do cheque especial alterado com sucesso para: R$ %.2f", limite)
	})

	http.HandleFunc("/conta/notificacoes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		contaID := r.FormValue("conta_id")
		ativarStr := r.FormValue("ativar")

		if contaID == "" || ativarStr == "" {
			http.Error(w, "Conta ID e status das notificações são obrigatórios", http.StatusBadRequest)
			return
		}

		ativar := ativarStr == "true"

		err := contaService.ConfigurarNotificacoes(r.Context(), contaID, ativar)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		status := "ativadas"
		if !ativar {
			status = "desativadas"
		}
		fmt.Fprintf(w, "Notificações %s com sucesso!", status)
	})

	http.HandleFunc("/pix/registrar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		contaID := r.FormValue("conta_id")
		tipoChave := r.FormValue("tipo_chave")
		valorChave := r.FormValue("valor_chave")

		if contaID == "" || tipoChave == "" || valorChave == "" {
			http.Error(w, "Todos os campos são obrigatórios", http.StatusBadRequest)
			return
		}

		id, err := pixService.RegistrarChave(r.Context(), contaID, tipoChave, valorChave)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Chave PIX registrada com sucesso! ID: %s", id)
	})

	http.HandleFunc("/pix/limite", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		chaveID := r.FormValue("chave_id")
		limiteStr := r.FormValue("limite")

		if chaveID == "" || limiteStr == "" {
			http.Error(w, "Chave ID e limite são obrigatórios", http.StatusBadRequest)
			return
		}

		limite, err := strconv.ParseFloat(limiteStr, 64)
		if err != nil {
			http.Error(w, "Limite inválido", http.StatusBadRequest)
			return
		}

		err = pixService.ConfigurarLimiteDiario(r.Context(), chaveID, limite)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Limite diário do PIX alterado com sucesso para: R$ %.2f", limite)
	})

	http.HandleFunc("/pix/contato", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		contaID := r.FormValue("conta_id")
		chaveID := r.FormValue("chave_id")
		nome := r.FormValue("nome")
		apelido := r.FormValue("apelido")

		if contaID == "" || chaveID == "" || nome == "" {
			http.Error(w, "Conta ID, Chave ID e nome são obrigatórios", http.StatusBadRequest)
			return
		}

		id, err := pixService.AdicionarContato(r.Context(), contaID, chaveID, nome, apelido)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Contato PIX adicionado com sucesso! ID: %s", id)
	})

	http.HandleFunc("/pix/qrcode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		chaveID := r.FormValue("chave_id")
		tipo := r.FormValue("tipo")
		valorStr := r.FormValue("valor")
		descricao := r.FormValue("descricao")
		dataExpiraStr := r.FormValue("data_expira")

		if chaveID == "" || tipo == "" || descricao == "" {
			http.Error(w, "Chave ID, tipo e descrição são obrigatórios", http.StatusBadRequest)
			return
		}

		var valor float64
		if valorStr != "" {
			var err error
			valor, err = strconv.ParseFloat(valorStr, 64)
			if err != nil {
				http.Error(w, "Valor inválido", http.StatusBadRequest)
				return
			}
		}

		var dataExpira *time.Time
		if dataExpiraStr != "" {
			data, err := time.Parse("2006-01-02T15:04:05", dataExpiraStr)
			if err != nil {
				http.Error(w, "Data de expiração inválida. Use o formato AAAA-MM-DDThh:mm:ss", http.StatusBadRequest)
				return
			}
			dataExpira = &data
		}

		id, err := pixService.GerarQRCode(r.Context(), chaveID, tipo, valor, descricao, dataExpira)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "QR Code PIX gerado com sucesso! ID: %s", id)
	})

	http.HandleFunc("/pix/transferir", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		chaveOrigem := r.FormValue("chave_origem")
		chaveDestino := r.FormValue("chave_destino")
		valorStr := r.FormValue("valor")

		if chaveOrigem == "" || chaveDestino == "" || valorStr == "" {
			http.Error(w, "Chave origem, chave destino e valor são obrigatórios", http.StatusBadRequest)
			return
		}

		valor, err := strconv.ParseFloat(valorStr, 64)
		if err != nil {
			http.Error(w, "Valor inválido", http.StatusBadRequest)
			return
		}

		id, err := pixService.Transferir(r.Context(), chaveOrigem, chaveDestino, valor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Transferência PIX realizada com sucesso! ID: %s", id)
	})

	http.HandleFunc("/pix/agendar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		chaveOrigem := r.FormValue("chave_origem")
		chaveDestino := r.FormValue("chave_destino")
		valorStr := r.FormValue("valor")
		dataStr := r.FormValue("data")

		if chaveOrigem == "" || chaveDestino == "" || valorStr == "" || dataStr == "" {
			http.Error(w, "Todos os campos são obrigatórios", http.StatusBadRequest)
			return
		}

		valor, err := strconv.ParseFloat(valorStr, 64)
		if err != nil {
			http.Error(w, "Valor inválido", http.StatusBadRequest)
			return
		}

		data, err := time.Parse("2006-01-02T15:04:05", dataStr)
		if err != nil {
			http.Error(w, "Data inválida. Use o formato AAAA-MM-DDThh:mm:ss", http.StatusBadRequest)
			return
		}

		id, err := pixService.AgendarTransferencia(r.Context(), chaveOrigem, chaveDestino, valor, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Transferência PIX agendada com sucesso! ID: %s", id)
	})

	http.HandleFunc("/pix/cancelar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		transferenciaID := r.FormValue("transferencia_id")
		if transferenciaID == "" {
			http.Error(w, "ID da transferência é obrigatório", http.StatusBadRequest)
			return
		}

		err := pixService.CancelarAgendamento(r.Context(), transferenciaID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Agendamento PIX cancelado com sucesso!")
	})

	http.HandleFunc("/cartao/criar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		contaID := r.FormValue("conta_id")
		if contaID == "" {
			http.Error(w, "Conta ID é obrigatório", http.StatusBadRequest)
			return
		}

		id, err := cartaoService.Criar(r.Context(), contaID, 1000) // Limite padrão de 1000
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Cartão criado com sucesso! ID: %s", id)
	})

	http.HandleFunc("/cartao/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		cartaoID := r.FormValue("cartao_id")
		novoStatus := r.FormValue("status")

		if cartaoID == "" || novoStatus == "" {
			http.Error(w, "Cartão ID e status são obrigatórios", http.StatusBadRequest)
			return
		}

		err := cartaoService.AlterarStatus(r.Context(), cartaoID, novoStatus)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Status do cartão alterado com sucesso para: %s", novoStatus)
	})

	http.HandleFunc("/cartao/limite", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		cartaoID := r.FormValue("cartao_id")
		novoLimiteStr := r.FormValue("limite")

		if cartaoID == "" || novoLimiteStr == "" {
			http.Error(w, "Cartão ID e limite são obrigatórios", http.StatusBadRequest)
			return
		}

		novoLimite, err := strconv.ParseFloat(novoLimiteStr, 64)
		if err != nil {
			http.Error(w, "Limite inválido", http.StatusBadRequest)
			return
		}

		err = cartaoService.AlterarLimite(r.Context(), cartaoID, novoLimite)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Limite do cartão alterado com sucesso para: R$ %.2f", novoLimite)
	})

	http.HandleFunc("/cartao/comprar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		cartaoID := r.FormValue("cartao_id")
		valorStr := r.FormValue("valor")
		estabelecimento := r.FormValue("estabelecimento")
		parcelasStr := r.FormValue("parcelas")

		if cartaoID == "" || valorStr == "" || estabelecimento == "" {
			http.Error(w, "Cartão ID, valor e estabelecimento são obrigatórios", http.StatusBadRequest)
			return
		}

		valor, err := strconv.ParseFloat(valorStr, 64)
		if err != nil {
			http.Error(w, "Valor inválido", http.StatusBadRequest)
			return
		}

		parcelas := 1
		if parcelasStr != "" {
			parcelas, err = strconv.Atoi(parcelasStr)
			if err != nil {
				http.Error(w, "Número de parcelas inválido", http.StatusBadRequest)
				return
			}
		}

		compraID, err := cartaoService.Comprar(r.Context(), cartaoID, valor, estabelecimento, parcelas)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Compra realizada com sucesso! ID: %s", compraID)
	})

	http.HandleFunc("/cartao/virtual", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		cartaoID := r.FormValue("cartao_id")
		if cartaoID == "" {
			http.Error(w, "Cartão ID é obrigatório", http.StatusBadRequest)
			return
		}

		numeroVirtual, err := cartaoService.GerarCartaoVirtual(r.Context(), cartaoID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Cartão virtual gerado com sucesso! Número: %s", numeroVirtual)
	})

	http.HandleFunc("/cartao/pagar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		cartaoID := r.FormValue("cartao_id")
		valorStr := r.FormValue("valor")

		if cartaoID == "" || valorStr == "" {
			http.Error(w, "Cartão ID e valor são obrigatórios", http.StatusBadRequest)
			return
		}

		valor, err := strconv.ParseFloat(valorStr, 64)
		if err != nil {
			http.Error(w, "Valor inválido", http.StatusBadRequest)
			return
		}

		err = cartaoService.PagarFatura(r.Context(), cartaoID, valor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Pagamento da fatura realizado com sucesso no valor de R$ %.2f", valor)
	})

	log.Println("Servidor iniciado na porta 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
