package engine

import (
	"database/sql"
	"fmt"
	"trading_app_cli/models"
)

var Orders []models.Order
var Trades []string

//	func ShowTrades() {
//		fmt.Println("Trade matching is not implemented in this simple CLI version.")
//	}
func ShowTrades(db *sql.DB) {
	rows, err := db.Query(`
		SELECT t.id, t.buy_order_id, b.user_id AS buyer, t.sell_order_id, s.user_id AS seller, t.amount, t.price, t.created_at
		FROM trades t
		JOIN orders b ON t.buy_order_id = b.id
		JOIN orders s ON t.sell_order_id = s.id
		ORDER BY t.created_at DESC
	`)
	if err != nil {
		fmt.Printf("Failed to fetch trades: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("Trades:")
	for rows.Next() {
		var tradeID int
		var buyOrderID, sellOrderID int
		var buyer, seller string
		var amount, price float64
		var createdAt string
		err := rows.Scan(&tradeID, &buyOrderID, &buyer, &sellOrderID, &seller, &amount, &price, &createdAt)
		if err != nil {
			fmt.Printf("Failed to scan trade: %v\n", err)
			continue
		}

		fmt.Printf("Trade ID: %d | Buyer: %s (Order ID: %d) | Seller: %s (Order ID: %d) | Amount: %.2f | Price: %.2f | Date: %s\n",
			tradeID, buyer, buyOrderID, seller, sellOrderID, amount, price, createdAt)
	}
}

// func MatchOrder(db *sql.DB, newOrder models.Order) (bool, string, error) {
// 	tx, err := db.Begin() // トランザクションを開始
// 	if err != nil {
// 		return false, "", err
// 	}

// 	defer tx.Rollback() // エラー時にロールバック

// 	// 新規注文を保存
// 	query := `INSERT INTO orders ("user_id", "type", "amount", "price", "status")
//           VALUES ($1, $2, $3, $4, 'open') RETURNING id`
// 	err = tx.QueryRow(query, newOrder.User, newOrder.Type, newOrder.Amount, newOrder.Price).Scan(&newOrder.ID)
// 	if err != nil {
// 		return false, "", err
// 	}

// 	// マッチング注文を検索
// 	matchQuery := `
// 		SELECT id, user, type, amount, price
// 		FROM orders
// 		WHERE status = 'open'
// 		  AND type = CASE WHEN $1 = 'buy' THEN 'sell' ELSE 'buy' END
// 		  AND price <= $2
// 		ORDER BY price ASC
// 		LIMIT 1;
// 	`
// 	var matchedOrder models.Order
// 	err = tx.QueryRow(matchQuery, newOrder.Type, newOrder.Price).Scan(&matchedOrder.ID, &matchedOrder.User, &matchedOrder.Type, &matchedOrder.Amount, &matchedOrder.Price)
// 	if err == sql.ErrNoRows {
// 		// マッチング注文なし
// 		return false, "Order added. No match found.", tx.Commit()
// 	} else if err != nil {
// 		return false, "", err
// 	}

// 	// マッチング成功：両方の注文を "matched" に更新
// 	updateQuery := "UPDATE orders SET status = 'matched' WHERE id = $1"
// 	_, err = tx.Exec(updateQuery, newOrder.ID)
// 	if err != nil {
// 		return false, "", err
// 	}

// 	_, err = tx.Exec(updateQuery, matchedOrder.ID)
// 	if err != nil {
// 		return false, "", err
// 	}

// 	// トランザクションをコミット
// 	err = tx.Commit()
// 	if err != nil {
// 		return false, "", err
// 	}

// 	// マッチング結果を返す
// 	message := fmt.Sprintf("Trade executed: %s %s %.2f units at %.2f with %s",
// 		newOrder.User, newOrder.Type, newOrder.Amount, matchedOrder.Price, matchedOrder.User)
// 	return true, message, nil
// }

func MatchOrder(db *sql.DB, newOrder models.Order) (bool, string, error) {
	tx, err := db.Begin() // トランザクションを開始
	if err != nil {
		return false, "", err
	}
	defer tx.Rollback() // エラー時にロールバック

	// マッチング注文を検索
	matchQuery := `
        SELECT id, user_id, type, amount, price
        FROM orders
        WHERE status = 'open'
        AND type = CASE WHEN $1 = 'buy' THEN 'sell' ELSE 'buy' END
        AND (
            (type = 'buy' AND price >= $2)
            OR
            (type = 'sell' AND price <= $2)
        )
        ORDER BY price ASC
        LIMIT 1;
    `
	var matchedOrder models.Order
	err = tx.QueryRow(matchQuery, newOrder.Type, newOrder.Price).Scan(
		&matchedOrder.ID,
		&matchedOrder.User,
		&matchedOrder.Type,
		&matchedOrder.Amount,
		&matchedOrder.Price)
	if err == sql.ErrNoRows {
		// マッチング注文なし
		return false, "Order added. No match found.", tx.Commit()
	} else if err != nil {
		return false, "", err
	}

	// 部分約定を計算
	tradeAmount := newOrder.Amount
	if newOrder.Amount > matchedOrder.Amount {
		tradeAmount = matchedOrder.Amount
	}

	// トレード履歴を保存
	tradeInsertQuery := `
		INSERT INTO trades (buy_order_id, sell_order_id, amount, price, created_at)
		VALUES ($1, $2, $3, $4, now())
	`
	if newOrder.Type == "buy" {
		_, err = tx.Exec(tradeInsertQuery, newOrder.ID, matchedOrder.ID, tradeAmount, matchedOrder.Price)
	} else {
		_, err = tx.Exec(tradeInsertQuery, matchedOrder.ID, newOrder.ID, tradeAmount, matchedOrder.Price)
	}
	if err != nil {
		return false, "", err
	}

	// 更新処理（注文量とステータス）
	if newOrder.Amount > matchedOrder.Amount {
		// 新規注文は部分約定
		updateQuery := "UPDATE orders SET amount = amount - $1 WHERE id = $2"
		_, err = tx.Exec(updateQuery, tradeAmount, newOrder.ID)
		if err != nil {
			return false, "", err
		}

		// 既存注文は完全約定
		_, err = tx.Exec("UPDATE orders SET status = 'matched' WHERE id = $1", matchedOrder.ID)
		if err != nil {
			return false, "", err
		}
	} else if newOrder.Amount < matchedOrder.Amount {
		// 既存注文は部分約定
		updateQuery := "UPDATE orders SET amount = amount - $1 WHERE id = $2"
		_, err = tx.Exec(updateQuery, tradeAmount, matchedOrder.ID)
		if err != nil {
			return false, "", err
		}

		// 新規注文は完全約定
		_, err = tx.Exec("UPDATE orders SET status = 'matched' WHERE id = $1", newOrder.ID)
		if err != nil {
			return false, "", err
		}
	} else {
		// 両方完全約定
		_, err = tx.Exec("UPDATE orders SET status = 'matched' WHERE id = $1 OR id = $2", newOrder.ID, matchedOrder.ID)
		if err != nil {
			return false, "", err
		}
	}

	// トランザクションをコミット
	err = tx.Commit()
	if err != nil {
		return false, "", err
	}

	// マッチング結果を返す
	message := fmt.Sprintf("Trade executed: %s %s %.2f units at %.2f with %s",
		newOrder.User, newOrder.Type, tradeAmount, matchedOrder.Price, matchedOrder.User)
	return true, message, nil
}
