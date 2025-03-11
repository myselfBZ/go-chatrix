package main

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

func ensureEnvExists(key string) string{
    value := os.Getenv(key)
    if value != ""{
        return value
    }
    log.Fatalf("%s wasn't set!", key)
    return ""
}


func main() {
    var (
        server_port = ensureEnvExists("SERVER_PORT")
        serverHost = ensureEnvExists("SERVERHOST")
        // db
        dbhost        = ensureEnvExists("DB_HOST")
        port        = ensureEnvExists("DB_PORT")
        user        = ensureEnvExists("DB_USER")
        password    = ensureEnvExists("DB_PASSWORD")
        db_name     = ensureEnvExists("DB_NAME")
        // redis 
        redisAddr = ensureEnvExists("REDISADDR")
    )

	config := Config{}
    // a very special line
	config.ListenAddr = server_port
    config.ServerAddr = serverHost
	config.Db = dbConfig{
		Addr: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, dbhost, port, db_name),
	}

    config.Redis = redisConfig{
        addr: redisAddr,
        listenChannel: config.ServerAddr,
    }


    config.WorkerPool = 5

	config.Auth = authConfig{
		Secret: "secret",
		Exp:    time.Hour * 24,
		Iss:    "some random ass thing",
	}

	s := NewServer(config)
    log.Println("Server started successfully")
	s.Run()
}
