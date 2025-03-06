package store

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type MessageStore struct {
	db *sql.DB
}

type Message struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	ChatID    int       `json:"chat_id"`
	CreatedAt time.Time `json:"created_at"`
	UserID    int       `json:"user_id"`
	Read      bool      `json:"read"`
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
	q := `SELECT id, user_id, chat_id, content, created_at, read FROM messages WHERE chat_id = $1`
	r, err := m.db.Query(q, id)
	if err != nil {
		return nil, err
	}
	var messages []*Message
	for r.Next() {
		var msg Message
		err := r.Scan(&msg.ID, &msg.UserID, &msg.ChatID, &msg.Content, &msg.CreatedAt, &msg.Read)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}

func (c *MessageStore) MarkRead(ids []int) error {
	q := `UPDATE messages SET read = true WHERE id = ANY($1)`

	_, err := c.db.Exec(q, pq.Array(ids))
	return err
}
