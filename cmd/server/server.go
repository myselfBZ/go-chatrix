package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-redis/redis/v8"
	_ "github.com/joho/godotenv/autoload"
	"github.com/myselfBZ/chatrix/internal/auth"
	"github.com/myselfBZ/chatrix/internal/db"
	"github.com/myselfBZ/chatrix/internal/events"
	"github.com/myselfBZ/chatrix/internal/kv"
	pubsub "github.com/myselfBZ/chatrix/internal/pub-sub"
	"github.com/myselfBZ/chatrix/internal/store"
)

type Config struct {
    FullAddr string
    IsDistributed bool
	Addr string
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
    kv *kv.KV
	store *store.Store

	Config Config
	auth   auth.Authenticator

    pubSub *pubsub.EventPubSub
	wsConns   sync.Map
	eventChan chan *events.Event
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
    server.wsConns = sync.Map{}
    server.auth = &auth.JWTAuthenticator{
        Secret: config.auth.Secret,
        Iss: config.auth.Iss,
        Aud: config.auth.Iss,
    }

    server.eventChan = make(chan *events.Event, 100)

    if config.IsDistributed{
        redisClient := redis.NewClient(&redis.Options{
            Addr: config.redis.addr,
        })

        server.pubSub = pubsub.New(redisClient, config.redis.listenChannel)
        server.kv = kv.New(redisClient)
    }

    return server

	// return &Server{
    //  kv: kv,
	// 	store:   store.New(db),
	// 	Config:  config,
	// 	wsConns: sync.Map{},
	// 	auth: &auth.JWTAuthenticator{
	// 		Secret: config.auth.Secret,
	// 		Iss:    config.auth.Iss,
	// 		Aud:    config.auth.Iss,
	// 	},
	//        pubSub: pubSub,
	// 	// arbitary number of chans
	// 	eventChan: make(chan *events.Event, 100),
	// }
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

func (s *Server) Run() error {
	go s.eventLoop()
    if s.Config.IsDistributed {
        go s.recieveFromPeer()
    }
	return http.ListenAndServe(s.Config.Addr, s.registerRoutes())
}
