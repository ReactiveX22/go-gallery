package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	DefaulResetDuration = 1 * time.Hour
)

type PasswordReset struct {
	ID        int
	UserID    int
	Token     string
	TokenHash string
	ExpiresAt time.Time
}

type PasswordResetService struct {
	DB            *sql.DB
	BytesPerToken int
	Duration      time.Duration
}

func (s *PasswordResetService) Create(email string) (*PasswordReset, error) {
	email = strings.ToLower(email)
	var userID int
	row := s.DB.QueryRow(`SELECT id FROM users WHERE email=$1`, email)
	err := row.Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	newToken, err := newToken(s.BytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	duration := s.Duration
	if duration == 0 {
		duration = DefaulResetDuration
	}
	pwReset := PasswordReset{
		UserID:    userID,
		Token:     newToken.Token,
		TokenHash: newToken.TokenHash,
		ExpiresAt: time.Now().Add(duration),
	}

	row = s.DB.QueryRow(`INSERT INTO password_resets (user_id, token_hash, expires_at) VALUES ($1, $2, $3) 
	ON CONFLICT (user_id) DO UPDATE SET token_hash=$2, expires_at=$3 RETURNING id`, pwReset.UserID, pwReset.TokenHash, pwReset.ExpiresAt)
	err = row.Scan(&pwReset.ID)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	return &pwReset, nil
}

func (s *PasswordResetService) Consume(token string) (*User, error) {

	tokenHash := hash(token)
	var user User
	var pwReset PasswordReset
	row := s.DB.QueryRow(`
	SELECT password_resets.id, password_resets.expires_at, 
	users.id, users.email, users.password_hash
	FROM password_resets 
	JOIN users ON users.id = password_resets.user_id
	WHERE password_resets.token_hash = $1
	`, tokenHash)

	err := row.Scan(&pwReset.ID, &pwReset.ExpiresAt, &user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("comsume: %w", err)
	}

	if time.Now().After(pwReset.ExpiresAt) {
		return nil, fmt.Errorf("token expired: %v", err)
	}

	err = s.delete(pwReset.ID)
	if err != nil {
		return nil, fmt.Errorf("comsume: %w", err)
	}

	return &user, nil
}

func (s *PasswordResetService) delete(id int) error {
	_, err := s.DB.Exec(`DELETE FROM password_resets WHERE id=$1;`, id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}
