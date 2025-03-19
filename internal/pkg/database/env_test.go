package database

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestEnv(t *testing.T) {
	// fmt.Println(os.Getwd())
	if err := godotenv.Load("../.env"); err != nil {
		t.Fatalf("Failed to load .env file: %v", err)
	}
	var (
		mydbUser     = os.Getenv("MYSQL_USER")
		mydbPassword = os.Getenv("MYSQL_PASSWORD")
		mydbHost     = os.Getenv("MYSQL_HOST")
		mydbPort     = os.Getenv("MYSQL_PORT")
		mydbName     = os.Getenv("MYSQL_DATABASE")

		pgdbHost     = os.Getenv("PGSQL_HOST")
		pgdbUser     = os.Getenv("PGSQL_USER")
		pgdbPassword = os.Getenv("PGSQL_PASSWORD")
		pgdbName     = os.Getenv("PGSQL_DBNAME")
		pgdbPort     = os.Getenv("PGSQL_PORT")
		pgdbSSLMode  = os.Getenv("PGSQL_SSLMODE")
		pgdbTimeZone = os.Getenv("PGSQL_TIMEZONE")
	)
	fmt.Println("MySQL:", mydbUser, mydbPassword, mydbHost, mydbPort, mydbName)
	fmt.Println("PostgreSQL:", pgdbHost, pgdbUser, pgdbPassword, pgdbName, pgdbPort, pgdbSSLMode, pgdbTimeZone)
}

func TestMysql(t *testing.T) {
	InitMysqlDB()
	defer CloseMysqlDB()
}
