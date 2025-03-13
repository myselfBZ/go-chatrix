package main

import (
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
        dbConnUrl = ensureEnvExists("DB_CONNECTION_URL")
        redisAddr = ensureEnvExists("REDISADDR")
    )

	config := Config{}
	config.ListenAddr = server_port
    config.ServerAddr = serverHost
	config.Db = dbConfig{
		Addr: dbConnUrl,
	}

    config.Redis = redisConfig{
        addr: redisAddr,
        listenChannel: config.ServerAddr,
    }


    config.WorkerPool = 5

	config.Auth = authConfig{
		Secret: "secret",
		Exp:    time.Hour * 24,
		Iss:    "chatrix-server",
	}

	s := NewServer(config)
    log.Println("Server started successfully")
	s.Run()
}
