package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-redis/redis/v8"
    "github.com/myselfBZ/chatrix/internal/messaging"
	"github.com/myselfBZ/chatrix/internal/store"
)



func (s *Server) sendInitialUserData(ctx context.Context, conn *websocket.Conn, user *store.User) error {
    ctxTimeout, cancel := context.WithTimeout(ctx, time.Second * 5)
    defer cancel()

	if err := wsjson.Write(ctxTimeout, conn, messaging.ServerMessage{Type: messaging.PROFILE_INFO, Body: user}); err != nil {
        if websocket.CloseStatus(err) == -1{
            conn.Close(websocket.StatusInternalError, "couldn't write json")
            return err
        }
        // client diconnected
        return err
	}

	chatPreviews, err := s.store.Chats.ChatPreviews(user.ID)

    if err == nil{
        if err := wsjson.Write(ctxTimeout, conn, messaging.ServerMessage{Type: messaging.CHATPREVIEWS, Body: chatPreviews}); err != nil {
            if websocket.CloseStatus(err) != -1{
                conn.Close(websocket.StatusInternalError, "client disconnected")
                return err
            }
            return err
        }
    }

    if err != nil {
        switch err{
        case sql.ErrNoRows:
            wsjson.Write(ctx, conn, messaging.ServerMessage{Type: messaging.CHATPREVIEWS, Body: nil})
        default:
            wsjson.Write(ctx, conn, messaging.ServerMessage{Type: messaging.ERR, Body: InternalServerError })
            log.Println("SERVER ERR: ", err)
        }
    }

    return nil
}

func (s *Server) handleWebSocketConn(ctx context.Context, conn *websocket.Conn) {
    user, _, err := s.webSocketAuth(ctx, conn)

    if err != nil{
        return
    }
	log.Printf("%s has just gone online", user.Username)

    err = s.sendInitialUserData(ctx, conn, user)

    if err != nil{
        return 
    }

    client := messaging.NewClient(user, conn)

    s.pool.Add(client)

    go s.readLoop(conn, user)
}

func (s *Server) readLoop(c *websocket.Conn, user *store.User) {
    ctx := context.Background()
	for {
		var event messaging.Event
		if err := wsjson.Read(ctx, c, &event); err != nil {
            if websocket.CloseStatus(err) != -1{
                s.pool.Remove(user.Username)
                return
            }
            wsInvalidJSONPayload(ctx, c)
            continue
		}
		event.From = user.Username
		event.FromID = user.ID
		s.eventChan <- &event
	}
}

func (s *Server) accept(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		s.badRequestResponse(w, r, err)
		return
	}
	s.handleWebSocketConn(r.Context(), conn)
}

func (s *Server) eventLoop() {
	// worker pool, to avoid bottleneck
	for i := 0; i < 5; i++ {
		go func() {
			for event := range s.eventChan {
                errContext, cancel := context.WithTimeout(context.Background(), time.Second * 5)
                defer cancel()
				switch event.Type {
				case messaging.TEXT:
					s.handleText(event)
				case messaging.SearchUserRequest:
					var r messaging.SearchUserPayload
					if err := json.Unmarshal([]byte(event.Body), &r); err != nil {
						client := s.pool.Get(event.From)
						if client != nil {
                            wsInvalidJSONPayload(errContext, client.Conn)
                            cancel()
						}
						continue
					}
					r.From = event.From
					s.handleUserSearch(&r)
				case messaging.LoadChatHistoryRequest:
					var p messaging.LoadChatHistoryReqPayload
					if err := json.Unmarshal([]byte(event.Body), &p); err != nil {
						client := s.pool.Get(event.From)
						if client != nil {
                            wsInvalidJSONPayload(errContext, client.Conn)
                            cancel()
						}
                        continue
					}
					s.handleLoadChatHistory(&p, event.From)

                case messaging.MARK_READ:
                    s.handleMarkRead(event)
				}
                cancel()
			}
		}()
	}
}

func (s *Server) ensureChatExists(fromID, toID int) (*store.Chat, error) {
	chat, err := s.store.Chats.GetByUsersID(fromID, toID)
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {

			chat := &store.Chat{UserID: fromID, With: toID}
			if err := s.store.Chats.Create(chat); err != nil {
				return nil, err
			}
			return chat, nil
		}

		return nil, err
	}

	return chat, nil
}

func (s *Server) storeMessage(msg *messaging.IncomingMessagePayload, chatID int) (int, error) {
	message := &store.Message{Content: msg.Content, ChatID: chatID, UserID: msg.FromUserID}
	err := s.store.Messages.Create(message)
	if err != nil {
		return 0, err
	}
	return message.ID, nil
}

func (s *Server) sendMessage(ctx context.Context, msgID int, t *messaging.IncomingMessagePayload) {
	out := &messaging.OutGoingMessage{
        MsgID: msgID,
        To: t.To,
        From:      t.From,
        Content:   t.Content,
        Timestamp: time.Now().Unix(),
    }

	s.sendServerMessage(ctx, t.To, &messaging.ServerMessage{Type: messaging.TEXT, Body: out})
	s.sendServerMessage(
		ctx, t.From,
		&messaging.ServerMessage{
			Type: messaging.DELIVERED,
			Body: messaging.Delivered{
				MessageID: msgID,
				Mark:      t.MessageMark,
				TimeStamp: time.Now().Unix(),
			},
		},
	)
}


func (s *Server) sendServerMessage(ctx context.Context, to string, msg *messaging.ServerMessage) {
	client := s.pool.Get(to)

	if client != nil {
		wsjson.Write(ctx, client.Conn, msg)
        return
	}

    // look up the redis, if the client doesn't exist in 
    // the current server
    peerAddr, err := s.pool.GetServerAddr(to)

    if err != nil{
        // connection doesn't exist
        if err == redis.Nil{
            return
        } 
        log.Println("DEBUG: ", err)
        return
    }
    s.pubsub.Publish(peerAddr, &messaging.PeerMessage{To: to, Msg:msg})
}
