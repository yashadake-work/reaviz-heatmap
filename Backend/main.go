// main.go
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"

	"github.com/gorilla/mux"
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
	// Database connection
	db, err := sql.Open("postgres", "postgres://username:password@localhost:5432/dbname?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/api/treedata", func(w http.ResponseWriter, r *http.Request) {
		// Fetch accounts
		rows, err := db.Query("SELECT account_no, account_ccy FROM accounts")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Group by currency
		currencyMap := make(map[string][]string)
		for rows.Next() {
			var acc Account
			if err := rows.Scan(&acc.AccountNo, &acc.Currency); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			currencyMap[acc.Currency] = append(currencyMap[acc.Currency], acc.AccountNo)
		}

		// Format response
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
		json.NewEncoder(w).Encode(result)
	})

	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET"},
	})

	log.Fatal(http.ListenAndServe(":8080", c.Handler(r)))
}
