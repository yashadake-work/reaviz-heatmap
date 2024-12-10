// main.go
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
}

type TreeMapData struct {
	Key  string `json:"key"`
	Data []struct {
		Key  string `json:"key"`
		Data int    `json:"data"`
	} `json:"data"`
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

	// API endpoint to fetch TreeMap data
	r.HandleFunc("/api/treedata", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT account_no, account_ccy FROM accounts")
		if err != nil {
			http.Error(w, "Database query failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		currencyMap := make(map[string][]string)
		for rows.Next() {
			var acc Account
			if err := rows.Scan(&acc.AccountNo, &acc.Currency); err != nil {
				http.Error(w, "Row scan failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			currencyMap[acc.Currency] = append(currencyMap[acc.Currency], acc.AccountNo)
		}

		var result []TreeMapData
		for currency, accounts := range currencyMap {
			var accountData []struct {
				Key  string `json:"key"`
				Data int    `json:"data"`
			}

			for _, accNo := range accounts {
				accountData = append(accountData, struct {
					Key  string `json:"key"`
					Data int    `json:"data"`
				}{
					Key:  accNo,
					Data: rand.Intn(50) + 10, // Random value between 10-60
				})
			}

			result = append(result, TreeMapData{
				Key:  currency,
				Data: accountData,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(result)
	})

	// CORS middleware to allow all origins
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "OPTIONS"},
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
