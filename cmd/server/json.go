package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func WSreadJSONWithErr(conn *websocket.Conn, data []byte, d interface{} ) error {
    if err := json.Unmarshal(data, d); err != nil{
        return wsjson.Write(context.TODO(), conn, ErrEnvelope{Error: ErrInvalidJsonPayload})
    }
    return nil
}


func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_578 // 1mb
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

func writeJSONError(w http.ResponseWriter, status int, err string) {
	type envelope struct {
		Error string `json:"error"`
	}
	writeJSON(w, status, &envelope{Error: err})
}
