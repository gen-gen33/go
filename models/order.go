package models

type Order struct {
	ID     int
	User   string
	Type   string // "buy" or "sell"
	Amount float64
	Price  float64
}
