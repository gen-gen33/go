package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatalf("DB_URL is not set in .env file")
	}

	// Initialize DB connection
	var dbErr error
	DB, dbErr = sql.Open("postgres", dbURL)
	if dbErr != nil {
		log.Fatalf("Failed to connect to CockroachDB: %v", dbErr)
	}

	// Ping the database
	if pingErr := DB.Ping(); pingErr != nil {
		log.Fatalf("Failed to ping database: %v", pingErr)
	}

	fmt.Println("Connected to CockroachDB!")
}

func SetupTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name STRING UNIQUE NOT NULL,
			balance DECIMAL NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			user_id STRING NOT NULL,
			type STRING NOT NULL,
			amount DECIMAL NOT NULL,
			price DECIMAL NOT NULL,
			created_at TIMESTAMPTZ DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS trades (
			id SERIAL PRIMARY KEY,
			buy_order_id INT NOT NULL,
			sell_order_id INT NOT NULL,
			amount DECIMAL NOT NULL,
			price DECIMAL NOT NULL,
			created_at TIMESTAMPTZ DEFAULT now()
		)`,
	}

	for _, query := range queries {
		_, execErr := DB.Exec(query)
		if execErr != nil {
			log.Fatalf("Failed to execute query: %v", execErr)
		}
	}

	fmt.Println("Tables created successfully!")
}

func CreateUser(name string, balance float64) {
	_, err := DB.Exec("INSERT INTO users (name, balance) VALUES ($1, $2)", name, balance)
	if err != nil {
		fmt.Printf("Failed to create user: %v ", err)
		return
	}
	fmt.Println("User created successfully!")
}

func CreateOrder(user, orderType string, amount, price float64) {
	_, err := DB.Exec("INSERT INTO orders (user_id, type, amount, price) VALUES ($1, $2, $3, $4)",
		user, orderType, amount, price)
	if err != nil {
		fmt.Printf("Failed to create order: %v ", err)
		return
	}
	fmt.Println("Order created successfully!")
}

func ShowOrders() {
	// Buyオーダーを取得
	rows, err := DB.Query("SELECT user_id, amount, price FROM orders WHERE type = 'buy'")
	if err != nil {
		fmt.Printf("Failed to fetch buy orders: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("Buy Orders:")
	for rows.Next() {
		var userID string
		var amount, price float64
		rows.Scan(&userID, &amount, &price)
		fmt.Printf("User: %s, Amount: %.2f, Price: %.2f\n", userID, amount, price)
	}

	// Sellオーダーを取得
	rows, err = DB.Query("SELECT user_id, amount, price FROM orders WHERE type = 'sell'")
	if err != nil {
		fmt.Printf("Failed to fetch sell orders: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("Sell Orders:")
	for rows.Next() {
		var userID string
		var amount, price float64
		rows.Scan(&userID, &amount, &price)
		fmt.Printf("User: %s, Amount: %.2f, Price: %.2f\n", userID, amount, price)
	}
}
