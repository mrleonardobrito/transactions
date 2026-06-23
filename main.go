package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "transactions/docs"
)

type Transaction struct {
	ID          int       `json:"id" example:"1"`
	UsuarioID   int       `json:"usuario_id" example:"2"`
	Tipo        string    `json:"tipo" example:"receita" enums:"receita,despesa"`
	Valor       float64   `json:"valor" example:"1500.50"`
	Descricao   string    `json:"descricao" example:"Salário"`
	DataCriacao time.Time `json:"data_criacao" example:"2026-06-23T16:00:00Z"`
}

type TransactionInput struct {
	UsuarioID int     `json:"usuario_id" example:"2"`
	Tipo      string  `json:"tipo" example:"receita" enums:"receita,despesa"`
	Valor     float64 `json:"valor" example:"1500.50"`
	Descricao string  `json:"descricao" example:"Salário"`
}

var db *sql.DB

// @title           Transactions API
// @version         1.0
// @description     API para gerenciamento de transações financeiras (receitas e despesas).
// @termsOfService  http://swagger.io/terms/

// @contact.name   Suporte
// @contact.email  lbritogit@gmail.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @schemes http
func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://financeiro:financeiro@localhost:5434/financeiro?sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Não foi possível alcançar o banco de dados:", err)
	}

	http.HandleFunc("/transactions", transactionsHandler)
	http.Handle("/docs/", httpSwagger.WrapHandler)

	fmt.Println("Servidor rodando na porta 8080...")
	fmt.Println("Documentação Swagger disponível em http://localhost:8080/docs/index.html")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func transactionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		getTransactions(w, r)
	} else if r.Method == http.MethodPost {
		addTransaction(w, r)
	} else {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
	}
}

// getTransactions godoc
// @Summary      Lista as transações
// @Description  Retorna todas as transações ordenadas pela data de criação (mais recentes primeiro).
// @Tags         transactions
// @Produce      json
// @Success      200  {array}   Transaction
// @Failure      500  {string}  string  "Erro interno ao buscar transações"
// @Router       /transactions [get]
func getTransactions(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT id, usuario_id, tipo, valor, descricao, data_criacao
		FROM operacoes
		ORDER BY data_criacao DESC
	`)
	if err != nil {
		http.Error(w, "Erro ao buscar transações: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	transactions := []Transaction{}
	for rows.Next() {
		var t Transaction
		var desc sql.NullString

		err := rows.Scan(&t.ID, &t.UsuarioID, &t.Tipo, &t.Valor, &desc, &t.DataCriacao)
		if err != nil {
			http.Error(w, "Erro ao processar transações: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if desc.Valid {
			t.Descricao = desc.String
		}

		transactions = append(transactions, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// addTransaction godoc
// @Summary      Cria uma transação
// @Description  Registra uma nova transação financeira. O campo "tipo" deve ser "receita" ou "despesa".
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        transaction  body      TransactionInput  true  "Dados da transação"
// @Success      201          {object}  Transaction
// @Failure      400          {string}  string  "Dados inválidos"
// @Failure      500          {string}  string  "Erro interno ao salvar a transação"
// @Router       /transactions [post]
func addTransaction(w http.ResponseWriter, r *http.Request) {
	var input TransactionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	if input.Tipo != "receita" && input.Tipo != "despesa" {
		http.Error(w, "Tipo deve ser 'receita' ou 'despesa'", http.StatusBadRequest)
		return
	}

	var desc sql.NullString
	if input.Descricao != "" {
		desc.String = input.Descricao
		desc.Valid = true
	}

	t := Transaction{
		UsuarioID: input.UsuarioID,
		Tipo:      input.Tipo,
		Valor:     input.Valor,
		Descricao: input.Descricao,
	}

	err := db.QueryRow(`
		INSERT INTO operacoes (usuario_id, tipo, valor, descricao)
		VALUES ($1, $2, $3, $4)
		RETURNING id, data_criacao
	`, t.UsuarioID, t.Tipo, t.Valor, desc).Scan(&t.ID, &t.DataCriacao)

	if err != nil {
		http.Error(w, "Erro ao salvar transação: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}
