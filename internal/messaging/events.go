package messaging

import (
	"encoding/json"

	"github.com/myselfBZ/chatrix/internal/store"
)

type EventType int

const (
	TEXT EventType = iota
	DELIVERED
	MARK_READ
	ERR
	PROFILE_INFO
	CHATPREVIEWS

	SearchUserRequest
	SearchUserResponse

	LoadChatHistoryRequest
	LoadChatHistoryResponse
)

type ServerMessage struct {
	Type EventType `json:"type"`
	Body any       `json:"body"`

    ToPeer      bool `json:"-"`
}


type PeerMessage struct{
    To  string `json:"to"`
    Msg *ServerMessage `json:"msg"`
}

type Event struct {
	Type   EventType `json:"type"`

    // this feild lets the server know if this event came from connections 
    // in the current server
    FromPeer  bool      `json:"-"`
    // user's username and id
	From   string    `json:"-"`
	FromID int       `json:"-"`

    // the payload is unmarshaled saperately
	Body   json.RawMessage    `json:"body"`
}

type IncomingMessagePayload struct {
	To          string `json:"to"`
	Content     string `json:"content"`
	MessageMark int    `json:"mark"`
	From        string `json:"-"`

	ToUserID   int `json:"to_id"`
	FromUserID int `json:"-"`
}

type OutGoingMessage struct {
    MsgID     int   `json:"msg_id"`
	From      string `json:"from"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
    // set only for peer to peer communication
    To        string `json:"to"`
}

type Delivered struct {
	MessageID int   `json:"message_id"`
	Mark      int   `json:"mark"`
	TimeStamp int64 `json:"timestamp"`
}

type SearchUserPayload struct {
	Username string `json:"username"`
	From     string `json:"-"`
}


// Reading many messages in one go
type MarkReadRequestPayload struct {
	MessageIds []int `json:"message_ids"`

    // the one who needs to be informed about the read event
	To   string     `json:"to"`
    // the one who read the messages
	From string
}

type LoadChatHistoryReqPayload struct {
	User1ID int `json:"user1_id"`
	User2ID int `json:"user2_id"`
}

type ChatHistory struct {
	Messages []*store.Message `json:"messages"`
}

type MarkReadPayloadFromPeer struct{
    MessageIds []int   `json:"message_ids"` 
    To         string   `json:"to"`
}

