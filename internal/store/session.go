package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Session struct {
	ID         string
	RoleConfig string
	Goals      string
	Dimensions string
	Status     string
	RoundLimit int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Message struct {
	ID          int64
	SessionID   string
	Role        string
	Content     string
	SequenceNum int
	CreatedAt   time.Time
}

type SessionStore struct {
	db *sql.DB
}

func NewSessionStore(db *sql.DB) *SessionStore {
	return &SessionStore{db: db}
}

func (s *SessionStore) CreateSession(ctx context.Context, session *Session) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (id, role_config, goals, dimensions, status, round_limit, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		session.ID, session.RoleConfig, session.Goals, session.Dimensions,
		session.Status, session.RoundLimit, session.CreatedAt, session.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (s *SessionStore) GetSession(ctx context.Context, id string) (*Session, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, role_config, goals, dimensions, status, round_limit, created_at, updated_at
		 FROM sessions WHERE id = ?`, id,
	)
	var sess Session
	if err := row.Scan(&sess.ID, &sess.RoleConfig, &sess.Goals, &sess.Dimensions,
		&sess.Status, &sess.RoundLimit, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &sess, nil
}

func (s *SessionStore) UpdateSessionStatus(ctx context.Context, id string, status string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update session status: %w", err)
	}
	return nil
}

func (s *SessionStore) CreateMessage(ctx context.Context, msg *Message) error {
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO messages (session_id, role, content, sequence_num, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		msg.SessionID, msg.Role, msg.Content, msg.SequenceNum, msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}
	msg.ID, _ = result.LastInsertId()
	return nil
}

func (s *SessionStore) ListMessages(ctx context.Context, sessionID string) ([]Message, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, session_id, role, content, sequence_num, created_at
		 FROM messages WHERE session_id = ? ORDER BY sequence_num ASC`, sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.SequenceNum, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (s *SessionStore) CountMessages(ctx context.Context, sessionID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM messages WHERE session_id = ?`, sessionID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count messages: %w", err)
	}
	return count, nil
}
