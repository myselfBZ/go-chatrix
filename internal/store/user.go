package store

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int      `json:"id"`
	Username string   `json:"username"`
	Password password `json:"-"`
	Name     string   `json:"name"`
}

type UserStore struct {
	db *sql.DB
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

func (p *password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))
}

func (s *UserStore) Create(u *User) error {
	q := `INSERT INTO users(username, password, name) VALUES($1, $2, $3) RETURNING id`
	err := s.db.QueryRow(q, u.Username, u.Password.hash, u.Name).Scan(&u.ID)
	return err
}

func (s *UserStore) GetByID(id int) (*User, error) {
	q := `SELECT id, username, password, name FROM users WHERE id = $1`
	var u User
	err := s.db.QueryRow(q, id).Scan(
		&u.ID,
		&u.Username,
		&u.Password.hash,
		&u.Name,
	)

	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) GetByUsername(username string) (*User, error) {
	q := `SELECT id, username, password, name FROM users WHERE username = $1`
	var u User
	err := s.db.QueryRow(q, username).Scan(
		&u.ID,
		&u.Username,
		&u.Password.hash,
		&u.Name,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (s *UserStore) SearchByUsername(username string) ([]*User, error) {
	username = "%" + username + "%"
	q := `SELECT name, username, id FROM users WHERE username ILIKE $1`
	var users []*User
	r, err := s.db.Query(q, username)
	if err != nil {
		return nil, err
	}

	for r.Next() {
		var user User
		err := r.Scan(&user.Name, &user.Username, &user.ID)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}
