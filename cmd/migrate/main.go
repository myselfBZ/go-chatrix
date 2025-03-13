package main

import (
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/myselfBZ/chatrix/internal/db"
)
func ensureEnvExists(key string) string{
    value := os.Getenv(key)
    if value != ""{
        return value
    }
    log.Fatalf("%s wasn't set!", key)
    return ""
}

var (
    connUrl = ensureEnvExists("DB_CONNECTION_URL")
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: migrate <path_to_schema>")
	}

	schema, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal("couldn't open the schema", err)
	}

	db, err := db.New(connUrl, 30, 30, "15m")
	if err != nil {
		log.Fatal("couldn't connect to database", err)
	}

	_, err = db.Exec(string(schema))

	if err != nil {
		log.Fatal("couldn't migrate ", err)
	}
}
