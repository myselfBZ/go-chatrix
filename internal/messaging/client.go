package messaging

import (
	"github.com/coder/websocket"
	"github.com/myselfBZ/chatrix/internal/store"
)


type Client struct{
    Conn *websocket.Conn
    User *store.User
    closeChan chan struct{}
}

func NewClient(u *store.User, conn *websocket.Conn) *Client{
    return &Client{
        Conn: conn,
        User: u,
        closeChan: make(chan struct{}),
    }
}
