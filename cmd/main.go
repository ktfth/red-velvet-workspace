package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
