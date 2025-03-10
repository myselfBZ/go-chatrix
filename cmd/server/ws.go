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
	"github.com/myselfBZ/chatrix/internal/events"
	"github.com/myselfBZ/chatrix/internal/store"
)


type Client struct {
	Conn *websocket.Conn
}


func (s *Server) sendInitialUserData(ctx context.Context, conn *websocket.Conn, user *store.User) error {
    ctxTimeout, cancel := context.WithTimeout(ctx, time.Second * 5)
    defer cancel()

	if err := wsjson.Write(ctxTimeout, conn, events.ServerMessage{Type: events.PROFILE_INFO, Body: user}); err != nil {
        if websocket.CloseStatus(err) == -1{
            conn.Close(websocket.StatusInternalError, "couldn't write json")
            return err
        }
        // client diconnected
        return err
	}

	chatPreviews, err := s.store.Chats.ChatPreviews(user.ID)

    if err == nil{
        if err := wsjson.Write(ctxTimeout, conn, events.ServerMessage{Type: events.CHATPREVIEWS, Body: chatPreviews}); err != nil {
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
            wsjson.Write(ctx, conn, events.ServerMessage{Type: events.CHATPREVIEWS, Body: nil})
        default:
            wsjson.Write(ctx, conn, events.ServerMessage{Type: events.ERR, Body: InternalServerError })
            log.Println("SERVER ERR: ", err)
        }
    }

    return nil
}

func (s *Server) handleHandShake(ctx context.Context, conn *websocket.Conn) {

    user, _, err := s.webSocketAuth(ctx, conn)

    if err != nil{
        return
    }


	log.Printf("%s has just gone online", user.Username)

    err = s.sendInitialUserData(ctx, conn, user)

    if err != nil{
        return 
    }

	s.wsConns.Store(user.Username, &Client{Conn: conn})

    s.registerUserKV(user.Username)
    go s.readLoop(conn, user)
}

func (s *Server) readLoop(c *websocket.Conn, user *store.User) {
    ctx := context.Background()
	for {
		var event events.Event
		if err := wsjson.Read(ctx, c, &event); err != nil {
            if websocket.CloseStatus(err) != -1{
                s.wsConns.Delete(user.Username)
                s.kv.Del(user.Username)
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
	s.handleHandShake(r.Context(), conn)
}

func (s *Server) eventLoop() {
	// worker pool, to avoid bottleneck
	for i := 0; i < 5; i++ {
		go func() {
			for event := range s.eventChan {
                errContext, cancel := context.WithTimeout(context.Background(), time.Second * 5)
                defer cancel()
				switch event.Type {
				case events.TEXT:
					s.handleText(event)
				case events.SearchUserRequest:
					var r events.SearchUserPayload
					if err := json.Unmarshal([]byte(event.Body), &r); err != nil {
						client := s.getClient(event.From)
						if client != nil {
                            wsInvalidJSONPayload(errContext, client.Conn)
                            cancel()
						}
						continue
					}
					r.From = event.From
					s.handleUserSearch(&r)
				case events.LoadChatHistoryRequest:
					var p events.LoadChatHistoryReqPayload
					if err := json.Unmarshal([]byte(event.Body), &p); err != nil {
						client := s.getClient(event.From)
						if client != nil {
                            wsInvalidJSONPayload(errContext, client.Conn)
                            cancel()
						}
                        continue
					}
					s.handleLoadChatHistory(&p, event.From)

                case events.MARK_READ:
                    if event.FromPeer{
                        var p MarkReadPayloadFromPeer
                        log.Println("really is this what's gonna happen", event.Body)
                        if err := json.Unmarshal([]byte(event.Body), &p); err != nil{
                            log.Println("DEBUG: ", err)
                            continue
                        }
                        s.sendServerMessage(context.TODO(), p.To, &events.ServerMessage{Type: events.MARK_READ, Body: p.MessageIds})
                        continue
                    }

					var p events.MarkReadRequestPayload 
					if err := json.Unmarshal([]byte(event.Body), &p); err != nil {
						client := s.getClient(event.From)
						if client != nil {
                            wsInvalidJSONPayload(errContext, client.Conn)
                            cancel()
						}
						continue
					}
                    p.From = event.From
                    s.handleMarkRead(&p)
				}
                cancel()
			}
		}()
	}
}

func (s *Server) getClient(username string) *Client {
	v, ok := s.wsConns.Load(username)
    // search the redis too
	if !ok {
		return nil
	}
	return v.(*Client)
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

func (s *Server) storeMessage(msg *events.IncomingMessagePayload, chatID int) (int, error) {
	message := &store.Message{Content: msg.Content, ChatID: chatID, UserID: msg.FromUserID}
	err := s.store.Messages.Create(message)
	if err != nil {
		return 0, err
	}
	return message.ID, nil
}

func (s *Server) sendMessage(ctx context.Context, msgID int, t *events.IncomingMessagePayload) {
	out := &events.OutGoingMessage{
        MsgID: msgID,
        To: t.To,
        From:      t.From,
        Content:   t.Content,
        Timestamp: time.Now().Unix(),
    }

	s.sendServerMessage(ctx, t.To, &events.ServerMessage{Type: events.TEXT, Body: out})
	s.sendServerMessage(
		ctx, t.From,
		&events.ServerMessage{
			Type: events.DELIVERED,
			Body: events.Delivered{
				MessageID: msgID,
				Mark:      t.MessageMark,
				TimeStamp: time.Now().Unix(),
			},
		},
	)
}

func (s *Server) handlePeerEvent(event *events.Event) {
    var outMsg events.OutGoingMessage

    if err := json.Unmarshal([]byte(event.Body), &outMsg); err != nil{
        if errors.Is(err, &json.SyntaxError{}){
            s.sendServerMessage(context.TODO(), event.From, &events.ServerMessage{Type: events.ERR, Body: ErrEnvelope{Error: err}})
            return
        }
        log.Println("DEBUG: ", err)
    }

    log.Println("Successfully decoded the message coming from the other peer")
    log.Println("it is actually going to", outMsg.To)

    s.sendServerMessage(context.TODO(), outMsg.To, &events.ServerMessage{Type: events.TEXT, Body: outMsg})
}

func (s *Server) handleText(event *events.Event) {
    if event.FromPeer{
        s.handlePeerEvent(event)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    var t events.IncomingMessagePayload

    if err := json.Unmarshal([]byte(event.Body), &t); err != nil {
        client := s.getClient(event.From)
        if client != nil {
            wsInvalidJSONPayload(context.TODO(), client.Conn)
            cancel()
        }
    }

    t.From = event.From
    t.FromUserID = event.FromID

    chat, err := s.ensureChatExists(t.FromUserID, t.ToUserID)

    if err != nil {
        s.sendServerMessage(ctx, t.From, &events.ServerMessage{Type: events.ERR, Body: InternalServerError})
        return
    }
    id, err := s.storeMessage(&t, chat.ID)

    if err != nil {
        log.Println("DEBUG: couldn't store the message: ", err)
        s.sendServerMessage(ctx, t.From, &events.ServerMessage{Type: events.ERR, Body: InternalServerError})
        return
    }
    s.sendMessage(ctx, id, &t)
}

func (s *Server) handleUserSearch(query *events.SearchUserPayload) {
    ctx := context.TODO()
	users, err := s.store.Users.SearchByUsername(query.Username)
	if err != nil {
        if errors.Is(err, sql.ErrNoRows){
            s.sendServerMessage(ctx, query.From, &events.ServerMessage{Type: events.ERR, Body: ErrEnvelope{ Error: errors.New("user not found") }})
            return
        }
		log.Println("DEBUG: ", err)
		return
	}
    s.sendServerMessage(ctx, query.From, &events.ServerMessage{Type: events.SearchUserResponse, Body: users})
}

func (s *Server) handleLoadChatHistory(p *events.LoadChatHistoryReqPayload, from string) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
	chat, err := s.store.Chats.GetByUsersID(p.User1ID, p.User2ID)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
            s.sendServerMessage(ctx, from, &events.ServerMessage{Type: events.LoadChatHistoryResponse, Body: nil})
			return
		}
        log.Println("SERVER ERROR:", err)
        s.sendServerMessage(ctx, from, &events.ServerMessage{Type: events.ERR, Body: InternalServerError})
		return
	}
	messages, err := s.store.Messages.GetByChatID(chat.ID)
	if err != nil {
        s.sendServerMessage(ctx, from, &events.ServerMessage{Type: events.ERR, Body: InternalServerError })
		return
	}
    s.sendServerMessage(ctx, from, &events.ServerMessage{Type: events.LoadChatHistoryResponse, Body: messages})
}

func (s *Server) sendServerMessage(ctx context.Context, to string, msg *events.ServerMessage) {
	client := s.getClient(to)

	if client != nil {
		wsjson.Write(ctx, client.Conn, msg)
        return
	}

    // look up the redis, if the client doesn't exist in 
    // the current server
    peerAddr, err := s.kv.Get(to)
    if err != nil{
        // connection doesn't exist
        if err == redis.Nil{
            return
        } 
        log.Println("DEBUG: ", err)
        return
    }

    body, err := json.Marshal(msg.Body)
    if err != nil{
        log.Println("ERROR: ", err)
    }
    s.pubSub.Pub.Publish(ctx, peerAddr, &events.Event{Type: msg.Type, Body:string(body) })
}

func (s *Server) handleMarkRead(m *events.MarkReadRequestPayload) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
	err := s.store.Messages.MarkRead(m.MessageIds)
	if err != nil {
		log.Println("DEBUG: ", err)
        s.sendServerMessage(ctx, m.From, &events.ServerMessage{Type: events.ERR, Body: InternalServerError})
		return
	}
    s.sendServerMessage(ctx, m.To, &events.ServerMessage{Type: events.MARK_READ, Body: map[string]any{"body":m.MessageIds, "to":m.To}})
}
