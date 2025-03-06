package main

import (
	"context"
	"errors"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type ErrEnvelope struct{
    Error error `json:"error"`
}

var InternalServerError = &ErrEnvelope{Error: errors.New("server ecountered a problem")}


func wsWriteJSONError(ctx context.Context, conn *websocket.Conn, err error){
    wsjson.Write(ctx, conn, &ServerMessage{Type: ERR, Body: err})
}

func wsInvalidJSONPayload(ctx context.Context, conn *websocket.Conn) {
    wsWriteJSONError(ctx, conn, errors.New("invalid json payload"))
}

func WsServerError(ctx context.Context, conn *websocket.Conn) {
    wsWriteJSONError(ctx, conn, errors.New("server encountered a problem"))
}

func IsCloseErr(ctx context.Context, err error) bool{
    return errors.As(err, &websocket.CloseError{})
}
