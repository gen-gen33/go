package main

import (
	"fmt"
	"trading_app_cli/db"
	"trading_app_cli/engine"
)

func main() {

	// Initialize the database connection
	db.InitDB()
	defer db.DB.Close()

	// Setup database tables
	db.SetupTables()

	fmt.Println("Welcome to the CLI Trading App!")
	fmt.Println("Commands:")
	fmt.Println(" - create user <name> <balance>")
	fmt.Println(" - login <name>")
	fmt.Println(" - buy <amount> <price>")
	fmt.Println(" - sell <amount> <price>")
	fmt.Println(" - orders")
	fmt.Println(" - trades")
	fmt.Println(" - logout")
	fmt.Println(" - help")
	fmt.Println(" - exit")

	var loggedInUser string

	for {
		fmt.Print("> ")
		var command string
		fmt.Scan(&command)

		switch command {
		case "create":
			var name string
			var balance float64
			fmt.Scan(&name, &balance)
			db.CreateUser(name, balance)

		case "login":
			var name string
			fmt.Scan(&name)
			loggedInUser = name

		case "buy":
			if loggedInUser == "" {
				fmt.Println("You must log in first.")
				continue
			}
			var amount, price float64
			fmt.Scan(&amount, &price)
			db.CreateOrder(loggedInUser, "buy", amount, price)

		case "sell":
			if loggedInUser == "" {
				fmt.Println("You must log in first.")
				continue
			}
			var amount, price float64
			fmt.Scan(&amount, &price)
			db.CreateOrder(loggedInUser, "sell", amount, price)

		case "orders":
			db.ShowOrders()

		case "trades":
			engine.ShowTrades()

		case "logout":
			loggedInUser = ""

		case "exit":
			fmt.Println("Goodbye!")
			return

		case "help":
			fmt.Println("Welcome to the CLI Trading App!")
			fmt.Println("Commands:")
			fmt.Println(" - create user <name> <balance>")
			fmt.Println(" - login <name>")
			fmt.Println(" - buy <amount> <price>")
			fmt.Println(" - sell <amount> <price>")
			fmt.Println(" - orders")
			fmt.Println(" - trades")
			fmt.Println(" - logout")
			fmt.Println(" - help")
			fmt.Println(" - exit")

		default:
			fmt.Println("Invalid command.")
		}
	}
}
