package store

import (
	"database/sql"
	"time"
)

type MessageStore struct{
    db *sql.DB
}



type Message struct{
    ID int      `json:"id"`
    Content string `json:"content"`
    ChatID int `json:"chat_id"`
    CreatedAt time.Time `json:"created_at"`
    UserID int      `json:"user_id"`
}


func (m *MessageStore) Create(msg *Message) error {
    q := `INSERT INTO messages 
    (user_id, chat_id, content) 
    VALUES(
        $1, 
        $2,
        $3
    ) RETURNING id, created_at` 
    err := m.db.QueryRow(q, msg.UserID, msg.ChatID, msg.Content).Scan(&msg.ID, &msg.CreatedAt)
    return err
}


func (m *MessageStore) GetByChatID(id int) ([]*Message, error) {
    q := `SELECT user_id, chat_id, content, created_at FROM messages WHERE chat_id = $1`
    r,  err := m.db.Query(q, id)
    if err != nil{
        return nil, err
    }
    var messages []*Message
    for r.Next(){
        var msg Message
        err := r.Scan(&msg.UserID, &msg.ChatID, &msg.Content, &msg.CreatedAt)
        if err != nil{
            return nil, err
        }
        messages = append(messages, &msg)
    }
    return messages, nil
}


