package main

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/myselfBZ/chatrix/internal/store"
)

type tokenEnvelope struct {
	Token string `json:"token"`
}

type registerUserPayload struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Name     string `json:"name" validate:"required"`
}

func (s *Server) registerUser(w http.ResponseWriter, r *http.Request) {
	var payload registerUserPayload

	if err := readJSON(w, r, &payload); err != nil {
		s.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		s.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Name:     payload.Name,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		s.internalServerError(w, r, err)
		return
	}

	if err := s.store.Users.Create(user); err != nil {
		s.internalServerError(w, r, err)
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(s.Config.auth.Exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": s.Config.auth.Iss,
		"aud": s.Config.auth.Iss,
	}

	token, err := s.auth.GenerateToken(claims)
	if err != nil {
		s.internalServerError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, tokenEnvelope{Token: token})

}

type loginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var payload loginPayload
	if err := readJSON(w, r, &payload); err != nil {
		s.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(&payload); err != nil {
		s.badRequestResponse(w, r, err)
		return
	}

	user, err := s.store.Users.GetByUsername(payload.Username)
	if err != nil {
		s.notFoundResponse(w, r, err)
		return
	}

	if err := user.Password.Compare(payload.Password); err != nil {
		s.unauthorizedErrorResponse(w, r, err)
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(s.Config.auth.Exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": s.Config.auth.Iss,
		"aud": s.Config.auth.Iss,
	}

	token, err := s.auth.GenerateToken(claims)
	if err != nil {
		s.internalServerError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, tokenEnvelope{Token: token})

}
