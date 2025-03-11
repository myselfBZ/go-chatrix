package main

import (
	"context"

	"github.com/coder/websocket/wsjson"
	"github.com/myselfBZ/chatrix/internal/messaging"
)


func (s *Server) peerMsgLoop() {
    for msg := range s.peerMsgChan {
        s.forwardToClient(msg.To, msg.Msg)
	}
}

func (s *Server) forwardToClient(to string, m *messaging.ServerMessage) {
    client := s.pool.Get(to)
    if client != nil{
        wsjson.Write(context.TODO(), client.Conn, m)
    }
}
