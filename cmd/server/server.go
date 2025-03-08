package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/joho/godotenv/autoload"
	"github.com/myselfBZ/chatrix/internal/auth"
	"github.com/myselfBZ/chatrix/internal/db"
	"github.com/myselfBZ/chatrix/internal/store"
)

type Config struct {
	Addr string
	Db   dbConfig
	auth authConfig
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

	wsConns   sync.Map
	eventChan chan Event
}

func failOnError(msg string, err error) {
	if err != nil {
		log.Fatal(msg, err)
	}
}

func NewServer(config Config) *Server {
	db, err := db.New(config.Db.Addr, 30, 30, "15m")
	failOnError("db", err)
	return &Server{
		store:   store.New(db),
		Config:  config,
		wsConns: sync.Map{},
		auth: &auth.JWTAuthenticator{
			Secret: config.auth.Secret,
			Iss:    config.auth.Iss,
			Aud:    config.auth.Iss,
		},
		// arbitary number of chans
		eventChan: make(chan Event, 100),
	}
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
	return http.ListenAndServe(s.Config.Addr, s.registerRoutes())
}
