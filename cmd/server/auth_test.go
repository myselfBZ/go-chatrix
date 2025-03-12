package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestAuth(t *testing.T) {
    s := newTestServer()

    mux := s.registerRoutes()
    
    t.Run("create new user", func(t *testing.T) {
        payload, err := json.Marshal(&registerUserPayload{
            Username: "user123",
            Password: "supersecretpassword",
            Name: "do'mbra qo'zi",
        })

        if err != nil{
            t.Fatalf("couldn't marshal the payload: %v", err)
        }

        req, err := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(payload))

        if err != nil{
            t.Fatalf("couldn't create a request %v", err)
        }

        rr := executeRequest(req, mux)
        checkResponseCode(t, 200, rr.Result().StatusCode)
    })

    t.Run("login", func(t *testing.T) {
        payload, err := json.Marshal(&loginPayload{
            Username: "boburmirzo",
            Password: "supersecretpassword",
        })

        if err != nil{
            t.Fatalf("couldn't marshal the payload: %v", err)
        }

        req, err := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(payload))

        if err != nil{
            t.Fatalf("couldn't create a request %v", err)
        }
        rr := executeRequest(req, mux)
        checkResponseCode(t, 200, rr.Result().StatusCode)
        
    })
}
