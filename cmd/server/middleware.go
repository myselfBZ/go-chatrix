package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/golang-jwt/jwt/v5"
	"github.com/myselfBZ/chatrix/internal/events"
	"github.com/myselfBZ/chatrix/internal/store"
)

const (
	UserKey = "user"
)

func (app *Server) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}

		token := parts[1]
		jwtToken, err := app.auth.ValidateToken(token)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		userID, err := strconv.Atoi(fmt.Sprintf("%.f", claims["sub"]))
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.store.Users.GetByID(userID)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}




func (s *Server) webSocketAuth(ctx context.Context, conn *websocket.Conn) (user *store.User, status websocket.StatusCode, err error) {

    defer func(){
        if err != nil{
            conn.Close(status, err.Error())
        }
    }()

	type envelope struct {
		Token string `json:"token"`
	}

	var tok envelope
	if err := wsjson.Read(ctx, conn, &tok); err != nil {
        wsInvalidJSONPayload(ctx, conn)
		return nil, websocket.StatusInvalidFramePayloadData, ErrInvalidJsonPayload
	}

	jwtToken, err := s.auth.ValidateToken(tok.Token)

	if err != nil {
		wsjson.Write(ctx, conn, events.ServerMessage{Type: events.ERR, Body: ErrEnvelope{Error: ErrInvalidToken}})
        return nil, websocket.StatusPolicyViolation, ErrInvalidToken
	}

	claims, _ := jwtToken.Claims.(jwt.MapClaims)

	userID, err := strconv.Atoi(fmt.Sprintf("%.f", claims["sub"]))

	if err != nil {
		return nil, websocket.StatusPolicyViolation, ErrInvalidUserId
	}

	user, err = s.store.Users.GetByID(userID)

	if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            wsjson.Write(ctx, conn, events.ServerMessage{Type: events.ERR, Body: ErrEnvelope{Error: ErrUserNotFound}})
            return nil, websocket.StatusPolicyViolation, ErrUserNotFound
        } 

        log.Println("DEBUG", err.Error())
        return nil, websocket.StatusInternalError, ErrInteralServer
	}

    return user, 0, nil
}




