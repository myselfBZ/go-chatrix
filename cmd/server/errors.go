package main

import (
	"log"
	"net/http"
)

func (s *Server) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (s *Server) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	log.Println("forbidden", "method", r.Method, "path", r.URL.Path, "error")

	writeJSONError(w, http.StatusForbidden, "forbidden")
}

func (s *Server) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (s *Server) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("conflict response", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusConflict, err.Error())
}

func (s *Server) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusNotFound, "not found")
}

func (s *Server) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (s *Server) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}

