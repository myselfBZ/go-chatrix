package store

import (
	"database/sql"
	"time"
)

type ChatStore struct {
	db *sql.DB
}

// with stands for "conversation with" so the other person in the conversation
type Chat struct {
	ID        int
	UserID    int
	With      int
	CreatedAt time.Time
}

type ChatPreview struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
    Name     string `json:"name"`
}

func (c *ChatStore) Create(chat *Chat) error {
	q := `INSERT INTO chats(user_1_id, user_2_id) VALUES($1, $2) RETURNING id`
	err := c.db.QueryRow(q, chat.UserID, chat.With).Scan(&chat.ID)
	return err
}

func (c *ChatStore) ChatPreviews(userID int) ([]*ChatPreview, error) {
	q := `SELECT 
            users.id,
            users.name,
            users.username AS with_username
          FROM chats
          JOIN users ON users.id = (CASE WHEN chats.user_1_id = $1 THEN chats.user_2_id ELSE chats.user_1_id END)
          WHERE chats.user_1_id = $1 OR chats.user_2_id = $1;`
	r, err := c.db.Query(q, userID)
	if err != nil {
		return nil, err
	}
	var chats []*ChatPreview
	for r.Next() {
		var chat ChatPreview
		err := r.Scan(&chat.ID, &chat.Name, &chat.Username)
		if err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}
	return chats, nil
}

func (c *ChatStore) GetByUsersID(user1, user2 int) (*Chat, error) {
	q := `SELECT id, user_1_id, user_2_id FROM chats WHERE (user_1_id = $1 AND user_2_id = $2) OR (user_1_id = $2 AND user_2_id = $1)`
	var chat Chat
	err := c.db.QueryRow(q, user1, user2).Scan(&chat.ID, &chat.UserID, &chat.With)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (c *ChatStore) HasChatWith(user, other int) error {
	q := `SELECT 1 
    FROM chats 
    WHERE (user_1_id = $1 AND user_2_id = $2) 
    OR (user_1_id = $2 AND user_2_id = $1) 
    LIMIT 1;
    `
	var exists bool
	err := c.db.QueryRow(q, user, other).Scan(&exists)
	return err
}
