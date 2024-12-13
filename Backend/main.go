package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
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

type FilterRequest struct {
	GroupBy string `json:"groupby"`
}

func maskAccountNo(accountNo string) string {
	if len(accountNo) < 6 {
		return accountNo
	}
	return fmt.Sprintf("%s***%s", accountNo[:4], accountNo[len(accountNo)-2:])
}

func roundToThreeDecimals(value float64) float64 {
	return math.Round(value*1000) / 1000
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
			http.Error(w, "Invalid groupby value. Must be 'opening_balance_currency' or 'country_id'", http.StatusBadRequest)
			return
		}

		query := `
			WITH latest_records AS (
				SELECT DISTINCT ON (account_no) *
				FROM account_balance
				ORDER BY account_no, account_balance_date DESC
			)
			SELECT ` + req.GroupBy + `, account_no, 
				opening_balance_amount, opening_balance_cdtdbtind,
				closing_balance_amount, closing_balance_cdtdbtind
			FROM latest_records
		`

		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, "Database query failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		dataMap := make(map[string][]struct {
			Key  string  `json:"key"`
			Data float64 `json:"data"`
		})

		for rows.Next() {
			var groupKey, accountNo, openingIndicator, closingIndicator string
			var openingBalance, closingBalance float64

			if err := rows.Scan(&groupKey, &accountNo, &openingBalance, &openingIndicator, &closingBalance, &closingIndicator); err != nil {
				http.Error(w, "Row scan failed: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if openingIndicator == "DBIT" {
				openingBalance = -openingBalance
			}
			if closingIndicator == "DBIT" {
				closingBalance = -closingBalance
			}

			percentChange := 0.0
			if closingBalance != 0 {
				percentChange = ((closingBalance - openingBalance) / closingBalance) * 100
			}
			percentChange = roundToThreeDecimals(percentChange)

			maskedAccountNo := maskAccountNo(accountNo)
			dataMap[groupKey] = append(dataMap[groupKey], struct {
				Key  string  `json:"key"`
				Data float64 `json:"data"`
			}{
				Key:  maskedAccountNo,
				Data: percentChange,
			})
		}

		var result []TreeMapData
		for key, accounts := range dataMap {
			result = append(result, TreeMapData{
				Key:  key,
				Data: accounts,
			})
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
