package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

type AccountBalance struct {
	AccountNo        string  `json:"account_no"`
	OpeningBalance   float64 `json:"opening_balance"`
	ClosingBalance   float64 `json:"closing_balance"`
	PercentageChange float64 `json:"percentage_change"`
	Currency         string  `json:"currency"`
	Country          string  `json:"country"`
}

type FilterRequest struct {
	GroupBy string `json:"groupby"` // Either 'opening_balance_currency' or 'country_id'
}

func main() {
	log.Println("Starting server initialization...")

	connStr := "host=localhost port=9003 user=postgres password=arthavedh123 dbname=statement_processing sslmode=disable"

	log.Println("Connecting to database...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Database ping failed:", err)
	}
	log.Println("Database connected successfully")

	r := mux.NewRouter()

	r.HandleFunc("/heatmap/filterdata", func(w http.ResponseWriter, r *http.Request) {
		var req FilterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON request", http.StatusBadRequest)
			return
		}

		if req.GroupBy != "opening_balance_currency" && req.GroupBy != "country_id" {
			http.Error(w, "Invalid groupby value. It must be either 'opening_balance_currency' or 'country_id'", http.StatusBadRequest)
			return
		}

		groupByColumn := ""
		if req.GroupBy == "opening_balance_currency" {
			groupByColumn = "opening_balance_currency"
		} else if req.GroupBy == "country_id" {
			groupByColumn = "country_id"
		}

		query := `
			WITH latest_records AS (
				SELECT DISTINCT ON (account_no) *
				FROM account_balance
				ORDER BY account_no, account_balance_date DESC
			)
			SELECT ` + groupByColumn + `, account_no,
				opening_balance_amount,
				opening_balance_cdtdbtind,
				closing_balance_amount,
				closing_balance_cdtdbtind
			FROM latest_records
			GROUP BY ` + groupByColumn + `, account_no,
				opening_balance_amount,
				opening_balance_cdtdbtind,
				closing_balance_amount,
				closing_balance_cdtdbtind`

		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, "Database query failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var result []AccountBalance
		for rows.Next() {
			var acc AccountBalance
			var openingBalanceIndicator, closingBalanceIndicator string
			var openingBalanceAmount, closingBalanceAmount float64

			if err := rows.Scan(&acc.Currency, &acc.AccountNo, &openingBalanceAmount, &openingBalanceIndicator, &closingBalanceAmount, &closingBalanceIndicator); err != nil {
				http.Error(w, "Row scan failed: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if openingBalanceIndicator == "DBIT" {
				openingBalanceAmount = -math.Abs(openingBalanceAmount)
			}
			if closingBalanceIndicator == "DBIT" {
				closingBalanceAmount = -math.Abs(closingBalanceAmount)
			}

			acc.OpeningBalance = openingBalanceAmount
			acc.ClosingBalance = closingBalanceAmount

			if closingBalanceAmount != 0 {
				acc.PercentageChange = ((closingBalanceAmount - openingBalanceAmount) / math.Abs(closingBalanceAmount)) * 100
			}

			result = append(result, acc)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Error iterating rows: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(result)
	}).Methods("POST")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         86400,
	})

	srv := &http.Server{
		Handler:      c.Handler(r),
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Println("Server starting on :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
