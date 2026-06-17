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
)

type Transaction struct {
	ID          int       `json:"id"`
	UsuarioID   int       `json:"usuario_id"`
	Tipo        string    `json:"tipo"`
	Valor       float64   `json:"valor"`
	Descricao   string    `json:"descricao"`
	DataCriacao time.Time `json:"data_criacao"`
}

var db *sql.DB

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/financeiro?sslmode=disable"
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

	fmt.Println("Servidor rodando na porta 8080...")
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

func addTransaction(w http.ResponseWriter, r *http.Request) {
	var t Transaction
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	if t.Tipo != "receita" && t.Tipo != "despesa" {
		http.Error(w, "Tipo deve ser 'receita' ou 'despesa'", http.StatusBadRequest)
		return
	}

	var desc sql.NullString
	if t.Descricao != "" {
		desc.String = t.Descricao
		desc.Valid = true
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
