package main

import (
	"fmt"
	"log"
	"os"

    _ "github.com/joho/godotenv/autoload"
	"github.com/myselfBZ/chatrix/internal/db"
)

var(
    host = os.Getenv("DB_HOST")
    port = os.Getenv("DB_PORT")
    user = os.Getenv("DB_USER")
    password = os.Getenv("DB_PASSWORD")
    db_name = os.Getenv("DB_NAME")
)

// "postgres://postgres:new_password@localhost:32768/wonderlust?sslmode=disable", 30, 30, "15m"
func main() {
    if len(os.Args) != 2{
        log.Fatal("usage: migrate <path_to_schema>")
    }

    schema, err := os.ReadFile(os.Args[1])
    if err != nil{
        log.Fatal("couldn't open the schema", err)
    }

    addr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, db_name)
	db, err := db.New(addr, 30, 30, "15m")
    if err != nil{
        log.Fatal("couldn't connect to database", err)
    }

    _, err = db.Exec(string(schema))
    
    if err != nil{
        log.Fatal("couldn't migrate ", err)
    }
}
