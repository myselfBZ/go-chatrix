package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/golang-jwt/jwt/v5"
	"github.com/myselfBZ/chatrix/internal/store"
    "github.com/myselfBZ/chatrix/internal/events"
)


type Client struct {
	Conn *websocket.Conn
}

func (s *Server) handleHandShake(ctx context.Context, conn *websocket.Conn) {

	type envelope struct {
		Token string `json:"token"`
	}

	var tok envelope
	if err := wsjson.Read(ctx, conn, &tok); err != nil {
        wsInvalidJSONPayload(ctx, conn)
        conn.Close(websocket.CloseStatus(err), "invalid json payload")
		return
	}

	jwtToken, err := s.auth.ValidateToken(tok.Token)

	if err != nil {
		wsjson.Write(ctx, conn, events.ServerMessage{Type: events.ERR, Body: ErrEnvelope{Error: errors.New("invalid token")}})
		conn.Close(websocket.StatusAbnormalClosure, "")
		return
	}

	claims, _ := jwtToken.Claims.(jwt.MapClaims)

	userID, err := strconv.Atoi(fmt.Sprintf("%.f", claims["sub"]))

	if err != nil {
		conn.Close(websocket.StatusPolicyViolation, "invalid user id")
		return
	}

	user, err := s.store.Users.GetByID(userID)

	if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            wsjson.Write(ctx, conn, events.ServerMessage{Type: events.ERR, Body: ErrEnvelope{Error: errors.New("user doesn't exist")}})
            conn.Close(websocket.StatusPolicyViolation, "")
            return
        } 

        log.Println("DEBUG", err.Error())
		conn.Close(websocket.StatusInternalError, "")
		return
	}

	if err := wsjson.Write(ctx, conn, events.ServerMessage{Type: events.PROFILE_INFO, Body: user}); err != nil {
        if websocket.CloseStatus(err) == -1{
            conn.Close(websocket.StatusInternalError, "couldn't write json")
            return
        }
        // client diconnected
        return
	}

	log.Printf("%s has just gone online", user.Username)

	chatPreviews, err := s.store.Chats.ChatPreviews(user.ID)

    if err == nil{
        if err := wsjson.Write(ctx, conn, events.ServerMessage{Type: events.CHATPREVIEWS, Body: chatPreviews}); err != nil {
            if websocket.CloseStatus(err) != -1{
                conn.Close(websocket.StatusInternalError, "client disconnected")
                return
            }
            // client diconnected
            return
        }
    }

	if err != nil {
        switch err{
            case sql.ErrNoRows:
                wsjson.Write(ctx, conn, events.ServerMessage{Type: events.CHATPREVIEWS, Body: nil})
            default:
                wsjson.Write(ctx, conn, events.ServerMessage{Type: events.ERR, Body: InternalServerError })
        }
	}


	s.wsConns.Store(user.Username, &Client{Conn: conn})
	go s.readLoop(conn, user)
}

func (s *Server) readLoop(c *websocket.Conn, user *store.User) {
    ctx := context.Background()
	for {
		var event events.Event
		if err := wsjson.Read(ctx, c, &event); err != nil {
            if IsCloseErr(ctx, err){
                s.wsConns.Delete(user.Username)
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
					var t events.IncomingMessagePayload
					if err := json.Unmarshal([]byte(event.Body), &t); err != nil {
						client := s.getClient(event.From)
						if client != nil {
                            wsInvalidJSONPayload(errContext, client.Conn)
                            cancel()
							continue
						}
					}

					t.From = event.From
					t.FromUserID = event.FromID

					s.handleText(&t)
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

func (s *Server) handleText(t *events.IncomingMessagePayload) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

	chat, err := s.ensureChatExists(t.FromUserID, t.ToUserID)
	if err != nil {
        s.sendServerMessage(ctx, t.From, &events.ServerMessage{Type: events.ERR, Body: InternalServerError})
		return
	}

	id, err := s.storeMessage(t, chat.ID)

	if err != nil {
		log.Println("DEBUG: couldn't store the message: ", err)
        s.sendServerMessage(ctx, t.From, &events.ServerMessage{Type: events.ERR, Body: InternalServerError})
		return
	}

	s.sendMessage(ctx, id, t)
}

func (s *Server) handleUserSearch(query *events.SearchUserPayload) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
	users, err := s.store.Users.SearchByUsername(query.Username)
	if err != nil {
        if errors.Is(err, sql.ErrNoRows){
            s.sendServerMessage(ctx, query.From, &events.ServerMessage{Type: events.ERR, Body: ErrEnvelope{ Error: errors.New("user not found") }})
            return
        }
		log.Println("DEBUG: ", err)
		return
	}
	client := s.getClient(query.From)
	if client != nil {
        wsjson.Write(ctx, client.Conn, events.ServerMessage{Type: events.SearchUserResponse, Body: users})
	}
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
	}
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
    s.sendServerMessage(ctx, m.To, &events.ServerMessage{Type: events.MARK_READ, Body: m.MessageIds})
}
