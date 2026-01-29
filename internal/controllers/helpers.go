package controllers

import (
	"strconv"

	"github.com/equitywala/backend/internal/models"
)

// parseInt parses a string to int, returns 0 if invalid
func parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

// getExchangeFromStock returns display exchange from stock (exchange, symbol, or default)
func getExchangeFromStock(stock *models.Stock) string {
	if stock == nil {
		return "NSE"
	}
	if stock.Exchange != nil && *stock.Exchange != "" {
		return *stock.Exchange
	}
	if stock.Symbol != nil && *stock.Symbol != "" {
		return *stock.Symbol
	}
	return "NSE"
}

// getVerdictFromAction converts action string to verdict
func getVerdictFromAction(action *string) string {
	if action == nil {
		return "Hold"
	}
	switch *action {
	case "buy":
		return "Buy"
	case "sell":
		return "Sell"
	default:
		return "Hold"
	}
}
