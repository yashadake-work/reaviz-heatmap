package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

type Account struct {
	AccountNo string `json:"account_no"`
	Currency  string `json:"account_ccy"`
	Country   string `json:"account_country"`
}

type TreeMapData struct {
	Key  string `json:"key"`
	Data []struct {
		Key  string `json:"key"`
		Data int    `json:"data"`
	} `json:"data"`
}

type FilterRequest struct {
	Filter string `json:"filter"` // Either 'account_ccy' or 'account_country'
}

func main() {
	log.Println("Starting server initialization...")

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Database connection string - update with your credentials
	connStr := "host=localhost port=9003 user=postgres password=arthavedh123 dbname=statement_processing sslmode=disable"

	log.Println("Connecting to database...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Database ping failed:", err)
	}
	log.Println("Database connected successfully")

	r := mux.NewRouter()

	// API endpoint to fetch TreeMap data with POST request
	r.HandleFunc("/api/filterdata", func(w http.ResponseWriter, r *http.Request) {
		var req FilterRequest
		// Parse the JSON body to get the filter type (account_ccy or account_country)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON request", http.StatusBadRequest)
			return
		}

		// Validate the filter value
		if req.Filter != "account_ccy" && req.Filter != "account_country" {
			http.Error(w, "Invalid filter value. It must be either 'account_ccy' or 'account_country'", http.StatusBadRequest)
			return
		}

		// Build the query based on the selected filter
		var query string
		var rows *sql.Rows
		if req.Filter == "account_ccy" {
			// For account_ccy, group by account_ccy
			query = `SELECT account_ccy, account_no FROM accounts GROUP BY account_ccy, account_no`
			rows, err = db.Query(query)
		} else if req.Filter == "account_country" {
			// For account_country, group by account_country
			query = `SELECT account_country, account_no FROM accounts GROUP BY account_country, account_no`
			rows, err = db.Query(query)
		}

		if err != nil {
			http.Error(w, "Database query failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Prepare the data structure to hold grouped data
		dataMap := make(map[string][]string)
		for rows.Next() {
			var acc Account
			if req.Filter == "account_ccy" {
				if err := rows.Scan(&acc.Currency, &acc.AccountNo); err != nil {
					http.Error(w, "Row scan failed: "+err.Error(), http.StatusInternalServerError)
					return
				}
				dataMap[acc.Currency] = append(dataMap[acc.Currency], acc.AccountNo)
			} else if req.Filter == "account_country" {
				if err := rows.Scan(&acc.Country, &acc.AccountNo); err != nil {
					http.Error(w, "Row scan failed: "+err.Error(), http.StatusInternalServerError)
					return
				}
				dataMap[acc.Country] = append(dataMap[acc.Country], acc.AccountNo)
			}
		}

		// Prepare the final result structure
		var result []TreeMapData
		for key, accounts := range dataMap {
			var accountData []struct {
				Key  string `json:"key"`
				Data int    `json:"data"`
			}

			// Random value between 10-60 for each account
			for _, accNo := range accounts {
				accountData = append(accountData, struct {
					Key  string `json:"key"`
					Data int    `json:"data"`
				}{
					Key:  accNo,
					Data: rand.Intn(50) + 10, // Random value between 10-60
				})
			}

			// Add the grouped data to the final result
			result = append(result, TreeMapData{
				Key:  key,
				Data: accountData,
			})
		}

		// Send the response as JSON
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(result)
	}).Methods("POST")

	// CORS middleware to allow all origins
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         86400, // 24 hours for preflight cache
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
