package store

import "database/sql"

type MessageStore struct{
    db *sql.DB
}

type Message struct{
    Content string
    ChatID int
    Timestamp int
    UserID int
}

// TODO batching

// func (m *MessageStore) Create(msg *Message) error {
//     q := `INSERT INTO messagse (user_id)`    
// }
