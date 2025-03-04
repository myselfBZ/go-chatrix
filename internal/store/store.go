package store

import "database/sql"

type Store struct{
    Users interface{
        GetByUsername(username string) (*User, error)
        Create(u *User) (error)
        GetByID(id int) (*User, error)
        SearchByUsername(username string) ([]*User, error) 
    }
    
    Chats interface{
        Create(*Chat) error
        ChatPreviews(int) ([]*ChatPreview, error)
        HasChatWith(user, other int) (error)
    }
}

func New(db *sql.DB) *Store {
    return &Store{
        Users: &UserStore{db},
        Chats: &ChatStore{db},
    }
}
