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


type errorEnv struct{
    Error string `json:"error"`
}


type Client struct{
    Conn *websocket.Conn
}


func (s *Server) handleHandShake(conn *websocket.Conn) {
    type envelope struct{
        Token string `json:"token"`
    }

    var tok envelope
    if err := wsjson.Read(context.TODO(), conn, &tok); err != nil{
        wsjson.Write(context.TODO(), conn, ServerMessage{Type: ERR, Body: errorEnv{Error: err.Error()} })
        return
    }

    jwtToken, err := s.auth.ValidateToken(tok.Token)
    if err != nil {
        wsjson.Write(context.TODO(), conn, ServerMessage{Type: ERR, Body: errorEnv{Error: errors.New("invalid token").Error()}})
        conn.Close(websocket.StatusAbnormalClosure, "")
        return
    }

    claims, _ := jwtToken.Claims.(jwt.MapClaims)

    userID, err := strconv.Atoi(fmt.Sprintf("%.f", claims["sub"]))
    if err != nil {
        wsjson.Write(context.TODO(), conn, ServerMessage{Type: ERR, Body: errorEnv{Error: errors.New("invalid user id").Error()}})
        conn.Close(websocket.StatusAbnormalClosure, "")
        return
    }


    user, err := s.store.Users.GetByID(userID)  

    if err != nil{
        log.Println("DEBUG",err.Error())
        wsjson.Write(context.TODO(), conn, ServerMessage{Type: ERR, Body: errorEnv{Error: errors.New("user doesn't exist").Error()}} )
        conn.Close(websocket.StatusAbnormalClosure, "")
        return
    }

    if err := wsjson.Write(context.TODO(), conn, ServerMessage{Type: PROFILE_INFO, Body: user}); err != nil{
        log.Println("DEBUG: couldn't write to conn", err)
        conn.Close(websocket.StatusAbnormalClosure, "")
    }

    log.Printf("%s has just gone online", user.Username)


    chatPreviews, err := s.store.Chats.ChatPreviews(user.ID)

    if err != nil{
        log.Println(len(chatPreviews))
        log.Println("DEBUG: ", err)
        wsjson.Write(context.TODO(), conn, ServerMessage{Type: ERR, Body: errorEnv{Error: "couldnt load past chats"}})
    }

    // send chats
    wsjson.Write(context.TODO(), conn, ServerMessage{Type: CHATPREVIEWS, Body: chatPreviews})

    s.wsConns.Store(user.Username, &Client{Conn: conn})
    go s.readLoop(conn, user)
}

func (s *Server) readLoop(c *websocket.Conn, user *store.User){
    for{
        var event Event
        if err := wsjson.Read(context.TODO(), c, &event); err != nil{
            log.Println("DEBUG: err", err)
            s.wsConns.Delete(user.Username)
            return
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
    if err != nil{
        s.badRequestResponse(w,r,err)
        return
    }
    s.handleHandShake(conn)
}

func (s *Server) eventLoop() {
    // worker pool, to avoid bottleneck
    for i := 0; i < 5; i++{
        go func(){
            for event := range s.eventChan {
                switch event.Type {
                case TEXT:
                    var t   IncomingMessagePayload 
                    if err := json.Unmarshal([]byte(event.Body), &t); err != nil{
                        client := s.getClient(event.From)
                        if client != nil{
                            wsjson.Write(context.TODO(), client.Conn, errorEnv{Error: "coulnd't parse message"})
                            continue
                        }
                    }

                    t.From = event.From
                    t.FromUserID = event.FromID

                    s.handleText(&t)
                case SearchUserRequest:
                    var r SearchUserPayload
                    if err := json.Unmarshal([]byte(event.Body), &r); err != nil{
                        log.Println("DEBUG: ", err)
                        continue
                    }
                    r.From = event.From
                    s.handleUserSearch(&r)
                case LoadChatHistoryRequest:
                    var p LoadChatHistoryReqPayload
                    if err := json.Unmarshal([]byte(event.Body), &p); err != nil{
                        log.Println("DEBUG: ", err)
                        continue
                    }
                    s.handleLoadChatHistory(&p, event.From)
                }
            }
        }()
    }
}

func (s *Server) getClient(username string) *Client {
    v, ok := s.wsConns.Load(username)
    if !ok{
        return nil
    }
    return v.(*Client)
}

func (s *Server) ensureChatExists(fromID, toID int) (*store.Chat, error) {
    chat, err := s.store.Chats.GetByUsersID(fromID, toID)
    if err != nil{

        if errors.Is(err, sql.ErrNoRows){

            chat := &store.Chat{UserID: fromID, With: toID}
            if err := s.store.Chats.Create(chat); err != nil{
                return nil, err
            }
            return chat, nil
        }

        return  nil, err
    }

    return chat, nil
}

func (s *Server) storeMessage(msg *IncomingMessagePayload, chatID int) error{
    message := &store.Message{Content: msg.Content, ChatID: chatID, UserID: msg.FromUserID}
    err := s.store.Messages.Create(message)
    return err
}

func (s *Server) sendMessage(ctx context.Context, t *IncomingMessagePayload) {
    from, to := s.getClient(t.From), s.getClient(t.To)

    out := &OutGoingMessage{
        From:      t.From,
        Content:   t.Content,
        Timestamp: time.Now().Unix(),
    }

    if to != nil {
        if err := wsjson.Write(ctx, to.Conn, &ServerMessage{Type: TEXT, Body: out}); err != nil && from != nil {
            from.Conn.Write(ctx, websocket.MessageText, []byte("{\"type\":3, \"error\":\"couldn't write json\"}"))
        }
    }

    if from != nil {
        wsjson.Write(ctx, from.Conn, ServerMessage{Type: DELIVERED, Body: &Delivered{TimeStamp: out.Timestamp, Mark: t.MessageMark}})
    }
}

func (s *Server) handleText(t *IncomingMessagePayload) {

    chat, err := s.ensureChatExists(t.FromUserID, t.ToUserID)
    if err != nil{
        log.Println("DEBUG: couldn't ensure chat exists", err)
        return
    }

    err = s.storeMessage(t, chat.ID)
    if err != nil{
        log.Println("DEBUG: couldn't store the message: ", err)
        return
    }
    
    ctx := context.TODO()
    s.sendMessage(ctx, t)
}

func (s *Server) handleUserSearch(query *SearchUserPayload) {
    users, err := s.store.Users.SearchByUsername(query.Username)
    if err != nil{
        log.Println("DEBUG: ", err)
        return
    }
    client := s.getClient(query.From)
    if client != nil{
        wsjson.Write(context.TODO(), client.Conn, ServerMessage{Type: SearchUserResponse, Body: users})
    }
}

func (s *Server) handleLoadChatHistory(p *LoadChatHistoryReqPayload, from string) {
    chat, err := s.store.Chats.GetByUsersID(p.User1ID, p.User2ID)
    if err != nil{
        log.Println("DEBUG: ", err)
        return
    }
    messages, err := s.store.Messages.GetByChatID(chat.ID)
    if err != nil{
        log.Println("DEBUG: ", err)
        return
    }

    client := s.getClient(from)

    if client != nil{
        wsjson.Write(context.TODO(), client.Conn, ServerMessage{Type: LoadChatHistoryResponse, Body: messages})
    }
}


func (s *Server) markRead(m *MarkRead) {
    
}
