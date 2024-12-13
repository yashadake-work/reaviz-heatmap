package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

type TreeMapData struct {
	Key  string `json:"key"`
	Data []struct {
		Key  string  `json:"key"`
		Data float64 `json:"data"`
	} `json:"data"`
}

type GroupByRequest struct {
	GroupBy string `json:"groupby"` // "opening_balance_currency" or "country_id"
}

func main() {
	log.Println("Starting server initialization...")

	// Database connection string - update with your credentials
	connStr := "host=localhost port=9003 user=postgres password=arthavedh123 dbname=statement_processing sslmode=disable"

	log.Println("Connecting to database...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	r := mux.NewRouter()

	// API endpoint to fetch grouped data
	r.HandleFunc("/heatmap/filterdata", func(w http.ResponseWriter, r *http.Request) {
		var req GroupByRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON request", http.StatusBadRequest)
			return
		}

		if req.GroupBy != "opening_balance_currency" && req.GroupBy != "country_id" {
			http.Error(w, "Invalid groupby value. It must be either 'opening_balance_currency' or 'country_id'", http.StatusBadRequest)
			return
		}

		query := `
			SELECT DISTINCT ON (account_no) 
				account_no, opening_balance_currency, country_id, 
				opening_balance_amount, opening_balance_cdtdbtind, 
				closing_balance_amount, closing_balance_cdtdbtind
			FROM account_balance
			ORDER BY account_no, account_balance_date DESC
		`

		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, "Database query failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		dataMap := make(map[string]map[string]float64) // groupby_value -> account_no -> percentage_change

		for rows.Next() {
			var (
				accountNo              string
				openingBalanceCurrency string
				countryID              string
				openingBalanceAmount   float64
				openingCdtDbtInd       string
				closingBalanceAmount   float64
				closingCdtDbtInd       string
			)

			if err := rows.Scan(&accountNo, &openingBalanceCurrency, &countryID, &openingBalanceAmount, &openingCdtDbtInd, &closingBalanceAmount, &closingCdtDbtInd); err != nil {
				http.Error(w, "Row scan failed: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Adjust balance amounts based on credit or debit indicators
			if openingCdtDbtInd == "DBIT" {
				openingBalanceAmount = -openingBalanceAmount
			}
			if closingCdtDbtInd == "DBIT" {
				closingBalanceAmount = -closingBalanceAmount
			}

			// Calculate percentage change
			percentChange := ((closingBalanceAmount - openingBalanceAmount) / closingBalanceAmount) * 100

			// Group by the specified key
			groupByKey := ""
			if req.GroupBy == "opening_balance_currency" {
				groupByKey = openingBalanceCurrency
			} else {
				groupByKey = countryID
			}

			if _, exists := dataMap[groupByKey]; !exists {
				dataMap[groupByKey] = make(map[string]float64)
			}
			dataMap[groupByKey][accountNo] = percentChange
		}

		// Prepare the final result
		var result []TreeMapData
		for key, accounts := range dataMap {
			var accountData []struct {
				Key  string  `json:"key"`
				Data float64 `json:"data"`
			}

			for accNo, percentChange := range accounts {
				accountData = append(accountData, struct {
					Key  string  `json:"key"`
					Data float64 `json:"data"`
				}{
					Key:  accNo,
					Data: percentChange,
				})
			}

			result = append(result, TreeMapData{
				Key:  key,
				Data: accountData,
			})
		}

		// Send the response
		w.Header().Set("Content-Type", "application/json")
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
