package store

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)



func NewMockStore() *Store {
    return &Store{
        Users: &MockUserStore{},
    }
}

var pass password

func init(){
    text := "supersecretpassword"
    hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
    if err != nil{
        log.Fatalf("couldn't hash a password")
    }

    pass = password{
        text: &text,
        hash: hash,
    }
}


type MockUserStore struct{}

func (s *MockUserStore) GetByID(id int) (*User, error) {
    return &User{ID: id}, nil
}

func (s *MockUserStore) GetByUsername(username string) (*User, error) {
    return &User{
        Username: username,
        Password: pass,
    }, nil
}

func (s *MockUserStore) SearchByUsername(username string) ([]*User, error) {
    return []*User{
        {Username: username},
    }, nil
}

		
func (s *MockUserStore) Create(u *User) error{
    u.ID = 1
    return nil
}
