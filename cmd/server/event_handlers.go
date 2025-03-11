package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"

	"github.com/myselfBZ/chatrix/internal/messaging"
)

func (s *Server) handleText(event *messaging.Event) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var t messaging.IncomingMessagePayload

	if err := json.Unmarshal([]byte(event.Body), &t); err != nil {
		client := s.pool.Get(event.From)
		if client != nil {
			wsInvalidJSONPayload(context.TODO(), client.Conn)
			cancel()
		}
	}

	t.From = event.From
	t.FromUserID = event.FromID

	chat, err := s.ensureChatExists(t.FromUserID, t.ToUserID)

	if err != nil {
		s.sendServerMessage(ctx, t.From, &messaging.ServerMessage{Type: messaging.ERR, Body: InternalServerError})
		return
	}

	id, err := s.storeMessage(&t, chat.ID)

	if err != nil {
		s.sendServerMessage(ctx, t.From, &messaging.ServerMessage{Type: messaging.ERR, Body: InternalServerError})
		return
	}
	s.sendMessage(ctx, id, &t)
}

func (s *Server) handleMarkRead(event *messaging.Event) {
    // else just handle it
    var p messaging.MarkReadRequestPayload
    if err := json.Unmarshal(event.Body, &p); err != nil {
    	client := s.pool.Get(event.From)
    	if client != nil {
            wsInvalidJSONPayload(context.TODO(), client.Conn)
    	}
    }
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
	err := s.store.Messages.MarkRead(p.MessageIds)
	if err != nil {
		log.Println("DEBUG: ", err)
        s.sendServerMessage(ctx, p.From, &messaging.ServerMessage{Type: messaging.ERR, Body: InternalServerError})
		return
	}
    
    s.sendServerMessage(ctx, p.To, &messaging.ServerMessage{Type: messaging.MARK_READ, Body: p.MessageIds})
}




func (s *Server) handleLoadChatHistory(p *messaging.LoadChatHistoryReqPayload, from string) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
	chat, err := s.store.Chats.GetByUsersID(p.User1ID, p.User2ID)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
            s.sendServerMessage(ctx, from, &messaging.ServerMessage{Type: messaging.LoadChatHistoryResponse, Body: nil})
			return
		}
        log.Println("SERVER ERROR:", err)
        s.sendServerMessage(ctx, from, &messaging.ServerMessage{Type: messaging.ERR, Body: InternalServerError})
		return
	}
	messages, err := s.store.Messages.GetByChatID(chat.ID)
	if err != nil {
        s.sendServerMessage(ctx, from, &messaging.ServerMessage{Type: messaging.ERR, Body: InternalServerError })
		return
	}
    s.sendServerMessage(ctx, from, &messaging.ServerMessage{Type: messaging.LoadChatHistoryResponse, Body: messages})
}

func (s *Server) handlePeerEvent(event *messaging.Event) {
    var outMsg messaging.OutGoingMessage

    if err := json.Unmarshal([]byte(event.Body), &outMsg); err != nil{
        if errors.Is(err, &json.SyntaxError{}){
            s.sendServerMessage(context.TODO(), event.From, &messaging.ServerMessage{Type: messaging.ERR, Body: ErrEnvelope{Error: err}})
            return
        }
        log.Println("DEBUG: ", err)
    }

    s.sendServerMessage(context.TODO(), outMsg.To, &messaging.ServerMessage{Type: messaging.TEXT, Body: outMsg})
}

func (s *Server) handleUserSearch(query *messaging.SearchUserPayload) {
    ctx := context.TODO()
	users, err := s.store.Users.SearchByUsername(query.Username)
	if err != nil {
        if errors.Is(err, sql.ErrNoRows){
            s.sendServerMessage(ctx, query.From, &messaging.ServerMessage{Type: messaging.ERR, Body: ErrEnvelope{ Error: errors.New("user not found") }})
            return
        }
		log.Println("DEBUG: ", err)
		return
	}
    s.sendServerMessage(ctx, query.From, &messaging.ServerMessage{Type: messaging.SearchUserResponse, Body: users})
}
