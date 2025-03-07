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
)


type Client struct {
	Conn *websocket.Conn
}

func (s *Server) handleHandShake(conn *websocket.Conn) {

    ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
    defer cancel()

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
		wsjson.Write(ctx, conn, ServerMessage{Type: ERR, Body: ErrEnvelope{Error: errors.New("invalid token")}})
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
            wsjson.Write(ctx, conn, ServerMessage{Type: ERR, Body: ErrEnvelope{Error: errors.New("user doesn't exist")}})
            conn.Close(websocket.StatusPolicyViolation, "")
            return
        } 

        log.Println("DEBUG", err.Error())
		conn.Close(websocket.StatusInternalError, "")
		return
	}

	if err := wsjson.Write(ctx, conn, ServerMessage{Type: PROFILE_INFO, Body: user}); err != nil {
        if websocket.CloseStatus(err) == -1{
            conn.Close(websocket.StatusInternalError, "couldn't write json")
            return
        }
        // client diconnected
        return
	}

	log.Printf("%s has just gone online", user.Username)

	chatPreviews, err := s.store.Chats.ChatPreviews(user.ID)

	if err != nil {
        switch err{
            case sql.ErrNoRows:
                break
            default:
                conn.Close(websocket.StatusInternalError, "")
                return
        }
	}

	if err := wsjson.Write(ctx, conn, ServerMessage{Type: CHATPREVIEWS, Body: chatPreviews}); err != nil {
        if websocket.CloseStatus(err) == -1{
            conn.Close(websocket.StatusInternalError, "couldn't write json")
            return
        }
        // client diconnected
        return
	}

	s.wsConns.Store(user.Username, &Client{Conn: conn})
	go s.readLoop(conn, user)
}

func (s *Server) readLoop(c *websocket.Conn, user *store.User) {
    ctx, cancel := context.WithCancel(context.Background())
	defer cancel() 
	for {
		var event Event
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
		s.eventChan <- event
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
	s.handleHandShake(conn)
}

func (s *Server) eventLoop() {
	// worker pool, to avoid bottleneck
	for i := 0; i < 5; i++ {
		go func() {
			for event := range s.eventChan {
                errContext, cancel := context.WithCancel(context.Background())
                defer cancel()
				switch event.Type {
				case TEXT:
					var t IncomingMessagePayload
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
				case SearchUserRequest:
					var r SearchUserPayload
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
				case LoadChatHistoryRequest:
					var p LoadChatHistoryReqPayload
					if err := json.Unmarshal([]byte(event.Body), &p); err != nil {
						client := s.getClient(event.From)
						if client != nil {
                            wsInvalidJSONPayload(errContext, client.Conn)
                            cancel()
						}
                        continue
					}
					s.handleLoadChatHistory(&p, event.From)
                case MARK_READ:
					var p MarkReadRequestPayload 
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

func (s *Server) storeMessage(msg *IncomingMessagePayload, chatID int) (int, error) {
	message := &store.Message{Content: msg.Content, ChatID: chatID, UserID: msg.FromUserID}
	err := s.store.Messages.Create(message)
	if err != nil {
		return 0, err
	}
	return message.ID, nil
}

func (s *Server) sendMessage(ctx context.Context, msgID int, t *IncomingMessagePayload) {
	out := &OutGoingMessage{
        MsgID: msgID,
		From:      t.From,
		Content:   t.Content,
		Timestamp: time.Now().Unix(),
	}

	s.sendServerMessage(ctx, t.To, &ServerMessage{Type: TEXT, Body: out})
	s.sendServerMessage(
		ctx, t.From,
		&ServerMessage{
			Type: DELIVERED,
			Body: Delivered{
				MessageID: msgID,
				Mark:      t.MessageMark,
				TimeStamp: time.Now().Unix(),
			},
		},
	)
}

func (s *Server) handleText(t *IncomingMessagePayload) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

	chat, err := s.ensureChatExists(t.FromUserID, t.ToUserID)
	if err != nil {
        s.sendServerMessage(ctx, t.From, &ServerMessage{ERR, InternalServerError})
		return
	}

	id, err := s.storeMessage(t, chat.ID)

	if err != nil {
		log.Println("DEBUG: couldn't store the message: ", err)
        s.sendServerMessage(ctx, t.From, &ServerMessage{ERR, InternalServerError})
		return
	}

	s.sendMessage(ctx, id, t)
}

func (s *Server) handleUserSearch(query *SearchUserPayload) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
	users, err := s.store.Users.SearchByUsername(query.Username)
	if err != nil {
        if errors.Is(err, sql.ErrNoRows){
            s.sendServerMessage(ctx, query.From, &ServerMessage{ERR, errors.New("user not found")})
            return
        }
		log.Println("DEBUG: ", err)
		return
	}
	client := s.getClient(query.From)
	if client != nil {
		wsjson.Write(ctx, client.Conn, ServerMessage{Type: SearchUserResponse, Body: users})
	}
}

func (s *Server) handleLoadChatHistory(p *LoadChatHistoryReqPayload, from string) {
	chat, err := s.store.Chats.GetByUsersID(p.User1ID, p.User2ID)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			client := s.getClient(from)
			// if chat doesn't exist send nothing
			if client != nil {
				wsjson.Write(context.TODO(), client.Conn, ServerMessage{Type: LoadChatHistoryResponse, Body: nil})
				return
			}

			log.Println("DEBUG: ", err)
			return
		}
		return
	}
	messages, err := s.store.Messages.GetByChatID(chat.ID)
	if err != nil {
		log.Println("DEBUG: ", err)
		return
	}

	client := s.getClient(from)

	if client != nil {
		wsjson.Write(context.TODO(), client.Conn, ServerMessage{Type: LoadChatHistoryResponse, Body: messages})
	}
}

func (s *Server) sendServerMessage(ctx context.Context, to string, msg *ServerMessage) {
	client := s.getClient(to)

	if client != nil {
		wsjson.Write(ctx, client.Conn, msg)
	}
}

func (s *Server) handleMarkRead(m *MarkReadRequestPayload) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
	err := s.store.Messages.MarkRead(m.MessageIds)
	if err != nil {
		log.Println("DEBUG: ", err)
        s.sendServerMessage(ctx, m.From, &ServerMessage{ERR, InternalServerError})
		return
	}
    s.sendServerMessage(ctx, m.To, &ServerMessage{Type: MARK_READ, Body: m.MessageIds})
}
