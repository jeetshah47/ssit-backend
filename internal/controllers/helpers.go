package controllers

import "strconv"

// parseInt parses a string to int, returns 0 if invalid
func parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}
