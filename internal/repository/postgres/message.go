package postgres

import (
	"context"
	"database/sql"

	"github.com/VaneZ444/chat-service/internal/entity"
)

type MessageRepo interface {
	CreateMessage(ctx context.Context, msg *entity.Message) error
	GetLastMessages(ctx context.Context, limit int) ([]entity.Message, error)
}

type messageRepo struct {
	db *sql.DB
}

func NewMessageRepo(db *sql.DB) MessageRepo {
	return &messageRepo{db: db}
}

func (r *messageRepo) CreateMessage(ctx context.Context, msg *entity.Message) error {
	query := `INSERT INTO messages (author_id, content, created_at) VALUES ($1, $2, $3) RETURNING id`
	return r.db.QueryRowContext(ctx, query, msg.AuthorID, msg.Content, msg.CreatedAt).Scan(&msg.ID)
}

func (r *messageRepo) GetLastMessages(ctx context.Context, limit int) ([]entity.Message, error) {
	query := `SELECT id, author_id, content, created_at FROM messages ORDER BY created_at DESC LIMIT $1`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []entity.Message
	for rows.Next() {
		var msg entity.Message
		if err := rows.Scan(&msg.ID, &msg.AuthorID, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
