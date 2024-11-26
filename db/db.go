package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"trading_app_cli/models"

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
	// type: 'buy' or 'sell'
	// status: 'open' or 'matched'
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
			status VARCHAR(10) DEFAULT 'open',
			created_at TIMESTAMPTZ DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS trades (
			id SERIAL PRIMARY KEY,
			buy_order_id INT NOT NULL,
			sell_order_id INT NOT NULL,
			amount DECIMAL NOT NULL,
			price DECIMAL NOT NULL,
			status VARCHAR(10) DEFAULT 'open',
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

func CreateOrder(user, orderType string, amount, price float64) (models.Order, error) {
	newOrder := models.Order{
		User:   user,
		Type:   orderType,
		Amount: amount,
		Price:  price,
	}
	query := "INSERT INTO orders (user_id, type, amount, price, status) VALUES ($1, $2, $3, $4, 'open') RETURNING id"
	err := DB.QueryRow(query, user, orderType, amount, price).Scan(&newOrder.ID)
	if err != nil {
		return models.Order{}, err
	}
	return newOrder, nil
}
func ShowOrders() {
	// Buyオーダーを取得
	buyRows, err := DB.Query(`
		SELECT user_id, amount, price, created_at
		FROM orders
		WHERE type = 'buy' AND status = 'open'
		ORDER BY created_at ASC
	`)
	if err != nil {
		fmt.Printf("Failed to fetch buy orders: %v\n", err)
		return
	}
	defer buyRows.Close()

	fmt.Println("Open Buy Orders:")
	for buyRows.Next() {
		var userID string
		var amount, price float64
		var createdAt string
		err := buyRows.Scan(&userID, &amount, &price, &createdAt)
		if err != nil {
			fmt.Printf("Failed to scan buy order: %v\n", err)
			continue
		}
		fmt.Printf("User: %s | Amount: %.2f | Price: %.2f | Date: %s\n", userID, amount, price, createdAt)
	}

	// Sellオーダーを取得
	sellRows, err := DB.Query(`
		SELECT user_id, amount, price, created_at
		FROM orders
		WHERE type = 'sell' AND status = 'open'
		ORDER BY created_at ASC
	`)
	if err != nil {
		fmt.Printf("Failed to fetch sell orders: %v\n", err)
		return
	}
	defer sellRows.Close()

	fmt.Println("\nOpen Sell Orders:")
	for sellRows.Next() {
		var userID string
		var amount, price float64
		var createdAt string
		err := sellRows.Scan(&userID, &amount, &price, &createdAt)
		if err != nil {
			fmt.Printf("Failed to scan sell order: %v\n", err)
			continue
		}
		fmt.Printf("User: %s | Amount: %.2f | Price: %.2f | Date: %s\n", userID, amount, price, createdAt)
	}
}

// func ShowOrders() {
// 	// Buyオーダーを取得
// 	rows, err := DB.Query(`
// 		SELECT user_id, amount, price, created_at
// 		FROM orders
// 		WHERE type = 'buy' AND status = 'open'
// 		ORDER BY created_at ASC
// 	`)
// 	if err != nil {
// 		fmt.Printf("Failed to fetch buy orders: %v\n", err)
// 		return
// 	}
// 	defer rows.Close()

// 	fmt.Println("Buy Orders:")
// 	for rows.Next() {
// 		var userID string
// 		var amount, price float64
// 		rows.Scan(&userID, &amount, &price)
// 		fmt.Printf("User: %s, Amount: %.2f, Price: %.2f\n", userID, amount, price)
// 	}

// 	// Sellオーダーを取得
// 	rows, err = DB.Query(`
// 		SELECT user_id, amount, price, created_at
// 		FROM orders
// 		WHERE type = 'sell' AND status = 'open'
// 		ORDER BY created_at ASC
// 	`)
// 	if err != nil {
// 		fmt.Printf("Failed to fetch sell orders: %v\n", err)
// 		return
// 	}
// 	defer rows.Close()

// 	fmt.Println("Sell Orders:")
// 	for rows.Next() {
// 		var userID string
// 		var amount, price float64
// 		rows.Scan(&userID, &amount, &price)
// 		fmt.Printf("User: %s, Amount: %.2f, Price: %.2f\n", userID, amount, price)
// 	}
// }

// ResetTables drops and recreates the orders table
func DropTables() {
	// テーブルを削除
	_, err1 := DB.Exec("DROP TABLE IF EXISTS users")
	if err1 != nil {
		log.Fatalf("Failed to drop table: %v", err1)
	}
	_, err2 := DB.Exec("DROP TABLE IF EXISTS orders")
	if err2 != nil {
		log.Fatalf("Failed to drop table: %v", err2)
	}
	_, err3 := DB.Exec("DROP TABLE IF EXISTS trades")
	if err3 != nil {
		log.Fatalf("Failed to drop table: %v", err3)
	}
	log.Println("Tables reset successfully")
}
