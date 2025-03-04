package main

import (
	"fmt"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

var(
    host = os.Getenv("DB_HOST")
    port = os.Getenv("DB_PORT")
    user = os.Getenv("DB_USER")
    password = os.Getenv("DB_PASSWORD")
    db_name = os.Getenv("DB_NAME")
)



func main(){
    config := Config{}
    config.Addr = ":6969"
    config.Db = dbConfig{
        Addr: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, db_name),
    }
    config.auth = authConfig{
        Secret: "secret",
        Exp: time.Hour * 24,
        Iss: "some random ass thing",
    }

    s := NewServer(config)
    s.Run()
}

