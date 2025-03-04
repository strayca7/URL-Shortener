package main

import (
	"url-shortener/internal/pkg/database"
)

func main() {
	database.InitDB()
	defer database.CloseDB()
}