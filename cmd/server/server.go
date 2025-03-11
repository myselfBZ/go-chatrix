package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-redis/redis/v8"
	_ "github.com/joho/godotenv/autoload"
	"github.com/myselfBZ/chatrix/internal/auth"
	"github.com/myselfBZ/chatrix/internal/db"
	"github.com/myselfBZ/chatrix/internal/distribution"
	"github.com/myselfBZ/chatrix/internal/messaging"
	"github.com/myselfBZ/chatrix/internal/store"
)

type Config struct {
    ServerAddr string
	ListenAddr string
    redis  redisConfig 
	Db   dbConfig
	auth authConfig
}

type redisConfig struct{
    addr string
    listenChannel string
}

type authConfig struct {
	Secret string
	Exp    time.Duration
	Iss    string
}

type dbConfig struct {
	Addr        string
	MaxConn     int
	MaxIdleConn int
	MaxIdleTime string
}

type Server struct {
	store *store.Store

	Config Config
	auth   auth.Authenticator

    pool      *messaging.Pool
    pubsub    *distribution.PubSub

	eventChan chan *messaging.Event
    peerMsgChan chan *messaging.PeerMessage
}

func failOnError(msg string, err error) {
	if err != nil {
		log.Fatal(msg, err)
	}
}

func NewServer(config Config) *Server {
    server := &Server{}
    server.Config = config

	db, err := db.New(config.Db.Addr, 30, 30, "15m")
	failOnError("db", err)

    server.store = store.New(db)
    server.auth = &auth.JWTAuthenticator{
        Secret: config.auth.Secret,
        Iss: config.auth.Iss,
        Aud: config.auth.Iss,
    }

    server.eventChan = make(chan *messaging.Event, 100)
    server.peerMsgChan = make(chan *messaging.PeerMessage, 100)

    redisClient := redis.NewClient(&redis.Options{
        Addr: config.redis.addr,
    })

    server.pubsub = distribution.NewPubSub(redisClient, server.redisPubSubHandler, config.ServerAddr)

    server.pool = messaging.NewPool(config.ServerAddr, redisClient)

    return server
}

func (s *Server) registerRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		// todo change this thing
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Route("/auth", func(r chi.Router) {
		// TODO
		r.Post("/register", s.registerUser)
		r.Post("/login", s.login)
	})

	r.Route("/ws", func(r chi.Router) {
		r.HandleFunc("/", s.accept)
	})

    r.Get("/health", func (w http.ResponseWriter, r *http.Request)  {
        w.Write([]byte("healthy af"))
    })

	return r

}

// for handling messages coming from redis pub/sub
func (s *Server) redisPubSubHandler(msg *redis.Message) {
    var m messaging.PeerMessage
    if err := json.Unmarshal([]byte(msg.Payload), &m); err != nil{
        log.Println("MARSHALING ERROR: ", err)
        return
    }
    s.peerMsgChan <- &m
}

func (s *Server) Run() error {
    s.pubsub.Start()
    
    go s.peerMsgLoop()
	go s.eventLoop()
	return http.ListenAndServe(s.Config.ListenAddr, s.registerRoutes())
}
