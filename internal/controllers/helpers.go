package controllers

import "strconv"

// parseInt parses a string to int, returns 0 if invalid
func parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
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
