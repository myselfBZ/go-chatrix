package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/coder/websocket/wsjson"
	"github.com/go-redis/redis/v8"
	"github.com/myselfBZ/chatrix/internal/messaging"
)

func (s *Server) redisPubSubHandler(msg *redis.Message) {
    var m messaging.PeerMessage
    if err := json.Unmarshal([]byte(msg.Payload), &m); err != nil{
        log.Println("MARSHALING ERROR: ", err)
        return
    }
    s.peerMsgChan <- &m
}


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
