package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"example/web-go/rand"
	"fmt"
)

const (
	MinBytesPerToken = 32
)

type Session struct {
	ID     int
	UserID int
	// Token is only set when creating a new session. Only store hash in db.
	Token     string
	TokenHash string
}

type SessionService struct {
	DB            *sql.DB
	BytesPerToken int
}

func (ss SessionService) Create(userID int) (*Session, error) {

	newToken, err := NewToken(ss.BytesPerToken)

	if err != nil {
		return nil, fmt.Errorf("create sesion: %w", err)
	}

	session := Session{
		UserID:    userID,
		Token:     newToken.Token,
		TokenHash: newToken.TokenHash,
	}

	row := ss.DB.QueryRow(`UPDATE sessions SET token_hash=$2 WHERE user_id=$1 RETURNING id`, session.UserID, session.TokenHash)
	err = row.Scan(&session.ID)
	if err == sql.ErrNoRows {
		row = ss.DB.QueryRow(`INSERT INTO sessions (user_id, token_hash) VALUES ($1, $2) RETURNING id`, session.UserID, session.TokenHash)
		err = row.Scan(&session.ID)
	}
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &session, nil
}

func (ss SessionService) User(token string) (*User, error) {
	var user User

	query := `
		SELECT u.id, u.email, u.password_hash
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.token_hash = $1
	`

	row := ss.DB.QueryRow(query, hash(token))
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash)

	if err != nil {
		return nil, fmt.Errorf("retrieving user with token: %w", err)
	}

	return &user, nil
}

func (ss SessionService) Delete(token string) error {
	tokenHash := hash(token)

	_, err := ss.DB.Exec(`DELETE FROM sessions WHERE token_hash=$1`, tokenHash)

	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}

type Token struct {
	Token     string
	TokenHash string
}

func NewToken(bytesPerToken int) (*Token, error) {
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("new: %w", err)
	}

	tokenHash := hash(token)
	return &Token{Token: token, TokenHash: tokenHash}, nil
}
